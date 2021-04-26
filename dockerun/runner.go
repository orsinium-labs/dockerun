package dockerun

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/docker/docker/client"
	"go.uber.org/zap"
)

type Runner struct {
	Docker *DockerConfig
	Logger *zap.Logger
}

func NewRunner() (Runner, error) {
	var err error
	runner := Runner{
		Docker: &DockerConfig{},
	}
	runner.Logger, err = zap.NewDevelopment()
	if err != nil {
		return Runner{}, fmt.Errorf("create logger: %v", err)
	}
	return runner, nil
}

func (runner Runner) Run(name string, args []string) error {
	cl, err := client.NewClientWithOpts(runner.Docker.Client()...)
	if err != nil {
		return fmt.Errorf("init Docker client: %v", err)
	}
	image, err := runner.getImage(cl, name)
	if err != nil {
		return fmt.Errorf("get image: %v", err)
	}

	// using docker sdk to run commands is a lot of hassle
	cmd := exec.Command("docker", "run", "--rm", "-i", "--init", image)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	cmd.Args = append(cmd.Args, args...)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("run image: %v", err)
	}
	return nil
}

func (runner Runner) getImage(cl *client.Client, name string) (string, error) {
	ctx := context.Background()
	opts := runner.Docker.Images()
	opts.Filters.Add("label", fmt.Sprintf("package-name=%s", name))
	images, err := cl.ImageList(ctx, opts)
	if err != nil {
		return "", fmt.Errorf("list images: %v", err)
	}
	if len(images) == 0 {
		return "", errors.New("no matching images found")
	}
	if len(images) > 1 {
		return "", errors.New("more than 1 matching image found")
	}
	return images[0].ID, nil
}
