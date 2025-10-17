variable "repository_name" {
  description = "Name of the repository"
  type        = string
}

variable "rulesets" {
  description = "List of rulesets to apply to repositories"
  type = list(object({
    name        = string
    target      = string
    enforcement = string
    rules = object({
      creation                = optional(bool)
      update                  = optional(bool)
      deletion                = optional(bool)
      non_fast_forward        = optional(bool)
      required_linear_history = optional(bool)
      required_signatures     = optional(bool)
      pull_request = optional(object({
        dismiss_stale_reviews_on_push     = optional(bool)
        require_code_owner_review         = optional(bool)
        require_last_push_approval        = optional(bool)
        required_approving_review_count   = optional(number)
        required_review_thread_resolution = optional(bool)
      }))
      required_status_checks = optional(object({
        required_check = list(object({
          context        = string
          integration_id = optional(number)
        }))
        strict_required_status_checks_policy = optional(bool)
      }))
      branch_name_pattern = optional(object({
        operator = string
        pattern  = string
        name     = optional(string)
        negate   = optional(bool)
      }))
      merge_queue = optional(object({
        check_response_timeout_minutes       = optional(number)
        grouping_strategy                   = optional(string)
        max_entries_to_build                = optional(number)
        max_entries_to_merge                = optional(number)
        merge_method                        = optional(string)
        min_entries_to_merge                = optional(number)
        min_entries_to_merge_wait_minutes   = optional(number)
      }))
    })
    conditions = optional(object({
      ref_name = optional(object({
        include = optional(list(string))
        exclude = optional(list(string))
      }))
    }))
  }))
}