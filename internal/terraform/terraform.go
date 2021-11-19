package terraform

import (
	"fmt"
	"os"
	"os/exec"
)

type WorkDir struct {
	dir string
}

func NewWorkDir(path string) (*WorkDir, error) {
	w := new(WorkDir)

	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("file path [%s] does not exist: %w", path, err)
	}

	if !fileInfo.IsDir() {
		return nil, fmt.Errorf(" file path [%s] is not a directory", path)
	}

	w.dir = path

	return w, nil
}

// Init runs `terraform init` in the working directory.
func (w *WorkDir) Init() error {
	cmd := exec.Command("terraform", "init")
	return w.runCmd(cmd)

}

// Apply runs `terraform apply` in the working directory.
func (w *WorkDir) Apply() error {
	cmd := exec.Command("terraform", "apply", "--auto-approve=true")
	return w.runCmd(cmd)
}

// Destroy runs `terraform destroy` in the working directory.
func (w *WorkDir) Destroy() error {
	cmd := exec.Command("terraform", "destroy", "--auto-approve=true")
	return w.runCmd(cmd)
}

func (w *WorkDir) runCmd(cmd *exec.Cmd) error {
	apiKey := os.Getenv("EC_GOLDEN_API_KEY")
	apiKeyEnvVar := fmt.Sprintf("EC_API_KEY=%s", apiKey)
	cmd.Env = append(cmd.Env, apiKeyEnvVar)

	cmd.Dir = w.dir

	return cmd.Run()
}
