package musiclib

import (
	"context"
	"net/url"
	"strings"
)

type BrowseType string

const (
	BrowseTypeFile        BrowseType = "file"
	BrowseTypeAlbumArtist            = "albumartist"
	BrowseTypeGenre                  = "genre"
	BrowseTypeYear                   = "year"
	BrowseTypeModified               = "modified"
)

type BrowseOptions struct {
	TextFilter string
	BrowseType BrowseType
}

type BrowseItem struct {
	Name     string
	URI      string
	ImageURI string
	Folder   bool
}

type Index interface {
	Roots(ctx context.Context) ([]*Node, error)
	Node(ctx context.Context, uri string) (*Node, error)
}

type WalkNodeFunc func(n *Node) error

type Node struct {
	Name      string
	LowerName string
	URI       string
	ImageURI  string
	Parent    *Node
	Children  []*Node
}

func (n *Node) AddChildren(nodes ...*Node) {
	n.Children = append(n.Children, nodes...)
}

func (n *Node) IsFolder() bool {
	return len(n.Children) > 0
}

func (n *Node) walkLeaves(walkFn WalkNodeFunc) error {
	if len(n.Children) == 0 {
		return walkFn(n)
	}
	for _, child := range n.Children {
		if err := child.walkLeaves(walkFn); err != nil {
			return err
		}
	}

	return nil
}

func nameSort(nodes []*Node) func(i, j int) bool {
	return func(i, j int) bool {
		return nodes[i].LowerName < nodes[j].LowerName
	}
}

func toBrowseItem(n *Node) *BrowseItem {
	return &BrowseItem{
		Name:     n.Name,
		URI:      n.URI,
		ImageURI: n.ImageURI,
		Folder:   len(n.Children) > 0,
	}
}

func encodeCustomURI(scheme string, paths ...string) string {
	var sb strings.Builder
	sb.WriteString(scheme)
	sb.WriteString("://")
	for _, path := range paths {
		sb.WriteByte('/')
		sb.WriteString(url.PathEscape(path))
	}
	return sb.String()
}

func encodeFileURI(filePath string) string {
	if filePath == "" {
		return ""
	}

	u := &url.URL{
		Scheme: "file",
		Path:   filePath,
	}
	return u.String()
}
