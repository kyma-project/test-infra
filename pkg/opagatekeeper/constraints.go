package opagatekeeper

type K8sPSPPrivilegedContainer struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace"`
	} `yaml:"metadata"`
	Spec struct {
		EnforcementAction string `yaml:"enforcementAction"`
		Match             struct {
			Namespaces []string `yaml:"namespaces"`
			Kinds      []struct {
				APIGroups []string `yaml:"apiGroups"`
				Kinds     []string `yaml:"kinds"`
			}
		}
		Parameters struct {
			ExemptImages []string `yaml:"exemptImages"`
		} `yaml:"parameters"`
	} `yaml:"spec"`
}
# (2025-03-04)