package main

import (
	"os"

	"github.com/ghulammuzz/clauzz-cli/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
