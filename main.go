package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/orsinium-labs/dockerun/dockerun/dockecli"
	"github.com/spf13/pflag"
)

func main() {
	cmd, err := dockecli.GetCommand(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if len(os.Args) > 2 && os.Args[2] == "--help" {
		fmt.Println(cmd.Descr)
		os.Exit(0)
	}

	err = cmd.Run(os.Args[1:])
	if errors.Is(err, pflag.ErrHelp) {
		os.Exit(0)
	}
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}
