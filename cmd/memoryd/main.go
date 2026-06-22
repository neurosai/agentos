package main

import (
	"fmt"
	"os"

	"github.com/neurosai/agentos/pkg/version"
)

func main() {
	fmt.Printf("%s %s (memoryd placeholder)\n", version.Name, version.Version)
	os.Exit(0)
}
