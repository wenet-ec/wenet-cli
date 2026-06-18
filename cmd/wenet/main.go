// cmd/wenet/main.go
package main

import (
	"fmt"
	"os"

	"github.com/wenet-ec/wenet-cli/internal/cli"
)

func main() {
	if err := cli.NewRootCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
