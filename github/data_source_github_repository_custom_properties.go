package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v66/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGithubRepositoryCustomProperties() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGithubOrgaRepositoryCustomProperties,

		Schema: map[string]*schema.Schema{
			"repository": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the repository which the custom properties should be on.",
			},
			"property": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "List of custom properties",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"property_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the custom property.",
						},
						"property_value": {
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "Value of the custom property.",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceGithubOrgaRepositoryCustomProperties(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*Owner).v3client
	ctx := context.Background()

	owner := meta.(*Owner).name

	repoName := d.Get("repository").(string)

	allCustomProperties, _, err := client.Repositories.GetAllCustomPropertyValues(ctx, owner, repoName)
	if err != nil {
		return err
	}

	results, err := flattenRepositoryCustomProperties(allCustomProperties)
	if err != nil {
		return err
	}

	d.SetId(buildTwoPartID(owner, repoName))
	d.Set("repository", repoName)
	d.Set("property", results)

	return nil
}

func flattenRepositoryCustomProperties(customProperties []*github.CustomPropertyValue) ([]interface{}, error) {

	results := make([]interface{}, 0)
	for _, prop := range customProperties {
		result := make(map[string]interface{})

		result["property_name"] = prop.PropertyName

		propertyValue, err := parseRepositoryCustomPropertyValueToStringSlice(prop)
		if err != nil {
			return nil, err
		}

		result["property_value"] = propertyValue

		results = append(results, result)
	}

	return results, nil
}

func parseRepositoryCustomPropertyValueToStringSlice(prop *github.CustomPropertyValue) ([]string, error) {
	switch value := prop.Value.(type) {
	case string:
		return []string{value}, nil
	case []string:
		return value, nil
	default:
		return nil, fmt.Errorf("custom property value couldn't be parsed as a string or a list of strings: %s", value)
	}
}
