package dockecli

import (
	"fmt"

	"github.com/orsinium-labs/dockerun/dockerun"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type Command struct {
	Descr string
	Run   func([]string) error
}

func Root() *cobra.Command {
	cmdRoot := &cobra.Command{
		Use:   "dockerun",
		Short: "build Docker images for CLI tools",
	}

	cmdBuild := &cobra.Command{
		Use:   "build",
		Short: "build an image for the given package",
	}
	for name, preset := range dockerun.Presets {
		builder := preset()
		cmdPreset := &cobra.Command{
			Use:   fmt.Sprintf("%s PACKAGE", name),
			Args:  cobra.MinimumNArgs(1),
			Short: fmt.Sprintf("install %s package", name),
			RunE:  build(&builder),
		}
		cmdPreset.Flags().SortFlags = false
		builder.AddFlags(cmdPreset.Flags())
		cmdBuild.AddCommand(cmdPreset)
	}
	cmdRoot.AddCommand(cmdBuild)

	cmdRoot.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "list installed packages",
		RunE:  list,
	})
	cmdRoot.AddCommand(&cobra.Command{
		Use:   "purge",
		Short: "remove all installed packages",
		RunE:  purge,
	})
	return cmdRoot
}

func build(b *dockerun.Builder) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var err error
		b.Logger, err = zap.NewDevelopment()
		if err != nil {
			return fmt.Errorf("create logger: %v", err)
		}
		b.Package = args[0]
		err = b.Format()
		if err != nil {
			return fmt.Errorf("format options: %w", err)
		}
		err = b.Build()
		if err != nil {
			return fmt.Errorf("install: %w", err)
		}
		return nil
	}
}

func list(cmd *cobra.Command, args []string) error {
	lister, err := dockerun.NewImages()
	if err != nil {
		return fmt.Errorf("init images: %w", err)
	}
	images, err := lister.List()
	if err != nil {
		return err
	}
	for _, line := range images {
		fmt.Println(line)
	}
	return nil
}

func purge(cmd *cobra.Command, args []string) error {
	lister, err := dockerun.NewImages()
	if err != nil {
		return fmt.Errorf("init images: %w", err)
	}
	return lister.Purge()
}
