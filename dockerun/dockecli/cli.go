package dockecli

import (
	"errors"
	"fmt"

	"github.com/orsinium-labs/dockerun/dockerun"
)

type Command struct {
	Descr string
	Run   func([]string) error
}

func build(args []string) error {
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

func list(args []string) error {
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

func purge(args []string) error {
	lister, err := dockerun.NewImages()
	if err != nil {
		return fmt.Errorf("cannot init images: %w", err)
	}
	return lister.Purge()
}

func presets(args []string) error {
	for name := range dockerun.Presets {
		fmt.Println(name)
	}
	return nil
}

func GetCommand(args []string) (Command, error) {
	name := ""
	if len(args) > 1 {
		name = args[1]
	}
	switch name {
	case "build", "install", "i":
		cmd := Command{
			Descr: "make an image for the given package",
			Run:   build,
		}
		return cmd, nil
	case "images", "list", "l":
		cmd := Command{
			Descr: "list installed packages",
			Run:   list,
		}
		return cmd, nil
	case "purge":
		cmd := Command{
			Descr: "remove all installed packages",
			Run:   purge,
		}
		return cmd, nil
	case "presets":
		cmd := Command{
			Descr: "list all available presets",
			Run:   presets,
		}
		return cmd, nil
	case "", "--help", "help", "-h":
		return Command{}, errors.New("Available commands: install, list")
	default:
		return Command{}, fmt.Errorf("Unknown command: %s\n", name)
	}
}
