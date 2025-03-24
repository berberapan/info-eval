package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/berberapan/info-eval/internal/vcs"
)

var version = vcs.Version()

func main() {

	versionFlag := flag.Bool("version", false, "Display version and exit")

	flag.Parse()

	if *versionFlag {
		fmt.Printf("Version: \t%s\n", version)
		os.Exit(0)
	}
}
