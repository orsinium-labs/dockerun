package dockecli

import (
	"fmt"

	"github.com/orsinium-labs/dockerun/dockerun"
	"github.com/spf13/cobra"
)

type Command struct {
	Descr string
	Run   func([]string) error
}

func Root() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dockerun",
		Short: "Install and run CLI tools using Docker",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "install",
		Short: "make an image for the given package",
		RunE:  build,
	})
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
	cmd.AddCommand(&cobra.Command{
		Use:   "presets",
		Short: "list all available presets",
		RunE:  presets,
	})
	return cmd
}

func build(cmd *cobra.Command, args []string) error {
	b, err := dockerun.NewBuilder(args[1:])
	if err != nil {
		return fmt.Errorf("cannot init builder: %w", err)
	}
	err = b.Parse(args)
	if err != nil {
		return fmt.Errorf("cannot parse flags: %w", err)
	}
	err = b.Format()
	if err != nil {
		return fmt.Errorf("cannot format options: %w", err)
	}
	err = b.Build()
	if err != nil {
		return fmt.Errorf("cannot install: %w", err)
	}
	return nil
}

func list(cmd *cobra.Command, args []string) error {
	lister, err := dockerun.NewImages()
	if err != nil {
		return fmt.Errorf("cannot init images: %w", err)
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
		return fmt.Errorf("cannot init images: %w", err)
	}
	return lister.Purge()
}

func presets(cmd *cobra.Command, args []string) error {
	for name := range dockerun.Presets {
		fmt.Println(name)
	}
	return nil
}
