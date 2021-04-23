package dockecli

import (
	"fmt"
	"os"

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
		Use:   "install PRESET PACKAGE",
		Args:  cobra.MinimumNArgs(1),
		Short: "make an image for the given package",
		RunE:  install,
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

func Run() error {
	cmd := Root()
	args := make([]string, 0, len(os.Args)+1)
	if len(os.Args) > 1 {
		args = append(args, os.Args[1:2]...)
		if len(os.Args) > 2 && os.Args[2] != "--help" {
			args = append(args, "--")
		}
		args = append(args, os.Args[2:]...)
	}
	cmd.SetArgs(args)
	return cmd.Execute()
}

func install(cmd *cobra.Command, args []string) error {
	b, err := dockerun.NewBuilder(args)
	if err != nil {
		return fmt.Errorf("cannot init builder: %w", err)
	}
	err = b.Parse(args[1:])
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
