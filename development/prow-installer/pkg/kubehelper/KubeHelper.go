package kubehelper

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
)

type KubeHelper struct {
	Kubeconfig string
}

func (c *KubeHelper) Apply() *KubeApplyCommand {
	return &KubeApplyCommand{Config: c.Kubeconfig}
}

func (c *KubeHelper) Get() *KubeResourceCommand {
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
		return fmt.Errorf("error executing command: %w %v", err, out)
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
	kac.Type = "-f"
	kac.Path = path
	return kac
}

func (kac *KubeApplyCommand) Directory(path string) *KubeApplyCommand {
	kac.Type = "-k"
	kac.Path = path
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
