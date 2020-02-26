package kubectl

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path/filepath"
)

type KubectlWrapper struct {
	Kubeconfig string
}

func (c *KubectlWrapper) Apply() *KubeApplyCommand {
	return &KubeApplyCommand{Config: c.Kubeconfig}
}

func (c *KubectlWrapper) Get() *KubeResourceCommand {
	return &KubeResourceCommand{Command: "get", Config: c.Kubeconfig}
}

type KubeResourceCommand struct {
	Command      string
	Nspace       string
	Resource     string
	ResourceName string
	Config       string
}

func (kc *KubeResourceCommand) Namespace(namespace string) *KubeResourceCommand {
	kc.Nspace = namespace
	return kc
}

func (kc *KubeResourceCommand) Name(name string) *KubeResourceCommand {
	kc.ResourceName = name
	return kc
}

func (kc *KubeResourceCommand) Pods() *KubeResourceCommand {
	kc.Resource = "pods"
	return kc
}

func (kc *KubeResourceCommand) Services() *KubeResourceCommand {
	kc.Resource = "svc"
	return kc
}

func (kc *KubeResourceCommand) Do() error {
	cmd := exec.Command("kubectl", kc.Command)
	cmd.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s", kc.Config))
	if kc.Nspace != "" {
		cmd.Args = append(cmd.Args, "-n", kc.Nspace)
	}
	if kc.Resource == "" {
		return fmt.Errorf("resource not defined")
	} else {
		cmd.Args = append(cmd.Args, kc.Resource)
	}

	log.Debugf("executing command: %s", cmd.String())
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("error executing command: %w %s", err, out)
	}
	return nil
}

type KubeApplyCommand struct {
	Type   string
	Path   string
	Nspace string
	Config string
}

func (kac *KubeApplyCommand) File(path string) *KubeApplyCommand {
	kac.Type = "-f" //file
	//change slashed path (dir/filename) to OS-dependent path.
	kac.Path = filepath.FromSlash(path)
	return kac
}

func (kac *KubeApplyCommand) Directory(path string) *KubeApplyCommand {
	kac.Type = "-k" //directory
	//change slashed path (dir/directory/) to OS-dependent path.
	kac.Path = filepath.FromSlash(path)
	return kac
}

func (kac *KubeApplyCommand) Do() error {
	cmd := exec.Command("kubectl", "apply")
	cmd.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s", kac.Config))

	if kac.Path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	cmd.Args = append(cmd.Args, kac.Type, kac.Path)

	log.Debugf("executing command: %s", cmd.String())
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("error executing command: %w %s", err, out)
	}
	return nil
}

// generate kubeconfig based on credentials provided in arguments
// the function returns path to the config file.
// it's needed to have GOOGLE_CREDENTIALS_APPLICATION env variable set
func GenerateKubeconfig(endpoint, cadata, name string) (string, error) {
	if endpoint == "" {
		return "", fmt.Errorf("endpoint cannot be empty")
	}
	if cadata == "" {
		return "", fmt.Errorf("cadata cannot be empty")
	}
	if name == "" {
		return "", fmt.Errorf("name cannot be empty")
	}

	path := filepath.FromSlash(".kube/" + name + "_config")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(".kube", 0700); err != nil {
			return "", fmt.Errorf("unexpected error during folder creation %w", err)
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("error creating kubeconfig file %w", err)
	}
	defer f.Close()
	kubeconfigTemplate := fmt.Sprintf(`apiVersion: v1
kind: Config
clusters:
- cluster:
    certificate-authority-data: %s
    server: https://%s
  name: gke-cluster
users:
- name: gke-user
  user:
    auth-provider:
      name: gcp
contexts:
- context:
    cluster: gke-cluster
    user: gke-user
  name: gke-cluster
current-context: gke-cluster`, cadata, endpoint)
	if _, err = f.WriteString(kubeconfigTemplate); err != nil {
		return "", fmt.Errorf("error writing to kubeconfig file %w", err)
	}
	return path, nil
}
