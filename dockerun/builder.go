package dockerun

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/docker/docker/client"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

const dockerfile = `
FROM {{.Image}}:{{.Tag}}
LABEL generated-by="dockerun"
LABEL io.whalebrew.name="{{.Name}}"
WORKDIR /opt/
RUN {{.Install}}
WORKDIR /workdir/
ENTRYPOINT {{.DEPoint}}
`

type Builder struct {
	Prefix     string // image name prefix
	Name       string // package name to use as label
	Image      string // base image name
	Tag        string // base image tag
	Install    string // command to run when building the container
	Package    string // package name to install
	EntryPoint string // comand name to run

	Debug  bool
	Docker *DockerConfig
	Logger *zap.Logger
}

func (b Builder) DEPoint() string {
	parts := strings.Split(b.EntryPoint, " ")
	return `["` + strings.Join(parts, `", "`) + `"]`
}

func (b Builder) format(pattern string) (string, error) {
	t, err := template.New("_").Funcs(tFuncs).Parse(pattern)
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

func (b *Builder) Format() error {
	var err error
	b.Name, err = b.format(b.Name)
	if err != nil {
		return fmt.Errorf("format name: %v", err)
	}
	b.Install, err = b.format(b.Install)
	if err != nil {
		return fmt.Errorf("format install: %v", err)
	}
	b.EntryPoint, err = b.format(b.EntryPoint)
	if err != nil {
		return fmt.Errorf("format entrypoint: %v", err)
	}
	return nil
}

func (b *Builder) AddFlags(flags *pflag.FlagSet) {
	flags.StringVar(&b.Prefix, "prefix", b.Prefix,
		"prefix to use for the local image name")
	flags.StringVar(&b.Name, "name", b.Name,
		"name of the tool used in run")
	flags.StringVar(&b.Image, "image", b.Image,
		"name of the image to pull")
	flags.StringVar(&b.Tag, "tag", b.Tag,
		"tag (version) of the image to pull")
	flags.StringVar(&b.Install, "install", b.Install,
		"command to execute when building the image")
	flags.StringVar(&b.EntryPoint, "entrypoint", b.EntryPoint,
		"docker entrypoint, the base command to execute")
	flags.BoolVar(&b.Debug, "debug", b.Debug,
		"enable debug output")
	b.Docker.AddFlags(flags)
}

func (b Builder) Build() error {
	var err error
	ctx := context.Background()

	cl, err := client.NewClientWithOpts(b.Docker.Client()...)
	if err != nil {
		return fmt.Errorf("init Docker client: %v", err)
	}
	_, err = cl.ImagePull(ctx, fmt.Sprintf("%s:%s", b.Image, b.Tag), b.Docker.Pull())
	if err != nil {
		return fmt.Errorf("pull image: %v", err)
	}
	tarBuf, err := b.layer()
	if err != nil {
		return fmt.Errorf("build image layer: %v", err)
	}

	// build image
	b.Logger.Debug("building image", zap.String("name", b.Name))
	bopts := b.Docker.Build()
	bopts.Tags = []string{
		fmt.Sprintf("%s%s:latest", b.Prefix, b.Name),
	}
	bresp, err := cl.ImageBuild(ctx, tarBuf, bopts)
	if b.Debug && bresp.Body != nil {
		_, err = io.Copy(os.Stdout, bresp.Body)
		if err != nil {
			return fmt.Errorf("read build response: %v", err)
		}
	}
	if err != nil {
		return fmt.Errorf("build image: %v", err)
	}
	return nil
}

func (b Builder) dockerfile() ([]byte, error) {
	t, err := template.New("Dockerfile").Parse(dockerfile)
	if err != nil {
		return nil, fmt.Errorf("parse Dockerfile: %v", err)
	}
	buf := bytes.NewBuffer(nil)
	err = t.Execute(buf, b)
	if err != nil {
		return nil, fmt.Errorf("generate Dockerfile: %v", err)
	}
	if b.Debug {
		fmt.Println(buf.String())
	}
	return buf.Bytes(), nil
}

func (b Builder) layer() (io.Reader, error) {
	dfile, err := b.dockerfile()
	if err != nil {
		return nil, err
	}

	tarHeader := &tar.Header{
		Name: "Dockerfile",
		Size: int64(len(dfile)),
	}
	tarBuf := bytes.NewBuffer(nil)
	tarW := tar.NewWriter(tarBuf)
	err = tarW.WriteHeader(tarHeader)
	if err != nil {
		return nil, fmt.Errorf("write tar header: %v", err)
	}
	_, err = tarW.Write(dfile)
	if err != nil {
		return nil, fmt.Errorf("write tar body: %v", err)
	}
	return tarBuf, nil
}
