package github

import (
	"encoding/json"
	"log"
	"reflect"
	"sort"

	"github.com/google/go-github/v66/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGithubRulesetObject(d *schema.ResourceData, org string) *github.Ruleset {
	isOrgLevel := len(org) > 0

	var source, sourceType string
	if isOrgLevel {
		source = org
		sourceType = "Organization"
	} else {
		source = d.Get("repository").(string)
		sourceType = "Repository"
	}

	return &github.Ruleset{
		Name:         d.Get("name").(string),
		Target:       github.String(d.Get("target").(string)),
		Source:       source,
		SourceType:   &sourceType,
		Enforcement:  d.Get("enforcement").(string),
		BypassActors: expandBypassActors(d.Get("bypass_actors").([]interface{})),
		Conditions:   expandConditions(d.Get("conditions").([]interface{}), isOrgLevel),
		Rules:        expandRules(d.Get("rules").([]interface{}), isOrgLevel),
	}
}

func expandBypassActors(input []interface{}) []*github.BypassActor {
	if len(input) == 0 {
		return nil
	}
	bypassActors := make([]*github.BypassActor, 0)

	for _, v := range input {
		inputMap := v.(map[string]interface{})
		actor := &github.BypassActor{}
		if v, ok := inputMap["actor_id"].(int); ok {
			actor.ActorID = github.Int64(int64(v))
		}

		if v, ok := inputMap["actor_type"].(string); ok {
			actor.ActorType = &v
		}

		if v, ok := inputMap["bypass_mode"].(string); ok {
			actor.BypassMode = &v
		}
		bypassActors = append(bypassActors, actor)
	}
	return bypassActors
}

func flattenBypassActors(bypassActors []*github.BypassActor) []interface{} {
	if bypassActors == nil {
		return []interface{}{}
	}

	actorsSlice := make([]interface{}, 0)
	for _, v := range bypassActors {
		actorMap := make(map[string]interface{})

		actorMap["actor_id"] = v.GetActorID()
		actorMap["actor_type"] = v.GetActorType()
		actorMap["bypass_mode"] = v.GetBypassMode()

		actorsSlice = append(actorsSlice, actorMap)
	}

	return actorsSlice
}

func expandConditions(input []interface{}, org bool) *github.RulesetConditions {
	if len(input) == 0 || input[0] == nil {
		return nil
	}
	rulesetConditions := &github.RulesetConditions{}
	inputConditions := input[0].(map[string]interface{})

	// ref_name is available for both repo and org rulesets
	if v, ok := inputConditions["ref_name"].([]interface{}); ok && v != nil && len(v) != 0 {
		inputRefName := v[0].(map[string]interface{})
		include := make([]string, 0)
		exclude := make([]string, 0)

		for _, v := range inputRefName["include"].([]interface{}) {
			if v != nil {
				include = append(include, v.(string))
			}
		}

		for _, v := range inputRefName["exclude"].([]interface{}) {
			if v != nil {
				exclude = append(exclude, v.(string))
			}
		}

		rulesetConditions.RefName = &github.RulesetRefConditionParameters{
			Include: include,
			Exclude: exclude,
		}
	}

	// org-only fields
	if org {
		// repository_name and repository_id
		if v, ok := inputConditions["repository_name"].([]interface{}); ok && v != nil && len(v) != 0 {
			inputRepositoryName := v[0].(map[string]interface{})
			include := make([]string, 0)
			exclude := make([]string, 0)

			for _, v := range inputRepositoryName["include"].([]interface{}) {
				if v != nil {
					include = append(include, v.(string))
				}
			}

			for _, v := range inputRepositoryName["exclude"].([]interface{}) {
				if v != nil {
					exclude = append(exclude, v.(string))
				}
			}

			protected := inputRepositoryName["protected"].(bool)

			rulesetConditions.RepositoryName = &github.RulesetRepositoryNamesConditionParameters{
				Include:   include,
				Exclude:   exclude,
				Protected: &protected,
			}
		} else if v, ok := inputConditions["repository_id"].([]interface{}); ok && v != nil && len(v) != 0 {
			repositoryIDs := make([]int64, 0)

			for _, v := range v {
				if v != nil {
					repositoryIDs = append(repositoryIDs, int64(v.(int)))
				}
			}

			rulesetConditions.RepositoryID = &github.RulesetRepositoryIDsConditionParameters{RepositoryIDs: repositoryIDs}
		}
	}

	return rulesetConditions
}

