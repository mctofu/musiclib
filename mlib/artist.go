package mlib

import (
	"context"
	"fmt"
	"log"
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

	err := files.WalkFiles(func(dir *PathMeta, file *PathMeta) error {
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
			song = file.Name
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
