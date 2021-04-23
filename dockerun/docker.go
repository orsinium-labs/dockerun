package dockerun

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/spf13/pflag"
)

type DockerConfig struct {
	Host         string
	RegistryAuth string
	Platform     string
	ShmSize      int64
}

func (d DockerConfig) AddFlags(flags *pflag.FlagSet) {
	flags.StringVar(&d.Host, "host", d.Host,
		"docker daemon socket to connect to")
	flags.StringVar(&d.RegistryAuth, "registry-auth", d.RegistryAuth,
		"base64 encoded credentials for the docker registry")
	flags.StringVar(&d.RegistryAuth, "platform", d.RegistryAuth,
		"image platform if server is multi-platform capable")
	flags.Int64Var(&d.ShmSize, "shm-size", d.ShmSize,
		"size of /dev/shm")
}

func (d DockerConfig) Client() []client.Opt {
	opts := make([]client.Opt, 0)
	opts = append(opts, client.FromEnv, client.WithAPIVersionNegotiation())
	if d.Host != "" {
		opts = append(opts, client.WithHost(d.Host))
	}
	return opts
}

func (d DockerConfig) Pull() types.ImagePullOptions {
	return types.ImagePullOptions{
		RegistryAuth: d.RegistryAuth,
		Platform:     d.Platform,
	}
}

func (d DockerConfig) Build() types.ImageBuildOptions {
	return types.ImageBuildOptions{
		// Squash:  true,
		ShmSize: d.ShmSize,
	}
}

func (d DockerConfig) Images() types.ImageListOptions {
	filters := filters.NewArgs()
	filters.Add("label", `generated-by=dockerun`)
	return types.ImageListOptions{Filters: filters}
}

func (d DockerConfig) Remove() types.ImageRemoveOptions {
	return types.ImageRemoveOptions{Force: true}
}
