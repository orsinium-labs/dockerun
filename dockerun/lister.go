package dockerun

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
	"go.uber.org/zap"
)

type Images struct {
	Docker *DockerConfig
	Logger *zap.Logger
}

func NewImages() (Images, error) {
	var err error
	im := Images{
		Docker: &DockerConfig{},
	}
	im.Logger, err = zap.NewDevelopment()
	if err != nil {
		return Images{}, fmt.Errorf("create logger: %v", err)
	}
	return im, nil
}

func (im Images) List() ([]string, error) {
	var err error
	ctx := context.Background()

	cl, err := client.NewClientWithOpts(im.Docker.Client()...)
	if err != nil {
		return nil, fmt.Errorf("init Docker client: %v", err)
	}

	// get images
	images, err := cl.ImageList(ctx, im.Docker.Images())
	if err != nil {
		return nil, fmt.Errorf("list images: %v", err)
	}
	result := make([]string, len(images))
	for i, image := range images {
		result[i] = image.Labels["package-name"]
	}
	return result, nil
}

func (im Images) Purge() error {
	var err error
	ctx := context.Background()

	cl, err := client.NewClientWithOpts(im.Docker.Client()...)
	if err != nil {
		return fmt.Errorf("init Docker client: %v", err)
	}

	// get images
	images, err := cl.ImageList(ctx, im.Docker.Images())
	if err != nil {
		return fmt.Errorf("list images: %v", err)
	}
	for _, image := range images {
		im.Logger.Debug("removing image", zap.String("name", image.RepoTags[0]))
		_, err = cl.ImageRemove(ctx, image.ID, im.Docker.Remove())
		if err != nil {
			return fmt.Errorf("remove image: %v", err)
		}
	}
	return nil
}
