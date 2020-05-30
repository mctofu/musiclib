package mlib

import (
	"context"
	"net/url"
)

type BrowseType string

const (
	BrowseTypeFile        BrowseType = "file"
	BrowseTypeAlbumArtist BrowseType = "albumartist"
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

func encodeArtistURI(artist string) string {
	return "artist:///" + url.PathEscape(artist)
}

func encodeArtistAlbumURI(artist, album string) string {
	return "artistalbum:///" + url.PathEscape(artist) + "/" + url.PathEscape(album)
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
