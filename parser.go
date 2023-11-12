package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/samber/lo"
)

// SwaggerDoc represents the top-level structure of the Swagger JSON.
type SwaggerDoc struct {
	Paths map[string]map[string]Endpoint
}

// Endpoint represents an API endpoint in the Swagger JSON.
type Endpoint struct {
	Summary string `json:"summary"`
	// Add other fields as needed
}

// TreeNode represents a node in the tree structure.
type TreeNode struct {
	Name     string
	Endpoint *Endpoint
	Children map[string]*TreeNode
}

func buildTree() {
	// Read and parse the Swagger JSON document.
	swaggerData, err := os.ReadFile("doc.json")
	if err != nil {
		fmt.Println("Error reading Swagger JSON:", err)
		os.Exit(1)
	}

	var swagger SwaggerDoc
	if err := json.Unmarshal(swaggerData, &swagger); err != nil {
		fmt.Println("Error parsing Swagger JSON:", err)
		os.Exit(1)
	}

	// Create the root node for the tree.
	root := &TreeNode{
		Name:     "Root",
		Endpoint: nil,
		Children: make(map[string]*TreeNode),
	}
	// Build the tree structure.
	for path, methods := range swagger.Paths {
		fmt.Println(path, methods)
		pathNames := strings.Split(path, "/")
		fmt.Println(pathNames)
		parent := root
		for _, name := range pathNames { // find the parent node
			if lo.Contains(openapiOps, name) {
				break
			}
			child, ok := parent.Children[name]
			if !ok {
				AddNode(parent, name, nil)
				parent = parent.Children[name]
			} else {
				parent = child
			}
		}
		for method, endpoint := range methods {
			// Customize the node name as needed.
			nodeName := fmt.Sprintf("%s %s", method, path)
			AddNode(root, nodeName, &endpoint)
		}
	}

	// Print the tree structure.
	printTree(root, 0)
}

// AddNode adds a node to the tree structure.
func AddNode(parent *TreeNode, name string, endpoint *Endpoint) {
	node := &TreeNode{
		Name:     name,
		Endpoint: endpoint,
		Children: make(map[string]*TreeNode),
	}

	parent.Children[name] = node
}

// printTree recursively prints the tree structure.
func printTree(node *TreeNode, depth int) {
	if node == nil {
		return
	}

	fmt.Printf("%s- %s\n", getIndent(depth), node.Name)

	for _, child := range node.Children {
		printTree(child, depth+1)
	}
}

// getIndent generates indentation for tree structure printing.
func getIndent(depth int) string {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}
	return indent
}
