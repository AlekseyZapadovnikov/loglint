package main

import (
	"fmt"
	"os"

	"github.com/AlekseyZapadovnikov/loglint/internal/analyzer"
	"github.com/AlekseyZapadovnikov/loglint/internal/config"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "loglint: get working directory: %v\n", err)
		os.Exit(1)
	}

	cfg, _, err := config.LoadStandaloneConfig(wd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "loglint: %v\n", err)
		os.Exit(1)
	}

	singlechecker.Main(analyzer.MustNew(cfg))
}
