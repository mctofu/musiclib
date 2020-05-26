package mlib

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"path"
)

type ArtistAlbumIndex struct {
	uriLookup map[string]*Node
	artists   []*Node
}

func (a *ArtistAlbumIndex) Roots(ctx context.Context) ([]*Node, error) {
	return a.artists, nil
}

func (a *ArtistAlbumIndex) Node(ctx context.Context, uri string) (*Node, error) {
	return a.uriLookup[uri], nil
}

func (a *ArtistAlbumIndex) Index(ctx context.Context, files *Files) error {
	if a.uriLookup == nil {
		a.uriLookup = make(map[string]*Node)
	}

	err := files.Walk(func(dir *FileMeta, file *FileMeta) error {
		m := file.TagMeta

		if m == nil {
			log.Printf("No tags for: %s\n", file.Path)
			return nil
		}

		artist := m.AlbumArtist()
		if artist == "" {
			alt := m.Raw()["album artist"]
			if alt != nil {
				artist = fmt.Sprintf("%s", alt)
			}
		}
		if artist == "" {
			artist = m.Artist()
		}
		if artist == "" {
			artist = "Unknown Artist"
		}
		artistURI := encodeArtistURI(artist)
		artistNode, ok := a.uriLookup[artistURI]
		if !ok {
			artistNode = &Node{
				Name: artist,
				URI:  artistURI,
			}
			a.artists = append(a.artists, artistNode)
			a.uriLookup[artistURI] = artistNode
		}

		album := m.Album()
		if album == "" {
			album = "Unknown Album"
		}
		albumURI := encodeArtistAlbumURI(artist, album)
		albumNode, ok := a.uriLookup[albumURI]
		if !ok {
			albumNode = &Node{
				Name:   album,
				URI:    albumURI,
				Parent: artistNode,
			}
			artistNode.AddChildren(albumNode)
			a.uriLookup[albumURI] = albumNode
		}

		song := m.Title()
		if song == "" {
			song = path.Base(file.Path)
		}
		songURI := encodeFileURI(file.Path)
		songNode := &Node{
			Name:   song,
			URI:    songURI,
			Parent: albumNode,
		}
		albumNode.AddChildren(songNode)

		return nil
	})

	return err
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

type Node struct {
	Name     string
	URI      string
	Parent   *Node
	Children []*Node
}

func (n *Node) AddChildren(nodes ...*Node) {
	n.Children = append(n.Children, nodes...)
}
