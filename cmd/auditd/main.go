package main

import (
	"fmt"
	"os"

	"github.com/neurosai/agentos/pkg/version"
)

func main() {
	fmt.Printf("%s %s (auditd placeholder)\n", version.Name, version.Version)
	os.Exit(0)
}
