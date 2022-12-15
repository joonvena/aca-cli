package main

import (
	"fmt"
	"os"

	"github.com/joonvena/aca-cli/cmd"
)

func main() {
	if err := cmd.NewCmdRoot().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
