package main

import (
	"fmt"
	"log"
	"os"

	"github.com/DavidGamba/go-getoptions"
)

type argsOptions struct {
	baseAddr    string
	openapiPath string
	startURL    string
}

func initOptions() argsOptions {
	var opts argsOptions
	opt := getoptions.New()
	opt.Bool("help", false, opt.Alias("h", "?"))
	opt.StringVar(&opts.baseAddr, "base-address", "http://localhost:8080",
		opt.Required(), opt.Alias("b"),
		opt.Description("The initial base address and port for the REPL."))

	opt.StringVar(&opts.openapiPath, "openapi", "/swagger/doc.json",
		opt.Required(), opt.Alias("o"),
		opt.Description("OpenAPI description path."))

	opt.StringVar(&opts.startURL, "start-url", "/api/v1",
		opt.Alias("u"), opt.Description("Start URL."))

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
	fmt.Println("Using a base address of", opts.baseAddr)
	return opts
}
