package mlib

import "net/url"

type BrowseOptions struct {
	TextFilter string
}

type BrowseItem struct {
	Name   string
	URI    string
	Folder bool
}

type Node struct {
	Name     string
	URI      string
	Parent   *Node
	Children []*Node
}

func (n *Node) AddChildren(nodes ...*Node) {
	n.Children = append(n.Children, nodes...)
}

func toBrowseItems(nodes []*Node) []*BrowseItem {
	results := make([]*BrowseItem, 0, len(nodes))
	for _, n := range nodes {
		results = append(results, &BrowseItem{
			Name:   n.Name,
			URI:    n.URI,
			Folder: len(n.Children) > 0,
		})
	}

	return results
}

func toBrowseItem(n *Node) *BrowseItem {
	return &BrowseItem{
		Name:   n.Name,
		URI:    n.URI,
		Folder: len(n.Children) > 0,
	}
}

func encodeArtistURI(artist string) string {
	return "artist:///" + url.PathEscape(artist)
}

func encodeArtistAlbumURI(artist, album string) string {
	return "artistalbum:///" + url.PathEscape(artist) + "/" + url.PathEscape(album)
}

func encodeFileURI(filePath string) string {
	u := &url.URL{
		Scheme: "file",
		Path:   filePath,
	}
	return u.String()
}
