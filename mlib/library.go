package mlib

import (
	"context"
	"fmt"
	"strings"
)

type Library struct {
	RootPaths    []string
	ArtistAlbums *ArtistAlbumIndex
	Files        *FileIndex
}

func NewLibrary(ctx context.Context, rootPaths []string) (*Library, error) {
	files, err := ScanRoots(rootPaths)
	if err != nil {
		return nil, err
	}

	artistAlbums := &ArtistAlbumIndex{}
	if err := artistAlbums.Index(ctx, files); err != nil {
		return nil, fmt.Errorf("failed to index artist/albums: %v", err)
	}

	filesIndex := &FileIndex{}
	if err := filesIndex.Index(ctx, files); err != nil {
		return nil, fmt.Errorf("failed to index files: %v", err)
	}

	return &Library{
		RootPaths:    rootPaths,
		ArtistAlbums: artistAlbums,
		Files:        filesIndex,
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

func (l *Library) index(t BrowseType) (Index, error) {
	switch t {
	case BrowseTypeFile:
		return l.Files, nil
	case BrowseTypeAlbumArtist:
		return l.ArtistAlbums, nil
	default:
		return nil, fmt.Errorf("unsupported browse type: %s", t)
	}
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
		if strings.Contains(p.LowerName, filter) {
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

	for _, child := range n.Children {
		if nodeMatch(child, filter) {
			return true
		}
	}

	return false
}
