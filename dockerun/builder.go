package dockerun

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"text/template"

	"github.com/docker/docker/client"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

const dockerfile = `
FROM {{.Image}}:{{.Tag}}
LABEL generated-by="dockerun"
WORKDIR /opt/
RUN {{.Install}}
ENTRYPOINT {{.entrypoint}}
`

type Builder struct {
	Prefix     string // image name prefix
	Name       string // image name
	Image      string // base image name
	Tag        string // base image tag
	Install    string // command to run when building the container
	Package    string // package name to install
	EntryPoint string // comand name to run

	Docker *DockerConfig
	Logger *zap.Logger
}

func (b Builder) format(pattern string) (string, error) {
	t, err := template.New("_").Parse(pattern)
	if err != nil {
		return pattern, fmt.Errorf("parse template: %v", err)
	}
	w := bytes.NewBufferString("")
	err = t.Execute(w, b)
	if err != nil {
		return pattern, fmt.Errorf("execute template: %v", err)
	}
	return w.String(), nil
}

func (b Builder) Format() error {
	var err error
	b.Name, err = b.format(b.Name)
	if err != nil {
		return fmt.Errorf("format name: %v", err)
	}
	b.Install, err = b.format(b.Install)
	if err != nil {
		return fmt.Errorf("format install: %v", err)
	}
	return nil
}

func NewBuilder(args []string) (Builder, error) {
	var err error

	// get preset
	flags := pflag.NewFlagSet("dockerun", pflag.ContinueOnError)
	var presetName string
	flags.StringVar(&presetName, "preset", "", "configuration preset to use")
	err = flags.Parse(args)
	if err != nil {
		return Builder{}, err
	}
	preset, ok := Presets[presetName]
	if !ok {
		return Builder{}, errors.New("preset is not found")
	}
	b := preset()
	b.Logger, err = zap.NewDevelopment()
	if err != nil {
		return Builder{}, fmt.Errorf("create logger: %v", err)
	}
	return b, nil
}

func (b *Builder) Parse(args []string) error {

	flags := pflag.NewFlagSet("dockerun", pflag.ContinueOnError)
	var presetName string
	flags.StringVar(&presetName, "preset", "", "configuration preset to use")
	flags.StringVar(&b.Prefix, "prefix", b.Prefix,
		"prefix to use for the local image name")
	flags.StringVar(&b.Name, "name", b.Name,
		"name of the local image")
	flags.StringVar(&b.Image, "image", b.Image,
		"name of the image to pull")
	flags.StringVar(&b.Tag, "tag", b.Tag,
		"tag (version) of the image to pull")
	flags.StringVar(&b.Install, "install", b.Install,
		"command to execute when building the image")
	flags.StringVar(&b.Package, "package", b.Package,
		"package name to install")
	flags.StringVar(&b.EntryPoint, "entrypoint", b.EntryPoint,
		"docker entrypoint, the base command to execute")
	b.Docker.AddFlags(flags)
	return flags.Parse(args)
}

func (b Builder) Build() error {
	var err error
	ctx := context.Background()

	cl, err := client.NewClientWithOpts(b.Docker.Client()...)
	if err != nil {
		return fmt.Errorf("init Docker client: %v", err)
	}

	// pull base image
	_, err = cl.ImagePull(ctx, fmt.Sprintf("%s:%s", b.Image, b.Tag), b.Docker.Pull())
	if err != nil {
		return fmt.Errorf("pull image: %v", err)
	}

	// generate dockerfile
	t, err := template.New("Dockerfile").Parse(dockerfile)
	if err != nil {
		return fmt.Errorf("parse Dockerfile: %v", err)
	}
	dfileR, dfileW := io.Pipe()
	var terr error
	go func() {
		terr = t.Execute(dfileW, b)
		dfileW.Close()
	}()

	// build image
	bopts := b.Docker.Build()
	bopts.Dockerfile = ""
	_, err = cl.ImageBuild(ctx, dfileR, bopts)
	if err != nil {
		return fmt.Errorf("build image: %v", err)
	}
	if terr != nil {
		return fmt.Errorf("generate Dockerfile: %v", terr)
	}

	return nil
}
