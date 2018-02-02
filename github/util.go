package github

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
)

const (
	// https://developer.github.com/guides/traversing-with-pagination/#basics-of-pagination
	maxPerPage = 100
)

func toGithubID(id string) int64 {
	githubID, _ := strconv.ParseInt(id, 10, 64)
	return githubID
}

func fromGithubID(id *int64) string {
	return strconv.FormatInt(*id, 64)
}

func validateValueFunc(values []string) schema.SchemaValidateFunc {
	return func(v interface{}, k string) (we []string, errors []error) {
		value := v.(string)
		valid := false
		for _, role := range values {
			if value == role {
				valid = true
				break
			}
		}

		if !valid {
			errors = append(errors, fmt.Errorf("%s is an invalid value for argument %s", value, k))
		}
		return
	}
}

// return the pieces of id `a:b` as a, b
func parseTwoPartID(id string) (string, string) {
	parts := strings.SplitN(id, ":", 2)
	return parts[0], parts[1]
}

// format the strings into an id `a:b`
func buildTwoPartID(a, b *string) string {
	return fmt.Sprintf("%s:%s", *a, *b)
}
