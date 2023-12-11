package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/davecgh/go-spew/spew"
)

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

func prettyJSON(body []byte) {
	// Create an empty interface to store the unmarshalled JSON data
	var parsedJSON interface{}

	// Unmarshal the JSON data into the interface
	err := json.Unmarshal(body, &parsedJSON)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Pretty print the JSON using MarshalIndent
	prettyJSON, err := json.MarshalIndent(parsedJSON, "", "  ")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Print the pretty printed JSON as a string
	fmt.Println(string(prettyJSON))
}

func getImpl(args []string, root *TreeNode) {
	if len(args) != 2 {
		fmt.Println("\x1b[31mError: Invalid number of arguments.\x1b[0m\n")
		return
	}
	query := fmt.Sprintf("%s%s/%v", baseAddr, gLabel, args[1])
	fmt.Println(query)
	resp, err := http.Get(query)
	if err != nil {
		fmt.Println("\x1b[31mError: Failed to get response.\x1b[0m\n")
	} else {
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response:", err)
			return
		}
		fmt.Println(resp.Status)
		spew.Dump(resp.Header)
		prettyJSON(body)
	}
}

func defaultCommand() {
	spew.Dump(gCurrentNode)
	fmt.Println("\x1b[31mNo matching command found.\x1b[0m")
	fmt.Println("\x1b[31mExecute 'help' to see available commands.\x1b[0m\n")
}
