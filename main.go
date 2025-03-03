package main

import (
	"context"
	"fmt"
	"os"

	"github.com/billbatista/ha-daikin-smart-ac-br/cmd"
)

func main() {
	ctx := context.Background()

	if err := cmd.Server(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
