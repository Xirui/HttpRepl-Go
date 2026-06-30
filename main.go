package main

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
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
	args, err := splitArgs(result)
	if err != nil {
		fmt.Printf("\x1b[31mError parsing command: %v\x1b[0m\n", err)
		return nil
	}
	return args
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

	// Initialize Cookie Jar
	jar, err := cookiejar.New(nil)
	if err != nil {
		fmt.Printf("Warning: Failed to initialize cookie jar: %v\n", err)
	} else {
		http.DefaultClient.Jar = jar
	}

mainloop:
	for {
		result := selectTest()
		if result == nil {
			continue
		}
		switch strings.ToLower(result[0]) {
		case "ls":
			lsImpl()
		case "cd":
			cdImpl(result, root)
		case "get", "post", "put", "delete", "patch", "head", "options":
			makeRequest(strings.ToUpper(result[0]), result)
		case "set":
			if len(result) >= 2 && result[1] == "header" {
				handleHeaderCommand(result)
			} else if len(result) >= 2 && result[1] == "cookie" {
				handleCookieCommand(result)
			} else {
				fmt.Println("Unknown set command. Use 'set header <name> <value>' or 'set cookie <name> <value>'")
			}
		case "clear":
			if len(result) >= 2 && result[1] == "header" {
				handleClearCommand(result)
			} else if len(result) >= 2 && result[1] == "cookie" {
				handleClearCookieCommand(result)
			} else {
				fmt.Println("Unknown clear command. Use 'clear header <name>' or 'clear cookie <name>'")
			}
		case "show":
			if len(result) >= 2 && result[1] == "cookies" {
				showCookiesImpl()
			} else {
				fmt.Println("Unknown show command. Use 'show cookies'")
			}
		case "tree":
			printTree(root, 0)
		case "help":
			helpImpl()
		case "exit":
			break mainloop
		default:
			defaultCommand()
		}
	}
}
