package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/davecgh/go-spew/spew"
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
	result, _ := prompt.Run()
	if result == "" {
		return nil
	}
	// fmt.Printf("Perform operation: %q\n", result)
	return strings.Split(result, " ")
}

func lsImpl() {
	fmt.Println(".")
	if gCurrentNode.Name != baseAddr {
		fmt.Println("..")
	}
	for _, d := range gCurrentNode.Children {
		fmt.Println(d.Name)
	}
	fmt.Println("")
}

func cdImpl(args []string, root *TreeNode) {
	if len(args) == 1 { // no argument -> go to root
		gLabel = "/"
		gCurrentNode = root
		return
	}
	dest := args[1]
	if dest == "." {
		return
	}
	if dest == ".." {
		if gCurrentNode.Parent == nil {
			return
		}
		gLabel = gLabel[:strings.LastIndex(gLabel, "/")]
		gCurrentNode = gCurrentNode.Parent
		if gLabel == "" { // add / when cd to root
			gLabel = "/"
		}
		return
	}
	label := gLabel
	if label[len(label)-1] != '/' {
		label += "/"
	}
	label += dest
	for _, d := range gCurrentNode.Children {
		if d.Name == dest {
			gLabel = label
			gCurrentNode = d
			return
		}
	}
	msg := fmt.Sprintf("Warning: The '%v' endpoint is not present in the OpenAPI description", label)
	fmt.Println("\x1b[33m" + msg + ".\x1b[0m\n")
}

func getImpl(args []string, root *TreeNode) {
	if len(args) != 2 {
		fmt.Println("\x1b[31mError: Invalid number of arguments.\x1b[0m\n")
		return
	}
	// {"method": "GET", "url": "/api/v1/alert/49", "request_id": "4RD3hf8qtg"}
	query := fmt.Sprintf("%s%s/%v", baseAddr, gLabel, args[1])
	fmt.Println(query)
	// logger.Infoln(query)
	if resp, err := http.Get(query); err != nil {
		fmt.Println("\x1b[31mError: Failed to get response.\x1b[0m\n")
	} else {
		spew.Dump(resp.Body)
	}

}

func defaultCommand() {
	fmt.Println("\x1b[31mNo matching command found.\x1b[0m")
	fmt.Println("\x1b[31mExecute 'help' to see available commands.\x1b[0m\n")
}

func main() {
	opts := initOptions()
	baseAddr = opts.baseAddr
	gLabel = "/"
	root := buildTree(baseAddr, opts.openapiPath)
	gCurrentNode = root
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
