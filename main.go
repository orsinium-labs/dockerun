package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/orsinium-labs/dockerun/dockerun"
	"github.com/spf13/pflag"
)

func build(args []string) error {
	b, err := dockerun.NewBuilder(args)
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

func main() {
	var err error
	var cmd func([]string) error
	name := ""
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	switch name {
	case "build", "install", "i":
		cmd = build
	case "", "--help", "help", "-h":
		fmt.Println("Command required. Available commands: install")
		os.Exit(0)
	default:
		fmt.Printf("Unknown command: %s\n", name)
		os.Exit(1)
	}

	err = cmd(os.Args[1:])
	if errors.Is(err, pflag.ErrHelp) {
		os.Exit(0)
	}
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}
