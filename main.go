package main

import (
	"fmt"
	"os"

	"github.com/orsinium-labs/dockerun/dockerun/dockecli"
)

func main() {
	err := dockecli.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
