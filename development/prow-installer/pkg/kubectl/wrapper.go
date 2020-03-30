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
