package github

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/go-github/v66/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceGithubActionsEnvironmentSecret() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubActionsEnvironmentSecretCreateOrUpdate,
		Read:   resourceGithubActionsEnvironmentSecretRead,
		Delete: resourceGithubActionsEnvironmentSecretDelete,
		Importer: &schema.ResourceImporter{
			State: resourceGithubActionsEnvironmentSecretImport,
		},

		Schema: map[string]*schema.Schema{
			"repository": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the repository.",
			},
			"environment": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the environment.",
			},
			"secret_name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				Description:      "Name of the secret.",
				ValidateDiagFunc: validateSecretNameFunc,
			},
			"encrypted_value": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Sensitive:        true,
				Description:      "Encrypted value of the secret using the GitHub public key in Base64 format.",
				ConflictsWith:    []string{"plaintext_value"},
				ValidateDiagFunc: toDiagFunc(validation.StringIsBase64, "encrypted_value"),
			},
			"plaintext_value": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Sensitive:     true,
				Description:   "Plaintext value of the secret to be encrypted.",
				ConflictsWith: []string{"encrypted_value"},
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Date of 'actions_environment_secret' creation.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Date of 'actions_environment_secret' update.",
			},
		},
	}
}

func resourceGithubActionsEnvironmentSecretCreateOrUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.Background()

	repoName := d.Get("repository").(string)
	envName := d.Get("environment").(string)
	escapedEnvName := url.PathEscape(envName)
	secretName := d.Get("secret_name").(string)
	plaintextValue := d.Get("plaintext_value").(string)
	var encryptedValue string

	repo, _, err := client.Repositories.Get(ctx, owner, repoName)
	if err != nil {
		return err
	}

	keyId, publicKey, err := getEnvironmentPublicKeyDetails(repo.GetID(), escapedEnvName, meta)
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

	_, err = client.Actions.CreateOrUpdateEnvSecret(ctx, int(repo.GetID()), escapedEnvName, eSecret)
	if err != nil {
		return err
	}

	d.SetId(buildThreePartID(repoName, envName, secretName))
	return resourceGithubActionsEnvironmentSecretRead(d, meta)
}

func resourceGithubActionsEnvironmentSecretRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.Background()

	repoName, envName, secretName, err := parseThreePartID(d.Id(), "repository", "environment", "secret_name")
	if err != nil {
		return err
	}
	escapedEnvName := url.PathEscape(envName)

	repo, _, err := client.Repositories.Get(ctx, owner, repoName)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusNotFound {
				log.Printf("[INFO] Removing environment secret %s from state because it no longer exists in GitHub",
					d.Id())
				d.SetId("")
				return nil
			}
		}
		return err
	}

	secret, _, err := client.Actions.GetEnvSecret(ctx, int(repo.GetID()), escapedEnvName, secretName)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusNotFound {
				log.Printf("[INFO] Removing environment secret %s from state because it no longer exists in GitHub",
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
		log.Printf("[INFO] The environment secret %s has been externally updated in GitHub", d.Id())
		d.SetId("")
	} else if !ok {
		if err = d.Set("updated_at", secret.UpdatedAt.String()); err != nil {
			return err
		}
	}

	return nil
}

func resourceGithubActionsEnvironmentSecretDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	repoName, envName, secretName, err := parseThreePartID(d.Id(), "repository", "environment", "secret_name")
	if err != nil {
		return err
	}
	escapedEnvName := url.PathEscape(envName)
	repo, _, err := client.Repositories.Get(ctx, owner, repoName)
	if err != nil {
		return err
	}
	log.Printf("[INFO] Deleting environment secret: %s", d.Id())
	_, err = client.Actions.DeleteEnvSecret(ctx, int(repo.GetID()), escapedEnvName, secretName)

	return err
}

func resourceGithubActionsEnvironmentSecretImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.Background()

	parts := strings.Split(d.Id(), "/")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid ID specified: supplied ID must be written as <repository>/<environment>/<secret_name>")
	}

	d.SetId(buildThreePartID(parts[0], parts[1], parts[2]))

	repoName, envName, secretName, err := parseThreePartID(d.Id(), "repository", "environment", "secret_name")
	if err != nil {
		return nil, err
	}

	repo, _, err := client.Repositories.Get(ctx, owner, repoName)

	secret, _, err := client.Actions.GetEnvSecret(ctx, int(repo.GetID()), escapedEnvName, secretName)
	if err != nil {
		return nil, err
	}

	d.Set("repository", repoName)
	d.Set("secret_name", secretName)
	d.Set("environment", envName)

	// encrypted_value or plaintext_value can not be imported

	d.Set("created_at", secret.CreatedAt.String())
	d.Set("updated_at", secret.UpdatedAt.String())

	return []*schema.ResourceData{d}, nil
}

func getEnvironmentPublicKeyDetails(repoID int64, envName string, meta interface{}) (keyId, pkValue string, err error) {
	client := meta.(*Owner).v3client
	ctx := context.Background()

	publicKey, _, err := client.Actions.GetEnvPublicKey(ctx, int(repoID), envName)
	if err != nil {
		return keyId, pkValue, err
	}

	return publicKey.GetKeyID(), publicKey.GetKey(), err
}
