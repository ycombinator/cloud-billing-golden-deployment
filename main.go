package main

import (
	"fmt"
	"github.com/ycombinator/cloud-billing-golden-deployment/internal/cmd"
	"os"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}