func flattenConditions(conditions *github.RulesetConditions, org bool) []interface{} {
	if conditions == nil || conditions.RefName == nil {
		return []interface{}{}
	}

	conditionsMap := make(map[string]interface{})
	refNameSlice := make([]map[string]interface{}, 0)

	refNameSlice = append(refNameSlice, map[string]interface{}{
		"include": conditions.RefName.Include,
		"exclude": conditions.RefName.Exclude,
	})

	conditionsMap["ref_name"] = refNameSlice

	// org-only fields
	if org {
		repositoryNameSlice := make([]map[string]interface{}, 0)

		if conditions.RepositoryName != nil {
			var protected bool

			if conditions.RepositoryName.Protected != nil {
				protected = *conditions.RepositoryName.Protected
			}

			repositoryNameSlice = append(repositoryNameSlice, map[string]interface{}{
				"include":   conditions.RepositoryName.Include,
				"exclude":   conditions.RepositoryName.Exclude,
				"protected": protected,
			})
			conditionsMap["repository_name"] = repositoryNameSlice
		}

		if conditions.RepositoryID != nil {
			conditionsMap["repository_id"] = conditions.RepositoryID.RepositoryIDs
		}
	}

	return []interface{}{conditionsMap}
}

func expandRules(input []interface{}, org bool) []*github.RepositoryRule {
	if len(input) == 0 || input[0] == nil {
		return nil
	}

	rulesMap := input[0].(map[string]interface{})
	rulesSlice := make([]*github.RepositoryRule, 0)

	// First we expand rules without parameters
	if v, ok := rulesMap["creation"].(bool); ok && v {
		rulesSlice = append(rulesSlice, github.NewCreationRule())
	}

	if v, ok := rulesMap["update"].(bool); ok && v {
		params := github.UpdateAllowsFetchAndMergeRuleParameters{}
		if fetchAndMerge, ok := rulesMap["update"].(bool); ok && fetchAndMerge {
			params.UpdateAllowsFetchAndMerge = true
		} else {
			params.UpdateAllowsFetchAndMerge = false
		}
		rulesSlice = append(rulesSlice, github.NewUpdateRule(&params))
	}

	if v, ok := rulesMap["deletion"].(bool); ok && v {
		rulesSlice = append(rulesSlice, github.NewDeletionRule())
	}

	if v, ok := rulesMap["required_linear_history"].(bool); ok && v {
		rulesSlice = append(rulesSlice, github.NewRequiredLinearHistoryRule())
	}

	if v, ok := rulesMap["required_signatures"].(bool); ok && v {
		rulesSlice = append(rulesSlice, github.NewRequiredSignaturesRule())
	}

	if v, ok := rulesMap["non_fast_forward"].(bool); ok && v {
		rulesSlice = append(rulesSlice, github.NewNonFastForwardRule())
	}

	// Required deployments rule
	if !org {
		if v, ok := rulesMap["required_deployments"].([]interface{}); ok && len(v) != 0 {
			requiredDeploymentsMap := make(map[string]interface{})
			// If the rule's block is present but has an empty environments list
			if v[0] == nil {
				requiredDeploymentsMap["required_deployment_environments"] = make([]interface{}, 0)
			} else {
				requiredDeploymentsMap = v[0].(map[string]interface{})
			}
			envs := make([]string, 0)
			for _, v := range requiredDeploymentsMap["required_deployment_environments"].([]interface{}) {
				envs = append(envs, v.(string))
			}

			params := &github.RequiredDeploymentEnvironmentsRuleParameters{
				RequiredDeploymentEnvironments: envs,
			}

			rulesSlice = append(rulesSlice, github.NewRequiredDeploymentsRule(params))
		}
	}

	// Pattern parameter rules
	for _, k := range []string{"commit_message_pattern", "commit_author_email_pattern", "committer_email_pattern", "branch_name_pattern", "tag_name_pattern"} {
		if v, ok := rulesMap[k].([]interface{}); ok && len(v) != 0 {
			patternParametersMap := v[0].(map[string]interface{})

			name := patternParametersMap["name"].(string)
			negate := patternParametersMap["negate"].(bool)

			params := &github.RulePatternParameters{
				Name:     &name,
				Negate:   &negate,
				Operator: patternParametersMap["operator"].(string),
				Pattern:  patternParametersMap["pattern"].(string),
			}

			switch k {
			case "commit_message_pattern":
				rulesSlice = append(rulesSlice, github.NewCommitMessagePatternRule(params))
			case "commit_author_email_pattern":
				rulesSlice = append(rulesSlice, github.NewCommitAuthorEmailPatternRule(params))
			case "committer_email_pattern":
				rulesSlice = append(rulesSlice, github.NewCommitterEmailPatternRule(params))
			case "branch_name_pattern":
				rulesSlice = append(rulesSlice, github.NewBranchNamePatternRule(params))
			case "tag_name_pattern":
				rulesSlice = append(rulesSlice, github.NewTagNamePatternRule(params))
			}
		}
	}

	// Pull request rule
	if v, ok := rulesMap["pull_request"].([]interface{}); ok && len(v) != 0 {
		pullRequestMap := v[0].(map[string]interface{})
		params := &github.PullRequestRuleParameters{
			DismissStaleReviewsOnPush:      pullRequestMap["dismiss_stale_reviews_on_push"].(bool),
			RequireCodeOwnerReview:         pullRequestMap["require_code_owner_review"].(bool),
			RequireLastPushApproval:        pullRequestMap["require_last_push_approval"].(bool),
			RequiredApprovingReviewCount:   pullRequestMap["required_approving_review_count"].(int),
			RequiredReviewThreadResolution: pullRequestMap["required_review_thread_resolution"].(bool),
		}

		rulesSlice = append(rulesSlice, github.NewPullRequestRule(params))
	}

	// Merge queue rule
	if v, ok := rulesMap["merge_queue"].([]interface{}); ok && len(v) != 0 {
		mergeQueueMap := v[0].(map[string]interface{})
		params := &github.MergeQueueRuleParameters{
			CheckResponseTimeoutMinutes:  mergeQueueMap["check_response_timeout_minutes"].(int),
			GroupingStrategy:             mergeQueueMap["grouping_strategy"].(string),
			MaxEntriesToBuild:            mergeQueueMap["max_entries_to_build"].(int),
			MaxEntriesToMerge:            mergeQueueMap["max_entries_to_merge"].(int),
			MergeMethod:                  mergeQueueMap["merge_method"].(string),
			MinEntriesToMerge:            mergeQueueMap["min_entries_to_merge"].(int),
			MinEntriesToMergeWaitMinutes: mergeQueueMap["min_entries_to_merge_wait_minutes"].(int),
		}

		rulesSlice = append(rulesSlice, github.NewMergeQueueRule(params))
	}

	// Required status checks rule
	if v, ok := rulesMap["required_status_checks"].([]interface{}); ok && len(v) != 0 {
		requiredStatusMap := v[0].(map[string]interface{})
		requiredStatusChecks := make([]github.RuleRequiredStatusChecks, 0)

		if requiredStatusChecksInput, ok := requiredStatusMap["required_check"]; ok {

			requiredStatusChecksSet := requiredStatusChecksInput.(*schema.Set)
			for _, checkMap := range requiredStatusChecksSet.List() {
				check := checkMap.(map[string]interface{})
				integrationID := github.Int64(int64(check["integration_id"].(int)))

				params := github.RuleRequiredStatusChecks{
					Context: check["context"].(string),
				}

				if *integrationID != 0 {
					params.IntegrationID = integrationID
				}

				requiredStatusChecks = append(requiredStatusChecks, params)
			}
		}

		params := &github.RequiredStatusChecksRuleParameters{
			RequiredStatusChecks:             requiredStatusChecks,
			StrictRequiredStatusChecksPolicy: requiredStatusMap["strict_required_status_checks_policy"].(bool),
		}
		rulesSlice = append(rulesSlice, github.NewRequiredStatusChecksRule(params))
	}

	// Required workflows to pass before merging rule
	if v, ok := rulesMap["required_workflows"].([]interface{}); ok && len(v) != 0 {
		requiredWorkflowsMap := v[0].(map[string]interface{})
		requiredWorkflows := make([]*github.RuleRequiredWorkflow, 0)

		if requiredWorkflowsInput, ok := requiredWorkflowsMap["required_workflow"]; ok {

			requiredWorkflowsSet := requiredWorkflowsInput.(*schema.Set)
			for _, workflowMap := range requiredWorkflowsSet.List() {
				workflow := workflowMap.(map[string]interface{})

				// Get all parameters
				repositoryID := github.Int64(int64(workflow["repository_id"].(int)))
				ref := github.String(workflow["ref"].(string))

				params := &github.RuleRequiredWorkflow{
					RepositoryID: repositoryID,
					Path:         workflow["path"].(string),
					Ref:          ref,
				}

				requiredWorkflows = append(requiredWorkflows, params)
			}
		}

		params := &github.RequiredWorkflowsRuleParameters{
			RequiredWorkflows: requiredWorkflows,
		}
		rulesSlice = append(rulesSlice, github.NewRequiredWorkflowsRule(params))
	}

	// Required code scanning to pass before merging rule
	if v, ok := rulesMap["required_code_scanning"].([]interface{}); ok && len(v) != 0 {
		requiredCodeScanningMap := v[0].(map[string]interface{})
		requiredCodeScanningTools := make([]*github.RuleRequiredCodeScanningTool, 0)

		if requiredCodeScanningInput, ok := requiredCodeScanningMap["required_code_scanning_tool"]; ok {

			requiredCodeScanningSet := requiredCodeScanningInput.(*schema.Set)
			for _, codeScanningMap := range requiredCodeScanningSet.List() {
				codeScanningTool := codeScanningMap.(map[string]interface{})

				// Get all parameters
				alertsThreshold := github.String(codeScanningTool["alerts_threshold"].(string))
				securityAlertsThreshold := github.String(codeScanningTool["security_alerts_threshold"].(string))
				tool := github.String(codeScanningTool["tool"].(string))

				params := &github.RuleRequiredCodeScanningTool{
					AlertsThreshold:         *alertsThreshold,
					SecurityAlertsThreshold: *securityAlertsThreshold,
					Tool:                    *tool,
				}

				requiredCodeScanningTools = append(requiredCodeScanningTools, params)
			}
		}

		params := &github.RequiredCodeScanningRuleParameters{
			RequiredCodeScanningTools: requiredCodeScanningTools,
		}
		rulesSlice = append(rulesSlice, github.NewRequiredCodeScanningRule(params))
	}

	return rulesSlice
}

