tflint {
  required_version = ">= 0.50"
}

config {
  call_module_type = "all"
  disabled_by_default = false
}

plugin "google" {
    enabled = true
    version = "0.38.0"
    source  = "github.com/terraform-linters/tflint-ruleset-google"
}

plugin "terraform" {
  enabled = true
  preset  = "recommended"
}