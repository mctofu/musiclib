package musiclib

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
)

type NodeBuilder func(lookup map[string]*Node, dir *PathMeta, file *PathMeta, uriPaths []string) (*Node, []string, bool)

type MetadataIndex struct {
	uriLookup map[string]*Node
	builders  []NodeBuilder
	roots     []*Node
}

func NewMetadataIndex(builders []NodeBuilder) *MetadataIndex {
	return &MetadataIndex{
		uriLookup: make(map[string]*Node),
		builders:  builders,
	}
}

func (a *MetadataIndex) Roots(ctx context.Context) ([]*Node, error) {
	return a.roots, nil
}

func (a *MetadataIndex) Node(ctx context.Context, uri string) (*Node, error) {
	return a.uriLookup[uri], nil
}

func (a *MetadataIndex) Index(ctx context.Context, files *Files) error {
	if err := files.WalkFiles(func(dir *PathMeta, file *PathMeta) error {
		if err := ctx.Err(); err != nil {
			return err
		}

		if file.Metadata == nil {
			log.Printf("No metadata for: %s\n", file.Path)
			return nil
		}

		uriPaths := make([]string, 0, len(a.builders))

		var parent *Node
		for _, builder := range a.builders {
			node, newPaths, added := builder(a.uriLookup, dir, file, uriPaths)
			uriPaths = newPaths
			if added {
				a.uriLookup[node.URI] = node
				node.LowerName = strings.ToLower(node.Name)
				if parent == nil {
					a.roots = append(a.roots, node)
				} else {
					parent.AddChildren(node)
					node.Parent = parent
				}
			}
			parent = node
		}

		return nil
	}); err != nil {
		return err
	}

	sort.Slice(a.roots, nameSort(a.roots))
	for _, root := range a.roots {
		sortChildren(root)
	}

	return nil
}

func sortChildren(node *Node) {
	if !node.IsFolder() {
		return
	}
	if !node.Children[0].IsFolder() {
		return
	}
	sort.Slice(node.Children, nameSort(node.Children))
	for _, child := range node.Children {
		sortChildren(child)
	}
}

func NewGenreIndex() *MetadataIndex {
	return NewMetadataIndex(
		[]NodeBuilder{
			genreNode,
			artistNodeBuilder("genreartist"),
			albumNodeBuilder("genrealbum"),
			songNode,
		})
}

func genreNode(lookup map[string]*Node, dir *PathMeta, file *PathMeta, uriPaths []string) (*Node, []string, bool) {
	genre := file.Metadata.Genre()
	uriPaths = append(uriPaths, genre)
	genreURI := encodeCustomURI("genre", uriPaths...)
	genreNode, ok := lookup[genreURI]
	if !ok {
		genreNode = &Node{
			Name: genre,
			URI:  genreURI,
		}
	}

	return genreNode, uriPaths, !ok
}

func NewModifiedAtIndex() *MetadataIndex {
	return NewMetadataIndex(
		[]NodeBuilder{
			modifiedYearNode,
			modifiedMonthNode,
			artistNodeBuilder("modartist"),
			albumNodeBuilder("modalbum"),
			songNode,
		})
}

func modifiedYearNode(lookup map[string]*Node, dir *PathMeta, file *PathMeta, uriPaths []string) (*Node, []string, bool) {
	year := strconv.Itoa(file.Metadata.Modified().Year())
	uriPaths = append(uriPaths, year)
	modYearURI := encodeCustomURI("modyear", uriPaths...)
	yearNode, ok := lookup[modYearURI]
	if !ok {
		yearNode = &Node{
			Name: year,
			URI:  modYearURI,
		}
	}

	return yearNode, uriPaths, !ok
}

func modifiedMonthNode(lookup map[string]*Node, dir *PathMeta, file *PathMeta, uriPaths []string) (*Node, []string, bool) {
	month := fmt.Sprintf("%02d", file.Metadata.Modified().Month())
	uriPaths = append(uriPaths, month)
	modMonthURI := encodeCustomURI("modmonth", uriPaths...)
	monthNode, ok := lookup[modMonthURI]
	if !ok {
		monthNode = &Node{
			Name: month,
			URI:  modMonthURI,
		}
	}

	return monthNode, uriPaths, !ok
}

func NewYearIndex() *MetadataIndex {
	return NewMetadataIndex(
		[]NodeBuilder{
			yearNode,
			artistNodeBuilder("modartist"),
			albumNodeBuilder("modalbum"),
			songNode,
		})
}

func yearNode(lookup map[string]*Node, dir *PathMeta, file *PathMeta, uriPaths []string) (*Node, []string, bool) {
	year := strconv.Itoa(file.Metadata.Year())
	if year == "0" {
		year = "Unknown year"
	}
	uriPaths = append(uriPaths, year)
	yearURI := encodeCustomURI("year", uriPaths...)
	yearNode, ok := lookup[yearURI]
	if !ok {
		yearNode = &Node{
			Name: year,
			URI:  yearURI,
		}
	}

	return yearNode, uriPaths, !ok
}

func NewArtistAlbumIndex() *MetadataIndex {
	return NewMetadataIndex(
		[]NodeBuilder{
			artistNodeBuilder("artist"),
			albumNodeBuilder("artistalbum"),
			songNode,
		})
}

func artistNodeBuilder(scheme string) NodeBuilder {
	return func(lookup map[string]*Node, dir *PathMeta, file *PathMeta, uriPaths []string) (*Node, []string, bool) {
		artist := file.Metadata.AlbumArtist()
		uriPaths = append(uriPaths, artist)
		artistURI := encodeCustomURI(scheme, uriPaths...)
		artistNode, ok := lookup[artistURI]
		if !ok {
			artistNode = &Node{
				Name:      artist,
				LowerName: strings.ToLower(artist),
				URI:       artistURI,
			}
		}
		return artistNode, uriPaths, !ok
	}
}

func albumNodeBuilder(scheme string) NodeBuilder {
	return func(lookup map[string]*Node, dir *PathMeta, file *PathMeta, uriPaths []string) (*Node, []string, bool) {
		album := file.Metadata.Album()
		uriPaths = append(uriPaths, album)
		albumURI := encodeCustomURI(scheme, uriPaths...)
		albumNode, ok := lookup[albumURI]
		if !ok {
			albumNode = &Node{
				Name:     album,
				URI:      albumURI,
				ImageURI: encodeFileURI(dir.ImagePath),
			}
		}

		return albumNode, uriPaths, !ok
	}
}

func songNode(lookup map[string]*Node, dir *PathMeta, file *PathMeta, uriPaths []string) (*Node, []string, bool) {
	artist := file.Metadata.AlbumArtist()
	song := file.Metadata.Song()
	songArtist := file.Metadata.Artist()
	if songArtist != artist {
		song = songArtist + " - " + song
	}
	songURI := encodeFileURI(file.Path)
	songNode := &Node{
		Name: song,
		URI:  songURI,
	}

	return songNode, uriPaths, true
}
