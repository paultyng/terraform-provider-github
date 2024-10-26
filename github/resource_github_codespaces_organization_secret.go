package github

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceGithubCodespacesOrganizationSecret() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubCodespacesOrganizationSecretCreateOrUpdate,
		Read:   resourceGithubCodespacesOrganizationSecretRead,
		Update: resourceGithubCodespacesOrganizationSecretCreateOrUpdate,
		Delete: resourceGithubCodespacesOrganizationSecretDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				if err := d.Set("secret_name", d.Id()); err != nil {
					return nil, err
				}
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"secret_name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				Description:      "Name of the secret.",
				ValidateDiagFunc: validateSecretNameFunc,
			},
			"encrypted_value": {
				Type:             schema.TypeString,
				ForceNew:         true,
				Optional:         true,
				Sensitive:        true,
				ConflictsWith:    []string{"plaintext_value"},
				Description:      "Encrypted value of the secret using the GitHub public key in Base64 format.",
				ValidateDiagFunc: toDiagFunc(validation.StringIsBase64, "encrypted_value"),
			},
			"plaintext_value": {
				Type:          schema.TypeString,
				ForceNew:      true,
				Optional:      true,
				Sensitive:     true,
				Description:   "Plaintext value of the secret to be encrypted.",
				ConflictsWith: []string{"encrypted_value"},
			},
			"visibility": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "Configures the access that repositories have to the organization secret. Must be one of 'all', 'private' or 'selected'. 'selected_repository_ids' is required if set to 'selected'.",
				ValidateDiagFunc: validateValueFunc([]string{"all", "private", "selected"}),
				ForceNew:         true,
			},
			"selected_repository_ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Set:         schema.HashInt,
				Optional:    true,
				Description: "An array of repository ids that can access the organization secret.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Date of 'codespaces_secret' creation.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Date of 'codespaces_secret' update.",
			},
		},
	}
}

func resourceGithubCodespacesOrganizationSecretCreateOrUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.Background()

	secretName := d.Get("secret_name").(string)
	plaintextValue := d.Get("plaintext_value").(string)
	var encryptedValue string

	visibility := d.Get("visibility").(string)
	selectedRepositories, hasSelectedRepositories := d.GetOk("selected_repository_ids")

	if visibility != "selected" && hasSelectedRepositories {
		return fmt.Errorf("cannot use selected_repository_ids without visibility being set to selected")
	}

	selectedRepositoryIDs := github.SelectedRepoIDs{}

	if hasSelectedRepositories {
		ids := selectedRepositories.(*schema.Set).List()

		for _, id := range ids {
			selectedRepositoryIDs = append(selectedRepositoryIDs, int64(id.(int)))
		}
	}

	keyId, publicKey, err := getCodespacesOrganizationPublicKeyDetails(owner, meta)
	if err != nil {
		return err
	}

	if encryptedText, ok := d.GetOk("encrypted_value"); ok {
		encryptedValue = encryptedText.(string)
	} else {
		encryptedBytes, err := encryptPlaintext(plaintextValue, publicKey)
		if err != nil {
			return err
		}
		encryptedValue = base64.StdEncoding.EncodeToString(encryptedBytes)
	}

	// Create an EncryptedSecret and encrypt the plaintext value into it
	eSecret := &github.EncryptedSecret{
		Name:                  secretName,
		KeyID:                 keyId,
		Visibility:            visibility,
		SelectedRepositoryIDs: selectedRepositoryIDs,
		EncryptedValue:        encryptedValue,
	}

	_, err = client.Codespaces.CreateOrUpdateOrgSecret(ctx, owner, eSecret)
	if err != nil {
		return err
	}

	d.SetId(secretName)
	return resourceGithubCodespacesOrganizationSecretRead(d, meta)
}

func resourceGithubCodespacesOrganizationSecretRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.Background()

	secret, _, err := client.Codespaces.GetOrgSecret(ctx, owner, d.Id())
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusNotFound {
				log.Printf("[WARN] Removing actions secret %s from state because it no longer exists in GitHub",
					d.Id())
				d.SetId("")
				return nil
			}
		}
		return err
	}

	if err = d.Set("encrypted_value", d.Get("encrypted_value")); err != nil {
		return err
	}
	if err = d.Set("plaintext_value", d.Get("plaintext_value")); err != nil {
		return err
	}
	if err = d.Set("created_at", secret.CreatedAt.String()); err != nil {
		return err
	}
	if err = d.Set("visibility", secret.Visibility); err != nil {
		return err
	}

	selectedRepositoryIDs := []int64{}

	if secret.Visibility == "selected" {
		opt := &github.ListOptions{
			PerPage: 30,
		}
		for {
			results, resp, err := client.Codespaces.ListSelectedReposForOrgSecret(ctx, owner, d.Id(), opt)
			if err != nil {
				return err
			}

			for _, repo := range results.Repositories {
				selectedRepositoryIDs = append(selectedRepositoryIDs, repo.GetID())
			}

			if resp.NextPage == 0 {
				break
			}
			opt.Page = resp.NextPage
		}
	}

	if err = d.Set("selected_repository_ids", selectedRepositoryIDs); err != nil {
		return err
	}

	// This is a drift detection mechanism based on timestamps.
	//
	// If we do not currently store the "updated_at" field, it means we've only
	// just created the resource and the value is most likely what we want it to
	// be.
	//
	// If the resource is changed externally in the meantime then reading back
	// the last update timestamp will return a result different than the
	// timestamp we've persisted in the state. In that case, we can no longer
	// trust that the value (which we don't see) is equal to what we've declared
	// previously.
	//
	// The only solution to enforce consistency between is to mark the resource
	// as deleted (unset the ID) in order to fix potential drift by recreating
	// the resource.
	if updatedAt, ok := d.GetOk("updated_at"); ok && updatedAt != secret.UpdatedAt.String() {
		log.Printf("[WARN] The secret %s has been externally updated in GitHub", d.Id())
		d.SetId("")
	} else if !ok {
		if err = d.Set("updated_at", secret.UpdatedAt.String()); err != nil {
			return err
		}
	}

	return nil
}

func resourceGithubCodespacesOrganizationSecretDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	orgName := meta.(*Owner).name
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	log.Printf("[DEBUG] Deleting secret: %s", d.Id())
	_, err := client.Codespaces.DeleteOrgSecret(ctx, orgName, d.Id())
	return err
}

func getCodespacesOrganizationPublicKeyDetails(owner string, meta interface{}) (keyId, pkValue string, err error) {
	client := meta.(*Owner).v3client
	ctx := context.Background()

	publicKey, _, err := client.Codespaces.GetOrgPublicKey(ctx, owner)
	if err != nil {
		return keyId, pkValue, err
	}

	return publicKey.GetKeyID(), publicKey.GetKey(), err
}
