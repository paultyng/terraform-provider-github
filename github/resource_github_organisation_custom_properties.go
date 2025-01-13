package github

import (
	"context"

	"github.com/google/go-github/v57/github"
	"github.com/hashicorp/terraform-plugin-sdk/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceGithubOrganizationCustomProperties() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubCustomPropertiesCreate,
		Read:   resourceGithubCustomPropertiesRead,
		Update: resourceGithubCustomPropertiesUpdate,
		Delete: resourceGithubCustomPropertiesDelete,
		Importer: &schema.ResourceImporter{
			State: resourceGithubCustomPropertiesImport,
		},

		CustomizeDiff: customdiff.Sequence(
			customdiff.ComputedIf("slug", func(d *schema.ResourceDiff, meta interface{}) bool {
				return d.HasChange("name")
			}),
		),

		Schema: map[string]*schema.Schema{
			"property_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the custom property",
			},
			"value_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The type of the custom property",
			},
			"required": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether the custom property is required",
			},
			"default_value": {
				Type:        schema.TypeString,
				Description: "The default value of the custom property",
				Optional:    true,
				Computed:    true,
			},
			"description": {
				Description: "The description of the custom property",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"allowed_values": {
				Description: "The allowed values of the custom property",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceGithubCustomPropertiesCreate(d *schema.ResourceData, meta interface{}) error {
	ctx := context.Background()
	client := meta.(*Owner).v3client
	ownerName := meta.(*Owner).name

	propertyName := d.Get("property_name").(string)
	valueType := d.Get("value_type").(string)
	required := d.Get("required").(bool)
	defaultValue := d.Get("default_value").(string)
	description := d.Get("description").(string)
	allowedValues := d.Get("allowed_values").([]interface{})
	var allowedValuesString []string
	for _, v := range allowedValues {
		allowedValuesString = append(allowedValuesString, v.(string))
	}

	customProperty, _, err := client.Organizations.CreateOrUpdateCustomProperty(ctx, ownerName, d.Get("property_name").(string), &github.CustomProperty{
		PropertyName:  &propertyName,
		ValueType:     valueType,
		Required:      &required,
		DefaultValue:  &defaultValue,
		Description:   &description,
		AllowedValues: allowedValuesString,
	})
	if err != nil {
		return err
	}

	d.SetId(*customProperty.PropertyName)
	return resourceGithubCustomPropertiesRead(d, meta)
}

func resourceGithubCustomPropertiesRead(d *schema.ResourceData, meta interface{}) error {
	ctx := context.Background()
	client := meta.(*Owner).v3client
	ownerName := meta.(*Owner).name

	customProperty, _, err := client.Organizations.GetCustomProperty(ctx, ownerName, d.Get("property_name").(string))
	if err != nil {
		return err
	}

	d.SetId(*customProperty.PropertyName)
	d.Set("allowed_values", customProperty.AllowedValues)
	d.Set("default_value", customProperty.DefaultValue)
	d.Set("description", customProperty.Description)
	d.Set("property_name", customProperty.PropertyName)
	d.Set("required", customProperty.Required)
	d.Set("value_type", customProperty.ValueType)

	return nil
}

func resourceGithubCustomPropertiesUpdate(d *schema.ResourceData, meta interface{}) error {
	if err := resourceGithubCustomPropertiesCreate(d, meta); err != nil {
		return err
	}
	return resourceGithubTeamRead(d, meta)
}

func resourceGithubCustomPropertiesDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	ownerName := meta.(*Owner).name

	_, err := client.Organizations.RemoveCustomProperty(context.Background(), ownerName, d.Get("property_name").(string))
	if err != nil {
		return err
	}

	return nil
}

func resourceGithubCustomPropertiesImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// WIP
	return []*schema.ResourceData{d}, nil
}
