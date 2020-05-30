package mlib

import (
	"context"
	"log"
	"sort"
	"strings"
)

type AlbumArtistIndex struct {
	uriLookup map[string]*Node
	artists   []*Node
}

func (a *AlbumArtistIndex) Roots(ctx context.Context) ([]*Node, error) {
	return a.artists, nil
}

func (a *AlbumArtistIndex) Node(ctx context.Context, uri string) (*Node, error) {
	return a.uriLookup[uri], nil
}

func (a *AlbumArtistIndex) Index(ctx context.Context, files *Files) error {
	if a.uriLookup == nil {
		a.uriLookup = make(map[string]*Node)
	}

	err := files.WalkFiles(func(dir *PathMeta, file *PathMeta) error {
		m := file.Metadata

		if m == nil {
			log.Printf("No metadata for: %s\n", file.Path)
			return nil
		}

		artist := m.AlbumArtist()
		artistURI := encodeArtistURI(artist)
		artistNode, ok := a.uriLookup[artistURI]
		if !ok {
			artistNode = &Node{
				Name:      artist,
				LowerName: strings.ToLower(artist),
				URI:       artistURI,
			}
			a.artists = append(a.artists, artistNode)
			a.uriLookup[artistURI] = artistNode
		}

		album := m.Album()
		albumURI := encodeArtistAlbumURI(artist, album)
		albumNode, ok := a.uriLookup[albumURI]
		if !ok {
			albumNode = &Node{
				Name:      album,
				LowerName: strings.ToLower(album),
				URI:       albumURI,
				ImageURI:  encodeFileURI(dir.ImagePath),
				Parent:    artistNode,
			}
			artistNode.AddChildren(albumNode)
			a.uriLookup[albumURI] = albumNode
		}

		song := m.Song()
		songArtist := m.Artist()
		if songArtist != artist {
			song = songArtist + " - " + song
		}
		songURI := encodeFileURI(file.Path)
		songNode := &Node{
			Name:      song,
			LowerName: strings.ToLower(song),
			URI:       songURI,
			Parent:    albumNode,
		}
		albumNode.AddChildren(songNode)

		return nil
	})
	if err != nil {
		return err
	}

	sort.Slice(a.artists, nameSort(a.artists))
	for _, artist := range a.artists {
		sort.Slice(artist.Children, nameSort(artist.Children))
	}

	return nil
}
