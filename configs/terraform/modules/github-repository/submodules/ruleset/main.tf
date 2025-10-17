resource "github_repository_ruleset" "this" {
  name        = var.ruleset.name
  repository  = var.repository_name
  target      = var.ruleset.target
  enforcement = var.ruleset.enforcement

  conditions {
    ref_name {
      include = var.ruleset.conditions.ref_name.include
      exclude = var.ruleset.conditions.ref_name.exclude
    }
  }

  rules {
    # Branch protection rules
    creation                = try(var.ruleset.rules.creation, null)
    update                  = try(var.ruleset.rules.update, null)
    deletion                = try(var.ruleset.rules.deletion, null)
    non_fast_forward        = try(var.ruleset.rules.non_fast_forward, null)
    required_linear_history = try(var.ruleset.rules.required_linear_history, null)
    required_signatures     = try(var.ruleset.rules.required_signatures, null)

    # Branch naming pattern rule
    dynamic "branch_name_pattern" {
      for_each = try(var.ruleset.rules.branch_name_pattern, null) != null ? [var.ruleset.rules.branch_name_pattern] : []
      content {
        operator = try(branch_name_pattern.value.operator, "regex")
        pattern  = try(branch_name_pattern.value.pattern, "")
        name     = try(branch_name_pattern.value.name, null)
        negate   = try(branch_name_pattern.value.negate, false)
      }
    }

    # Pull request rules
    dynamic "pull_request" {
      for_each = var.ruleset.rules.pull_request != null ? [var.ruleset.rules.pull_request] : []
      content {
        dismiss_stale_reviews_on_push     = try(pull_request.value.dismiss_stale_reviews_on_push, null)
        require_code_owner_review         = try(pull_request.value.require_code_owner_review, null)
        require_last_push_approval        = try(pull_request.value.require_last_push_approval, null)
        required_approving_review_count   = try(pull_request.value.required_approving_review_count, null)
        required_review_thread_resolution = try(pull_request.value.required_review_thread_resolution, null)
      }
    }

    # Status check rules
    dynamic "required_status_checks" {
      for_each = var.ruleset.rules.required_status_checks != null ? [var.ruleset.rules.required_status_checks] : []
      content {
        dynamic "required_check" {
          for_each = required_status_checks.value.required_check
          content {
            context        = required_check.value.context
            integration_id = try(required_check.value.integration_id, null)
          }
        }
        strict_required_status_checks_policy = try(required_status_checks.value.strict_required_status_checks_policy, null)
      }
    }

    # Merge queue rules
    dynamic "merge_queue" {
      for_each = var.ruleset.rules.merge_queue != null ? [var.ruleset.rules.merge_queue] : []
      content {
        check_response_timeout_minutes       = try(merge_queue.value.check_response_timeout_minutes, null)
        grouping_strategy                   = try(merge_queue.value.grouping_strategy, null)
        max_entries_to_build                = try(merge_queue.value.max_entries_to_build, null)
        max_entries_to_merge                = try(merge_queue.value.max_entries_to_merge, null)
        merge_method                        = try(merge_queue.value.merge_method, null)
        min_entries_to_merge                = try(merge_queue.value.min_entries_to_merge, null)
        min_entries_to_merge_wait_minutes   = try(merge_queue.value.min_entries_to_merge_wait_minutes, null)
      }
    }
  }

}