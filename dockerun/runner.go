package dockerun

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
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
	contID, err := runner.makeContainer(cl, image, args)
	if err != nil {
		return fmt.Errorf("make container: %v", err)
	}
	err = runner.attachContainer(cl, contID)
	if err != nil {
		return fmt.Errorf("attach container: %v", err)
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

func (runner Runner) makeContainer(cl *client.Client, image string, args []string) (string, error) {
	ctx := context.Background()
	conf := container.Config{
		Image:        image,
		Cmd:          args,
		Labels:       map[string]string{"generated-by": "dockerun"},
		AttachStdout: true,
		AttachStderr: true,
		AttachStdin:  true,
	}
	resp, err := cl.ContainerCreate(ctx, &conf, nil, nil, nil, "")
	if err != nil {
		return "", fmt.Errorf("create container: %v", err)
	}

	err = cl.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		return "", fmt.Errorf("start container: %v", err)
	}

	return resp.ID, nil
}

func (runner Runner) attachContainer(cl *client.Client, contID string) error {
	ctx := context.Background()
	atopts := types.ContainerAttachOptions{
		Stream: true,
		Stdin:  true,
		Stdout: true,
		Stderr: true,
	}
	resp, err := cl.ContainerAttach(ctx, contID, atopts)
	if err != nil {
		return fmt.Errorf("attach container: %v", err)
	}
	defer resp.Close()
	_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, resp.Reader)
	if err != nil {
		return fmt.Errorf("read container stdout: %v", err)
	}

	// wait for the container to finish
	statusCh, errCh := cl.ContainerWait(ctx, contID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("wait for container: %v", err)
		}
	case <-statusCh:
	}
	return nil
}
