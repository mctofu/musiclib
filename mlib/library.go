package mlib

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/dhowden/tag"
)

type Library struct {
	RootPaths    []string
	ArtistAlbums *ArtistAlbumIndex
}

func NewLibrary(ctx context.Context, rootPaths []string) (*Library, error) {
	artistAlbums := &ArtistAlbumIndex{}
	for _, rootPath := range rootPaths {
		artistAlbums.Index(ctx, rootPath)
	}

	return &Library{
		RootPaths:    rootPaths,
		ArtistAlbums: artistAlbums,
	}, nil
}

func (l *Library) Browse(ctx context.Context, browseURI string, opts BrowseOptions) ([]*BrowseItem, error) {
	index := l.ArtistAlbums

	if browseURI == "" {
		rootNodes, err := index.Roots(ctx)
		if err != nil {
			return nil, err
		}
		return filter(nil, rootNodes, opts.TextFilter)
	}

	node, err := index.Node(ctx, browseURI)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, nil
	}
	return filter(node, node.Children, opts.TextFilter)
}

func filter(parent *Node, nodes []*Node, filter string) ([]*BrowseItem, error) {
	var results []*BrowseItem

	parentMatches := parentMatch(parent, filter)

	for _, n := range nodes {
		if parentMatches || nodeMatch(n, filter) {
			results = append(results, toBrowseItem(n))
		}
	}

	return results, nil
}

func parentMatch(parent *Node, filter string) bool {
	if filter == "" {
		return true
	}

	for p := parent; p != nil; p = p.Parent {
		if strings.Contains(strings.ToLower(p.Name), filter) {
			return true
		}
	}

	return false
}

func nodeMatch(n *Node, filter string) bool {
	if filter == "" {
		return true
	}

	if strings.Contains(strings.ToLower(n.Name), filter) {
		return true
	}

	for _, child := range n.Children {
		if strings.Contains(strings.ToLower(child.Name), filter) {
			return true
		}
	}

	return false
}

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

func (a *ArtistAlbumIndex) Index(ctx context.Context, filePath string) error {
	if a.uriLookup == nil {
		a.uriLookup = make(map[string]*Node)
	}

	err := filepath.Walk(filePath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		f, err := os.Open(filePath)
		if err != nil {
			return err
		}

		m, err := tag.ReadFrom(f)
		if err != nil {
			if err == tag.ErrNoTagsFound {
				return nil
			}

			log.Printf("failed to read tag from %s: %v\n", filePath, err)
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
			song = path.Base(filePath)
		}
		songURI := encodeFileURI(filePath)
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
