package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

type treeNode struct {
	file     string
	children []treeNode
}

func getNodes(path string, withFiles bool) ([]treeNode, error) {
	nodes := []treeNode{}

	files, err := os.ReadDir(path)
	// sort.Slice(files, func(i, j int) bool {
	// 	return files[i].Name() < files[j].Name()
	// })
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if !withFiles && !file.IsDir() {
			continue
		}

		fi, err := file.Info()
		if err != nil {
			return nil, err
		}

		newNode := treeNode{}

		if file.IsDir() {
			children, err := getNodes(path+"/"+file.Name(), withFiles)
			if err != nil {
				return nil, err
			}
			newNode.file = fi.Name()
			newNode.children = children
		} else {
			size := fi.Size()
			var sizeStr string
			if size == 0 {
				sizeStr = " (empty)"
			} else {
				sizeStr = fmt.Sprintf(" (%db)", size)
			}
			newNode.file = fi.Name() + sizeStr
		}

		nodes = append(nodes, newNode)

	}

	return nodes, nil
}

func printTree(out io.Writer, tree []treeNode, parentPrefix string) {
	var (
		lastIndex   = len(tree) - 1
		prefix      = "├───"
		childPrefix = "│	"
	)

	for i, node := range tree {
		if i == lastIndex {
			prefix = "└───"
			childPrefix = "	"
		}

		f := node.file
		fmt.Fprint(out, parentPrefix, prefix, f, "\n")
		if node.children != nil && len(node.children) != 0 {
			printTree(out, node.children, parentPrefix+childPrefix)
		}
	}
}

func dirTree(out io.Writer, path string, printFiles bool) error {
	tree, err := getNodes(path, printFiles)
	if err != nil {
		return err
	}

	printTree(out, tree, "")

	return nil
}
