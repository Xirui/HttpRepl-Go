package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
)

var (
	baseAddr, gLabel string
	gCurrentNode     *TreeNode // current working node
	openapiOps       = []string{"delete", "list", "get", "create", "update", "condition"}
)

func selectTest() []string {
	prompt := promptui.Prompt{
		Label: baseAddr + gLabel,
	}
	result, err := prompt.Run()
	if err == promptui.ErrInterrupt || err == promptui.ErrEOF {
		fmt.Println("⏎ ")
		os.Exit(0)
	}
	if result == "" {
		return nil
	}
	// fmt.Printf("Perform operation: %q\n", result)
	return strings.Split(result, " ")
}

func startupURL(opts argsOptions, root *TreeNode) {
	if opts.startURL == "" {
		return
	}
	url := opts.startURL
	if url[0] != '/' {
		url = "/" + url
	}
	gLabel = url
	path := strings.Split(gLabel, "/")
	for _, p := range path {
		if p == "" {
			continue
		}
		gCurrentNode = gCurrentNode.Children[p]
		if gCurrentNode == nil {
			gLabel = "/"
			gCurrentNode = root
			return
		}
	}
}

func main() {
	opts := initOptions()
	baseAddr = opts.baseAddr
	gLabel = "/"
	root := buildTree(baseAddr, opts.openapiPath)
	gCurrentNode = root
	startupURL(opts, root)

mainloop:
	for {
		result := selectTest()
		if result == nil {
			continue
		}
		switch result[0] {
		case "ls":
			lsImpl()
		case "cd":
			cdImpl(result, root)
		case "get":
			getImpl(result, root)
		case "tree":
			printTree(root, 0)
		case "exit":
			break mainloop
		default:
			defaultCommand()
		}
	}
}
