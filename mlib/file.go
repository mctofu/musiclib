package mlib

import (
	"context"
)

type FileIndex struct {
	uriLookup map[string]*Node
	roots     []*Node
}

func (f *FileIndex) Roots(ctx context.Context) ([]*Node, error) {
	return f.roots, nil
}

func (f *FileIndex) Node(ctx context.Context, uri string) (*Node, error) {
	return f.uriLookup[uri], nil
}

func (f *FileIndex) Index(ctx context.Context, files *Files) error {
	if f.uriLookup == nil {
		f.uriLookup = make(map[string]*Node)
	}

	for _, root := range files.Roots {
		rootNode := f.addNode(nil, &root)
		f.roots = append(f.roots, rootNode)
	}

	return nil
}

func (f *FileIndex) addNode(parent *Node, filePath *PathMeta) *Node {
	node := &Node{
		Name:   filePath.Name,
		URI:    encodeFileURI(filePath.Path),
		Parent: parent,
	}

	for _, child := range filePath.Children {
		f.addNode(node, &child)
	}

	if parent != nil {
		parent.Children = append(parent.Children, node)
	}

	f.uriLookup[node.URI] = node

	return node
}
