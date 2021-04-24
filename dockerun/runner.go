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
	var err error
	ctx := context.Background()

	cl, err := client.NewClientWithOpts(runner.Docker.Client()...)
	if err != nil {
		return fmt.Errorf("init Docker client: %v", err)
	}

	// get the image
	opts := runner.Docker.Images()
	opts.Filters.Add("label", fmt.Sprintf("package-name=%s", name))
	images, err := cl.ImageList(ctx, opts)
	if err != nil {
		return fmt.Errorf("list images: %v", err)
	}
	if len(images) == 0 {
		return errors.New("no matching images found")
	}
	if len(images) > 1 {
		return errors.New("more than 1 matching image found")
	}
	image := images[0]

	// create container
	conf := container.Config{
		Image:        image.ID,
		Cmd:          args,
		Labels:       map[string]string{"generated-by": "dockerun"},
		AttachStdout: true,
		AttachStderr: true,
		AttachStdin:  true,
	}
	resp, err := cl.ContainerCreate(ctx, &conf, nil, nil, nil, "")
	if err != nil {
		return fmt.Errorf("create container: %v", err)
	}

	// start container
	err = cl.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		return fmt.Errorf("start container: %v", err)
	}

	// wait for the container to finish
	statusCh, errCh := cl.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("wait for container: %v", err)
		}
	case <-statusCh:
	}

	out, err := cl.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		return fmt.Errorf("read logs: %v", err)
	}
	_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	if err != nil {
		return fmt.Errorf("write logs: %v", err)
	}

	return nil
}
