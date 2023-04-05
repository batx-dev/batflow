package main

import (
	"fmt"
	"os"

	"github.com/batx-dev/batflow/internal/cmd"
)

func main() {
	if err := cmd.GetRootCommand().Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
