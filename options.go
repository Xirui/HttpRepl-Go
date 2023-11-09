package main

import (
	"fmt"
	"log"
	"os"

	"github.com/DavidGamba/go-getoptions"
)

func initOptions() string {
	var baseAddr string
	opt := getoptions.New()
	opt.Bool("help", false, opt.Alias("h", "?"))
	opt.StringVar(&baseAddr, "base-address", "config.yaml",
		opt.Required(), opt.Alias("b"),
		opt.Description("The initial base address for the REPL."))

	if opt.Called("help") {
		fmt.Fprint(os.Stderr, opt.Help())
		os.Exit(1)
	}
	remaining, err := opt.Parse(os.Args[1:])
	if opt.Called("help") {
		fmt.Fprint(os.Stderr, opt.Help())
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n\n", err)
		fmt.Fprint(os.Stderr, opt.Help(getoptions.HelpSynopsis))
		os.Exit(1)
	}
	if remaining != nil {
		log.Printf("Unhandled CLI args: %v\n", remaining)
	}
	return baseAddr
}
