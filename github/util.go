package github

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	// https://developer.github.com/guides/traversing-with-pagination/#basics-of-pagination
	maxPerPage = 100
)

func checkOrganization(meta interface{}) error {
	if !meta.(*Owner).IsOrganization {
		return fmt.Errorf("this resource can only be used in the context of an organization, %q is a user", meta.(*Owner).name)
	}

	return nil
}

func caseInsensitive() schema.SchemaDiffSuppressFunc {
	return func(k, old, new string, d *schema.ResourceData) bool {
		return strings.EqualFold(old, new)
	}
}

// wrapErrors is provided to easily turn errors into diag.Diagnostics
// until we go through the provider and replace error usage
func wrapErrors(errs []error) diag.Diagnostics {
	var diags diag.Diagnostics

	for _, err := range errs {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Error",
			Detail:   err.Error(),
		})
	}

	return diags
}

// toDiagFunc is a helper that operates on Hashicorp's helper/validation functions
// and converts them to the diag.Diagnostic format
// --> nolint: oldFunc needs to be schema.SchemaValidateFunc to keep compatibility with
// the old code until all uses of schema.SchemaValidateFunc are gone
func toDiagFunc(oldFunc schema.SchemaValidateFunc, keyName string) schema.SchemaValidateDiagFunc { //nolint:staticcheck
	return func(i interface{}, path cty.Path) diag.Diagnostics {
		warnings, errors := oldFunc(i, keyName)
		var diags diag.Diagnostics

		for _, err := range errors {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  err.Error(),
			})
		}

		for _, warn := range warnings {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  warn,
			})
		}

		return diags
	}
}

func validateValueFunc(values []string) schema.SchemaValidateDiagFunc {
	return func(v interface{}, k cty.Path) diag.Diagnostics {
		errs := make([]error, 0)
		value := v.(string)
		valid := false
		for _, role := range values {
			if value == role {
				valid = true
				break
			}
		}

		if !valid {
			errs = append(errs, fmt.Errorf("%s is an invalid value for argument %s", value, k))
		}
		return wrapErrors(errs)
	}
}

// return the pieces of id `left:right` as left, right
func parseTwoPartID(id, left, right string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("unexpected ID format (%q); expected %s:%s", id, left, right)
	}

	return parts[0], parts[1], nil
}

// format the strings into an id `a:b`
func buildTwoPartID(a, b string) string {
	return fmt.Sprintf("%s:%s", a, b)
}

// return the pieces of id `left:center:right` as left, center, right
func parseThreePartID(id, left, center, right string) (string, string, string, error) {
	parts := strings.SplitN(id, ":", 3)
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("unexpected ID format (%q). Expected %s:%s:%s", id, left, center, right)
	}

	return parts[0], parts[1], parts[2], nil
}

// format the strings into an id `a:b:c`
func buildThreePartID(a, b, c string) string {
	return fmt.Sprintf("%s:%s:%s", a, b, c)
}

func buildChecksumID(v []string) string {
	sort.Strings(v)

	h := md5.New()
	// Hash.Write never returns an error. See https://pkg.go.dev/hash#Hash
	_, _ = h.Write([]byte(strings.Join(v, "")))
	bs := h.Sum(nil)

	return fmt.Sprintf("%x", bs)
}

func expandStringList(configured []interface{}) []string {
	vs := make([]string, 0, len(configured))
	for _, v := range configured {
		val, ok := v.(string)
		if ok && val != "" {
			vs = append(vs, val)
		}
	}
	return vs
}

func flattenStringList(v []string) []interface{} {
	c := make([]interface{}, 0, len(v))
	for _, s := range v {
		c = append(c, s)
	}
	return c
}

func unconvertibleIdErr(id string, err error) *unconvertibleIdError {
	return &unconvertibleIdError{OriginalId: id, OriginalError: err}
}

type unconvertibleIdError struct {
	OriginalId    string
	OriginalError error
}

func (e *unconvertibleIdError) Error() string {
	return fmt.Sprintf("Unexpected ID format (%q), expected numerical ID. %s",
		e.OriginalId, e.OriginalError.Error())
}

