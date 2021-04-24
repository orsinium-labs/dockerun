package dockerun

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
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
LABEL package-name="{{.Package}}"
WORKDIR /opt/
RUN {{.Install}}
ENTRYPOINT {{.DEPoint}}
`

type Builder struct {
	Prefix     string // image name prefix
	Name       string // image name
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

func (b *Builder) Parse(args []string) error {

	flags := pflag.NewFlagSet("dockerun", pflag.ContinueOnError)
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
	flags.StringVar(&b.EntryPoint, "entrypoint", b.EntryPoint,
		"docker entrypoint, the base command to execute")
	flags.BoolVar(&b.Debug, "debug", b.Debug,
		"enable debug output")
	b.Docker.AddFlags(flags)
	err := flags.Parse(args)
	if err != nil {
		return err
	}
	b.Package = flags.Arg(0)
	if b.Package == "" {
		return errors.New("package name is required")
	}
	return nil
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
	dFileBuf := bytes.NewBuffer(nil)
	err = t.Execute(dFileBuf, b)
	if err != nil {
		return fmt.Errorf("generate Dockerfile: %v", err)
	}
	if b.Debug {
		fmt.Println(dFileBuf.String())
	}

	// generate tar archive with the file
	tarHeader := &tar.Header{
		Name: "Dockerfile",
		Size: int64(dFileBuf.Len()),
	}
	tarBuf := bytes.NewBuffer(nil)
	tarW := tar.NewWriter(tarBuf)
	err = tarW.WriteHeader(tarHeader)
	if err != nil {
		log.Fatal(err, "write tar header")
	}
	_, err = tarW.Write(dFileBuf.Bytes())
	if err != nil {
		log.Fatal(err, "write tar body")
	}

	// build image
	b.Logger.Debug("building image", zap.String("name", b.Name))
	bopts := b.Docker.Build()
	bopts.Tags = []string{b.Name}
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
