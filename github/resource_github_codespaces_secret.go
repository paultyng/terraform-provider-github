package github

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGithubCodespacesSecret() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubCodespacesSecretCreateOrUpdate,
		Read:   resourceGithubCodespacesSecretRead,
		Delete: resourceGithubCodespacesSecretDelete,
		Importer: &schema.ResourceImporter{
			State: resourceGithubCodespacesSecretImport,
		},

		Schema: map[string]*schema.Schema{
			"repository": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the repository.",
			},
			"secret_name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				Description:      "Name of the secret.",
				ValidateDiagFunc: validateSecretNameFunc,
			},
			"encrypted_value": {
				Type:          schema.TypeString,
				ForceNew:      true,
				Optional:      true,
				Sensitive:     true,
				ConflictsWith: []string{"plaintext_value"},
				Description:   "Encrypted value of the secret using the GitHub public key in Base64 format.",
			},
			"plaintext_value": {
				Type:          schema.TypeString,
				ForceNew:      true,
				Optional:      true,
				Sensitive:     true,
				ConflictsWith: []string{"encrypted_value"},
				Description:   "Plaintext value of the secret to be encrypted.",
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

func resourceGithubCodespacesSecretCreateOrUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.Background()

	repo := d.Get("repository").(string)
	secretName := d.Get("secret_name").(string)
	plaintextValue := d.Get("plaintext_value").(string)
	var encryptedValue string

	keyId, publicKey, err := getCodespacesPublicKeyDetails(owner, repo, meta)
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
		Name:           secretName,
		KeyID:          keyId,
		EncryptedValue: encryptedValue,
	}

	_, err = client.Codespaces.CreateOrUpdateRepoSecret(ctx, owner, repo, eSecret)
	if err != nil {
		return err
	}

	d.SetId(buildTwoPartID(repo, secretName))
	return resourceGithubCodespacesSecretRead(d, meta)
}

func resourceGithubCodespacesSecretRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.Background()

	repoName, secretName, err := parseTwoPartID(d.Id(), "repository", "secret_name")
	if err != nil {
		return err
	}

	secret, _, err := client.Codespaces.GetRepoSecret(ctx, owner, repoName, secretName)
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

func resourceGithubCodespacesSecretDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	orgName := meta.(*Owner).name
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	repoName, secretName, err := parseTwoPartID(d.Id(), "repository", "secret_name")
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Deleting secret: %s", d.Id())
	_, err = client.Codespaces.DeleteRepoSecret(ctx, orgName, repoName, secretName)

	return err
}

func resourceGithubCodespacesSecretImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.Background()

	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid ID specified: supplied ID must be written as <repository>/<secret_name>")
	}

	d.SetId(buildTwoPartID(parts[0], parts[1]))

	repoName, secretName, err := parseTwoPartID(d.Id(), "repository", "secret_name")
	if err != nil {
		return nil, err
	}

	secret, _, err := client.Codespaces.GetRepoSecret(ctx, owner, repoName, secretName)
	if err != nil {
		return nil, err
	}

	if err = d.Set("repository", repoName); err != nil {
		return nil, err
	}
	if err = d.Set("secret_name", secretName); err != nil {
		return nil, err
	}

	// encrypted_value or plaintext_value can not be imported

	if err = d.Set("created_at", secret.CreatedAt.String()); err != nil {
		return nil, err
	}
	if err = d.Set("updated_at", secret.UpdatedAt.String()); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func getCodespacesPublicKeyDetails(owner, repository string, meta interface{}) (keyId, pkValue string, err error) {
	client := meta.(*Owner).v3client
	ctx := context.Background()

	publicKey, _, err := client.Codespaces.GetRepoPublicKey(ctx, owner, repository)
	if err != nil {
		return keyId, pkValue, err
	}

	return publicKey.GetKeyID(), publicKey.GetKey(), err
}
