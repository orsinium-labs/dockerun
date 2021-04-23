package dockerun

import (
	"errors"
	"fmt"
	"os"
)

type Command func([]string) error

func build(args []string) error {
	b, err := NewBuilder(args)
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
	lister, err := NewImages()
	if err != nil {
		return err
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
	lister, err := NewImages()
	if err != nil {
		return err
	}
	return lister.Purge()
}

func GetCommand() (Command, error) {
	name := ""
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	switch name {
	case "build", "install", "i":
		return build, nil
	case "images", "list", "l":
		return list, nil
	case "purge":
		return purge, nil
	case "", "--help", "help", "-h":
		return nil, errors.New("Available commands: install, list")
	default:
		return nil, fmt.Errorf("Unknown command: %s\n", name)
	}
}
