package dockerun

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
)

type Images struct {
	Docker *DockerConfig
}

func NewImages() Images {
	return Images{
		Docker: &DockerConfig{},
	}
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
