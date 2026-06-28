package main

import (
	"encoding/json"
	"fmt"
	"net/http"
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
}

// TreeNode represents a node in the tree structure.
type TreeNode struct {
	Name     string
	Methods  map[string]*Endpoint
	Children map[string]*TreeNode
	Parent   *TreeNode
}

func getParent(root *TreeNode, pathNames []string) *TreeNode {
	parent := root
	for _, name := range pathNames { // find the parent node
		if name == "" {
			continue
		}
		if lo.Contains(openapiOps, name) {
			break
		}
		child, ok := parent.Children[name]
		if !ok {
			parent = AddNode(parent, name)
		} else {
			parent = child
		}
	}
	return parent
}

// buildTree is a function that builds a tree structure based on a given base address and OpenAPI path.
func buildTree(baseAddr string, openapiPath string) *TreeNode {
	fmt.Printf("Checking %v%v... ", baseAddr, openapiPath)
	resp, err := http.Get(baseAddr + openapiPath)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	fmt.Printf("\x1b[32mFound\x1b[0m\n")
	fmt.Printf("Parsing... ")
	var swagger SwaggerDoc
	if err := json.NewDecoder(resp.Body).Decode(&swagger); err != nil {
		panic(err)
	}

	root := &TreeNode{
		Name:     baseAddr,
		Methods:  make(map[string]*Endpoint),
		Parent:   nil,
		Children: make(map[string]*TreeNode),
	}
	// Build the tree structure.
	for path, methods := range swagger.Paths {
		pathNames := strings.Split(path, "/")
		parent := getParent(root, pathNames)
		for method, endpoint := range methods {
			parent.Methods[strings.ToUpper(method)] = &endpoint
		}
	}
	fmt.Printf("\x1b[32mSuccessful\x1b[0m\n")
	printTree(root, 0)
	return root
}

// AddNode adds a node to the tree structure.
func AddNode(parent *TreeNode, name string) *TreeNode {
	node := &TreeNode{
		Name:     name,
		Methods:  make(map[string]*Endpoint),
		Children: make(map[string]*TreeNode),
		Parent:   parent,
	}

	parent.Children[name] = node
	return node
}

// printTree recursively prints the tree structure.
func printTree(node *TreeNode, depth int) {
	if node == nil {
		return
	}

	indent := getIndent(depth)
	var methodList []string
	for m := range node.Methods {
		methodList = append(methodList, strings.ToLower(m))
	}
	
	if len(methodList) > 0 {
		fmt.Printf("%s- %s [%s]\n", indent, node.Name, strings.Join(methodList, "|"))
	} else {
		fmt.Printf("%s- %s\n", indent, node.Name)
	}

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
