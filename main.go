package main

import (
	"fmt"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/manifoldco/promptui"
)

var subdir = []string{
	"alert [POST]",
	"alertEmail [POST]",
	"customer [POST]",
}

var (
	baseAddr, gLabel string // http://localhost:8080/api/v1
	openapiOps       = []string{"delete", "list", "get", "create", "update", "conditions"}
)

func selectTest() []string {
	prompt := promptui.Prompt{
		Label: gLabel,
	}
	result, _ := prompt.Run()
	if result == "" {
		return nil
	}
	// fmt.Printf("Perform operation: %q\n", result)
	return strings.Split(result, " ")
}

func changeDir(args []string) {
	if len(args) == 1 {
		gLabel = baseAddr
		return
	}
	// assuming it is a valid endpoint
	gLabel = gLabel + "/" + args[1]
}

func main() {
	baseAddr = initOptions()
	gLabel = baseAddr
	buildTree()
mainloop:
	for {
		result := selectTest()
		if result == nil {
			continue
		}
		switch result[0] {
		case "ls":
			for _, d := range subdir {
				fmt.Println(d)
			}
		case "cd":
			changeDir(result)
		case "exit":
			break mainloop
		default:
			spew.Dump(result)
		}
	}

}