func validateTeamIDFunc(v interface{}, keyName string) (we []string, errors []error) {
	teamIDString, ok := v.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %s to be string", keyName)}
	}
	// Check that the team ID can be converted to an int
	if _, err := strconv.ParseInt(teamIDString, 10, 64); err != nil {
		return nil, []error{unconvertibleIdErr(teamIDString, err)}
	}

	return
}

func splitRepoFilePath(path string) (string, string) {
	parts := strings.Split(path, "/")
	return parts[0], strings.Join(parts[1:], "/")
}

func getTeamID(teamIDString string, meta interface{}) (int64, error) {
	// Given a string that is either a team id or team slug, return the
	// id of the team it is referring to.
	ctx := context.Background()
	client := meta.(*Owner).v3client
	orgName := meta.(*Owner).name

	teamId, parseIntErr := strconv.ParseInt(teamIDString, 10, 64)
	if parseIntErr == nil {
		return teamId, nil
	}

	// The given id not an integer, assume it is a team slug
	team, _, slugErr := client.Teams.GetTeamBySlug(ctx, orgName, teamIDString)
	if slugErr != nil {
		return -1, errors.New(parseIntErr.Error() + slugErr.Error())
	}
	return team.GetID(), nil
}

func getTeamSlug(teamIDString string, meta interface{}) (string, error) {
	// Given a string that is either a team id or team slug, return the
	// team slug it is referring to.
	ctx := context.Background()
	client := meta.(*Owner).v3client
	orgName := meta.(*Owner).name
	orgId := meta.(*Owner).id

	teamId, parseIntErr := strconv.ParseInt(teamIDString, 10, 64)
	if parseIntErr != nil {
		// The given id not an integer, assume it is a team slug
		team, _, slugErr := client.Teams.GetTeamBySlug(ctx, orgName, teamIDString)
		if slugErr != nil {
			return "", errors.New(parseIntErr.Error() + slugErr.Error())
		}
		return team.GetSlug(), nil
	}

	// The given id is an integer, assume it is a team id
	team, _, teamIdErr := client.Teams.GetTeamByID(ctx, orgId, teamId)
	if teamIdErr != nil {
		// There isn't a team with the given ID, assume it is a teamslug
		team, _, slugErr := client.Teams.GetTeamBySlug(ctx, orgName, teamIDString)
		if slugErr != nil {
			return "", errors.New(teamIdErr.Error() + slugErr.Error())
		}
		return team.GetSlug(), nil
	}
	return team.GetSlug(), nil
}

// https://docs.github.com/en/actions/reference/encrypted-secrets#naming-your-secrets
var secretNameRegexp = regexp.MustCompile("^[a-zA-Z_][a-zA-Z0-9_]*$")

func validateSecretNameFunc(v interface{}, path cty.Path) diag.Diagnostics {
	errs := make([]error, 0)
	name, ok := v.(string)
	if !ok {
		return wrapErrors([]error{fmt.Errorf("expected type of %s to be string", path)})
	}

	if !secretNameRegexp.MatchString(name) {
		errs = append(errs, errors.New("secret names can only contain alphanumeric characters or underscores and must not start with a number"))
	}

	if strings.HasPrefix(strings.ToUpper(name), "GITHUB_") {
		errs = append(errs, errors.New("secret names must not start with the GITHUB_ prefix"))
	}

	return wrapErrors(errs)
}

// deleteResourceOn404AndSwallow304OtherwiseReturnError will log and delete resource if error is 404 which indicates resource (or any of its ancestors)
// doesn't exist.
// resourceDescription represents a formatting string that represents the resource
// args will be passed to resourceDescription in `log.Printf`
func deleteResourceOn404AndSwallow304OtherwiseReturnError(err error, d *schema.ResourceData, resourceDescription string, args ...interface{}) error {
	if ghErr, ok := err.(*github.ErrorResponse); ok {
		if ghErr.Response.StatusCode == http.StatusNotModified {
			return nil
		}
		if ghErr.Response.StatusCode == http.StatusNotFound {
			log.Printf("[INFO] Removing "+resourceDescription+" from state because it no longer exists in GitHub",
				args...)
			d.SetId("")
			return nil
		}
	}
	return err
}