func flattenRules(rules []*github.RepositoryRule, org bool) []interface{} {
	if len(rules) == 0 || rules == nil {
		return []interface{}{}
	}

	rulesMap := make(map[string]interface{})
	for _, v := range rules {
		switch v.Type {
		case "creation", "deletion", "required_linear_history", "required_signatures", "non_fast_forward":
			rulesMap[v.Type] = true

		case "update":
			var params github.UpdateAllowsFetchAndMergeRuleParameters
			if v.Parameters != nil {
				err := json.Unmarshal(*v.Parameters, &params)
				if err != nil {
					log.Printf("[INFO] Unexpected error unmarshalling rule %s with parameters: %v",
						v.Type, v.Parameters)
				}
				rulesMap["update_allows_fetch_and_merge"] = params.UpdateAllowsFetchAndMerge
			} else {
				rulesMap["update_allows_fetch_and_merge"] = false
			}
			rulesMap[v.Type] = true

		case "commit_message_pattern", "commit_author_email_pattern", "committer_email_pattern", "branch_name_pattern", "tag_name_pattern":
			var params github.RulePatternParameters
			var name string
			var negate bool

			err := json.Unmarshal(*v.Parameters, &params)
			if err != nil {
				log.Printf("[INFO] Unexpected error unmarshalling rule %s with parameters: %v",
					v.Type, v.Parameters)
			}

			if params.Name != nil {
				name = *params.Name
			}
			if params.Negate != nil {
				negate = *params.Negate
			}

			rule := make(map[string]interface{})
			rule["name"] = name
			rule["negate"] = negate
			rule["operator"] = params.Operator
			rule["pattern"] = params.Pattern
			rulesMap[v.Type] = []map[string]interface{}{rule}

		case "required_deployments":
			if !org {
				var params github.RequiredDeploymentEnvironmentsRuleParameters

				err := json.Unmarshal(*v.Parameters, &params)
				if err != nil {
					log.Printf("[INFO] Unexpected error unmarshalling rule %s with parameters: %v",
						v.Type, v.Parameters)
				}

				rule := make(map[string]interface{})
				rule["required_deployment_environments"] = params.RequiredDeploymentEnvironments
				rulesMap[v.Type] = []map[string]interface{}{rule}
			}

		case "pull_request":
			var params github.PullRequestRuleParameters

			err := json.Unmarshal(*v.Parameters, &params)
			if err != nil {
				log.Printf("[INFO] Unexpected error unmarshalling rule %s with parameters: %v",
					v.Type, v.Parameters)
			}

			rule := make(map[string]interface{})
			rule["dismiss_stale_reviews_on_push"] = params.DismissStaleReviewsOnPush
			rule["require_code_owner_review"] = params.RequireCodeOwnerReview
			rule["require_last_push_approval"] = params.RequireLastPushApproval
			rule["required_approving_review_count"] = params.RequiredApprovingReviewCount
			rule["required_review_thread_resolution"] = params.RequiredReviewThreadResolution
			rulesMap[v.Type] = []map[string]interface{}{rule}

		case "required_status_checks":
			var params github.RequiredStatusChecksRuleParameters

			err := json.Unmarshal(*v.Parameters, &params)
			if err != nil {
				log.Printf("[INFO] Unexpected error unmarshalling rule %s with parameters: %v",
					v.Type, v.Parameters)
			}

			requiredStatusChecksSlice := make([]map[string]interface{}, 0)
			for _, check := range params.RequiredStatusChecks {
				integrationID := int64(0)
				if check.IntegrationID != nil {
					integrationID = *check.IntegrationID
				}
				requiredStatusChecksSlice = append(requiredStatusChecksSlice, map[string]interface{}{
					"context":        check.Context,
					"integration_id": integrationID,
				})
			}

			rule := make(map[string]interface{})
			rule["required_check"] = requiredStatusChecksSlice
			rule["strict_required_status_checks_policy"] = params.StrictRequiredStatusChecksPolicy
			rulesMap[v.Type] = []map[string]interface{}{rule}
		}
	}

	return []interface{}{rulesMap}
}

func bypassActorsDiffSuppressFunc(k, old, new string, d *schema.ResourceData) bool {
	// If the length has changed, no need to suppress
	if k == "bypass_actors.#" {
		return old == new
	}

	// Get change to bypass actors
	o, n := d.GetChange("bypass_actors")
	oldBypassActors := o.([]interface{})
	newBypassActors := n.([]interface{})

	sort.SliceStable(oldBypassActors, func(i, j int) bool {
		return oldBypassActors[i].(map[string]interface{})["actor_id"].(int) > oldBypassActors[j].(map[string]interface{})["actor_id"].(int)
	})
	sort.SliceStable(newBypassActors, func(i, j int) bool {
		return newBypassActors[i].(map[string]interface{})["actor_id"].(int) > newBypassActors[j].(map[string]interface{})["actor_id"].(int)
	})

	return reflect.DeepEqual(oldBypassActors, newBypassActors)
}
