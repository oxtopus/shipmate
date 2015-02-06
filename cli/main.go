package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"shipmate"
)

func main() {
	remote := flag.String("remote", "", "Remote repository URL")
	name := flag.String("name", "", "Local destination of bare repository")
	rev := flag.String("rev", "master", "Git revision")
	userDefinedPrefix := flag.String("prefix", "", "Limit builds to paths with this prefix.  If not specified, process all.")

	flag.Parse()

	if len(*name) == 0 || len(*remote) == 0 {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	shipmate.run(*remote, *name, *rev, *userDefinedPrefix, cwd)
}
