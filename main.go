package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/orsinium-labs/dockerun/dockerun/dockecli"
	"github.com/spf13/pflag"
)

func main() {
	cmd, err := dockecli.GetCommand()
	if err != nil {
		fmt.Println(err)
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
