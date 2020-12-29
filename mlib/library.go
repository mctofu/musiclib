package mlib

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
)

type Library struct {
	RootPaths    []string
	AlbumArtists *MetadataIndex
	Files        *FileIndex
	Genres       *MetadataIndex
	Years        *MetadataIndex
	ModifyDates  *MetadataIndex
}

func NewLibrary(ctx context.Context, rootPaths []string) (*Library, error) {
	files, err := ScanRoots(rootPaths)
	if err != nil {
		return nil, err
	}
	log.Println("Scanned root paths")

	artistAlbums := NewArtistAlbumIndex()
	if err := artistAlbums.Index(ctx, files); err != nil {
		return nil, fmt.Errorf("failed to index artist/albums: %v", err)
	}
	log.Println("Indexed artist/album")

	filesIndex := &FileIndex{}
	if err := filesIndex.Index(ctx, files); err != nil {
		return nil, fmt.Errorf("failed to index files: %v", err)
	}
	log.Println("Indexed file paths")

	genreIndex := NewGenreIndex()
	if err := genreIndex.Index(ctx, files); err != nil {
		return nil, fmt.Errorf("failed to index genres: %v", err)
	}
	log.Println("Indexed genres")

	yearIndex := NewYearIndex()
	if err := yearIndex.Index(ctx, files); err != nil {
		return nil, fmt.Errorf("failed to index years: %v", err)
	}
	log.Println("Indexed years")

	modIndex := NewModifiedAtIndex()
	if err := modIndex.Index(ctx, files); err != nil {
		return nil, fmt.Errorf("failed to index modified dates: %v", err)
	}
	log.Println("Indexed modified dates")

	return &Library{
		RootPaths:    rootPaths,
		AlbumArtists: artistAlbums,
		Files:        filesIndex,
		Genres:       genreIndex,
		Years:        yearIndex,
		ModifyDates:  modIndex,
	}, nil
}

func (l *Library) Browse(ctx context.Context, browseURI string, opts BrowseOptions) ([]*BrowseItem, error) {
	index, err := l.index(opts.BrowseType)
	if err != nil {
		return nil, err
	}

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

func (l *Library) Media(ctx context.Context, uri string, opts BrowseOptions) ([]string, error) {
	index, err := l.index(opts.BrowseType)
	if err != nil {
		return nil, err
	}

	if uri == "" {
		return nil, errors.New("must specify a uri")
	}

	node, err := index.Node(ctx, uri)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, nil
	}

	if parentMatch(node.Parent, opts.TextFilter) {
		var uris []string
		if err := node.walkLeaves(func(n *Node) error {
			uris = append(uris, n.URI)
			return nil
		}); err != nil {
			return nil, err
		}
		return uris, nil
	}

	return filterLeaves(node, opts.TextFilter)
}

func (l *Library) index(t BrowseType) (Index, error) {
	switch t {
	case BrowseTypeFile:
		return l.Files, nil
	case BrowseTypeAlbumArtist:
		return l.AlbumArtists, nil
	case BrowseTypeGenre:
		return l.Genres, nil
	case BrowseTypeYear:
		return l.Years, nil
	case BrowseTypeModified:
		return l.ModifyDates, nil
	default:
		return nil, fmt.Errorf("unsupported browse type: %s", t)
	}
}

func filterLeaves(node *Node, filter string) ([]string, error) {
	var uris []string

	if nodeMatch(node, filter) {
		if err := node.walkLeaves(func(n *Node) error {
			uris = append(uris, n.URI)
			return nil
		}); err != nil {
			return nil, err
		}
	} else {
		for _, child := range node.Children {
			childURIs, err := filterLeaves(child, filter)
			if err != nil {
				return nil, err
			}
			uris = append(uris, childURIs...)
		}
	}
	return uris, nil
}

func filter(parent *Node, nodes []*Node, filter string) ([]*BrowseItem, error) {
	var results []*BrowseItem

	parentMatches := parentMatch(parent, filter)

	for _, n := range nodes {
		if parentMatches || nodeOrDescendantMatch(n, filter) {
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
		if strings.Contains(p.LowerName, filter) {
			return true
		}
	}

	return false
}

func nodeOrDescendantMatch(n *Node, filter string) bool {
	if nodeMatch(n, filter) {
		return true
	}

	for _, child := range n.Children {
		if nodeOrDescendantMatch(child, filter) {
			return true
		}
	}

	return false
}

func nodeMatch(n *Node, filter string) bool {
	if filter == "" {
		return true
	}

	if strings.Contains(n.LowerName, filter) {
		return true
	}

	return false
}
