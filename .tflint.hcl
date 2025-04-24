config {
  format = "compact"

  call_module_type = "local"
  force = false
  disabled_by_default = false
}

plugin "google" {
    enabled = true
    version = "0.32.0"
    source  = "github.com/terraform-linters/tflint-ruleset-google"
}

plugin "terraform" {
  enabled = true
  preset  = "recommended"
}