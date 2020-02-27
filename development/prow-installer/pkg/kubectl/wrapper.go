package kubectl

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path/filepath"
)

type Wrapper struct {
	Kubeconfig string
}

func (c *Wrapper) Apply() *KubeApplyCommand {
	return &KubeApplyCommand{Config: c.Kubeconfig}
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
	if _, err := os.Stat(kac.Path); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist")
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

	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("could not get current working directory %w", err)
	}

	path := filepath.FromSlash(cwd + "/.kube/")
	file := name + "_config"

	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Debugf("%s does not exist. Creating directory...", path)
		if err := os.MkdirAll(path, 0700); err != nil {
			return "", fmt.Errorf("unexpected error during folder creation %w", err)
		}
	}
	log.Debugf("Creating path %s", path+file)
	f, err := os.Create(path + file)
	if err != nil {
		return "", fmt.Errorf("error creating kubeconfig file %w", err)
	}
	defer f.Close()

	log.Debugf("Generating GCP Kubeconfig file with credentials for host: %s, name: %s", endpoint, name)
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
