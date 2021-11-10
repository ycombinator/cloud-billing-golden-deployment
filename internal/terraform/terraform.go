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
	cmd.Dir = w.dir

	return cmd.Run()
}

// Apply runs `terraform apply` in the working directory.
func (w *WorkDir) Apply() error {
	cmd := exec.Command("terraform", "apply", "--auto-approve=true")
	cmd.Dir = w.dir

	return cmd.Run()
}

// Destroy runs `terraform destroy` in the working directory.
func (w *WorkDir) Destroy() error {
	cmd := exec.Command("terraform", "destroy", "--auto-approve=true")
	cmd.Dir = w.dir

	return cmd.Run()
}
