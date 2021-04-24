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
	cmd := &cobra.Command{
		Use:   "dockerun",
		Short: "install and run CLI tools using Docker",
	}

	subcmd := &cobra.Command{
		Use:   "install",
		Short: "make an image for the given package",
	}
	for name, preset := range dockerun.Presets {
		subcmd.AddCommand(&cobra.Command{
			Use: fmt.Sprintf("%s PACKAGE", name),
			// Args:  cobra.MinimumNArgs(1),
			Short: fmt.Sprintf("install %s package", name),
			RunE:  install(preset),
		})
	}
	cmd.AddCommand(subcmd)

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "list installed packages",
		RunE:  list,
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "purge",
		Short: "remove all installed packages",
		RunE:  purge,
	})
	return cmd
}

func Run() error {
	return Root().Execute()
}

func install(preset func() dockerun.Builder) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var err error
		b := preset()
		b.Logger, err = zap.NewDevelopment()
		if err != nil {
			return fmt.Errorf("create logger: %v", err)
		}
		err = b.Parse(args)
		if err != nil {
			return fmt.Errorf("parse flags: %w", err)
		}
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
