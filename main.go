package main

import (
	"fmt"
	"os"

	"github.com/ycombinator/cloud-billing-golden-deployment/internal/server"
)

func main() {
	//if err := cmd.Execute(); err != nil {
	//	fmt.Fprintln(os.Stderr, err)
	//	os.Exit(1)
	//}

	if err := server.Start(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
