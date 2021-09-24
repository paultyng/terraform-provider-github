package github

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/shurcooL/githubv4"
	"log"
	"reflect"
	"strings"
)

func dataSourceGithubUsers() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGithubUsersRead,

		Schema: map[string]*schema.Schema{
			"usernames": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
			},
			"logins": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"node_ids": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"unknown_logins": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
		},
	}
}

func dataSourceGithubUsersRead(d *schema.ResourceData, meta interface{}) error {
	usernames := expandStringList(d.Get("usernames").([]interface{}))

	// Create GraphQL variables and query struct
	type (
		UserFragment struct {
			Id    string
			Login string
		}
	)
	var fields []reflect.StructField
	variables := make(map[string]interface{})
	for idx, username := range usernames {
		label := fmt.Sprintf("User%d", idx)
		variables[label] = githubv4.String(username)
		fields = append(fields, reflect.StructField{
			Name: label, Type: reflect.TypeOf(UserFragment{}), Tag: reflect.StructTag(fmt.Sprintf("graphql:\"%[1]s: user(login: $%[1]s)\"", label)),
		})
	}
	query := reflect.New(reflect.StructOf(fields)).Elem()

	if len(usernames) > 0 {
		log.Printf("[INFO] Refreshing GitHub Users: %s", strings.Join(usernames, ", "))
		ctx := context.WithValue(context.Background(), ctxId, d.Id())
		client := meta.(*Owner).v4client
		err := client.Query(ctx, query.Addr().Interface(), variables)
		if err != nil && !strings.Contains(err.Error(), "Could not resolve to a User with the login of") {
			return err
		}
	}

	var logins, nodeIDs, unknownLogins []string
	for idx, username := range usernames {
		label := fmt.Sprintf("User%d", idx)
		user := query.FieldByName(label).Interface().(UserFragment)
		if user.Login != "" {
			logins = append(logins, user.Login)
			nodeIDs = append(nodeIDs, user.Id)
		} else {
			unknownLogins = append(unknownLogins, username)
		}
	}

	d.SetId(buildChecksumID(usernames))
	d.Set("logins", logins)
	d.Set("node_ids", nodeIDs)
	d.Set("unknown_logins", unknownLogins)

	return nil
}
