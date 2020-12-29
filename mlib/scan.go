package mlib

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	"github.com/dhowden/tag"
)

var tagExts = map[string]struct{}{
	".mp3":  {},
	".m4a":  {},
	".flac": {},
	".ogg":  {},
	".wav":  {},
}

var imgExts = map[string]struct{}{
	".jpg": {},
	".png": {},
	".gif": {},
}

type WalkFunc func(dir *PathMeta, file *PathMeta) error

type MediaMetadata interface {
	Artist() string
	AlbumArtist() string
	Album() string
	Song() string
	AlbumArtURI() string
	Track() int
	Genre() string
	Modified() time.Time
	Year() int
}

type Files struct {
	Roots []PathMeta
}

func (f *Files) WalkFiles(walkFn WalkFunc) error {
	for _, root := range f.Roots {
		if err := root.walkChildren(walkFn); err != nil {
			return err
		}
	}

	return nil
}

type PathMeta struct {
	Name      string
	Path      string
	ImagePath string
	Metadata  MediaMetadata
	Parent    *PathMeta
	Children  []PathMeta
}

func (f *PathMeta) IsDir() bool {
	return len(f.Children) > 0
}

func (f *PathMeta) walkChildren(walkFn WalkFunc) error {
	for _, child := range f.Children {
		if child.IsDir() {
			if err := child.walkChildren(walkFn); err != nil {
				return err
			}
			continue
		}
		if err := walkFn(f, &child); err != nil {
			return err
		}
	}

	return nil
}

func ScanRoots(roots []string) (*Files, error) {
	var rootMetas []PathMeta
	for _, root := range roots {
		meta, err := scanDir(path.Base(root), root)
		if err != nil {
			return nil, err
		}
		if meta != nil {
			rootMetas = append(rootMetas, *meta)
		}
	}

	return &Files{
		Roots: rootMetas,
	}, nil
}

func scanDir(name string, dir string) (*PathMeta, error) {
	meta := &PathMeta{
		Name: name,
		Path: dir,
	}

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, nil
	}

	for _, file := range files {
		if file.IsDir() {
			child, err := scanDir(file.Name(), path.Join(dir, file.Name()))
			if err != nil {
				return nil, err
			}
			if child != nil && len(child.Children) > 0 {
				child.Parent = meta
				meta.Children = append(meta.Children, *child)
			}

			continue
		}

		if !file.Mode().IsRegular() {
			continue
		}

		fileExt := path.Ext(file.Name())
		if _, ok := imgExts[fileExt]; ok {
			meta.ImagePath = path.Join(dir, file.Name())
			continue
		}
		if _, ok := tagExts[fileExt]; !ok {
			continue
		}

		// try to read tags from media files
		filePath := path.Join(dir, file.Name())
		child, err := readFile(filePath)
		if err != nil {
			return nil, err
		}
		child.Name = file.Name()

		child.Parent = meta
		meta.Children = append(meta.Children, *child)
	}

	return meta, nil
}

func readFile(filePath string) (*PathMeta, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tagMeta, err := tag.ReadFrom(f)
	if err != nil && err != tag.ErrNoTagsFound {
		log.Printf("failed to read tag from %s: %v\n", filePath, err)
	}

	info, err := f.Stat()
	if err != nil {
		log.Printf("failed to stat %s: %v\n", filePath, err)
	}

	meta := &PathMeta{
		Path: filePath,
	}
	meta.Metadata = &mediaMetadataReader{
		tagData: tagMeta,
		file:    meta,
		info:    info,
	}

	return meta, nil
}

const (
	unknownArtist = "Unknown Artist"
	unknownAlbum  = "Unknown Album"
	unknownGenre  = "Unknown Genre"
)

type mediaMetadataReader struct {
	tagData tag.Metadata
	file    *PathMeta
	info    os.FileInfo
}

func (m *mediaMetadataReader) Artist() string {
	if m.tagData == nil {
		return unknownArtist
	}
	artist := m.tagData.Artist()
	if artist != "" {
		return artist
	}
	return unknownArtist
}

func (m *mediaMetadataReader) AlbumArtist() string {
	if m.tagData == nil {
		return unknownArtist
	}
	artist := m.tagData.AlbumArtist()
	if artist != "" {
		return artist
	}
	alt := m.tagData.Raw()["album artist"]
	if alt != nil {
		artist = fmt.Sprintf("%s", alt)
	}
	if artist != "" {
		return artist
	}

	return m.Artist()
}

func (m *mediaMetadataReader) Album() string {
	if m.tagData == nil {
		return unknownAlbum
	}
	album := m.tagData.Album()
	if album != "" {
		return album
	}
	return unknownAlbum
}

func (m *mediaMetadataReader) Song() string {
	if m.tagData == nil {
		return m.file.Name
	}
	song := m.tagData.Title()
	if song != "" {
		return song
	}
	return m.file.Name
}

func (m *mediaMetadataReader) AlbumArtURI() string {
	return ""
}

func (m *mediaMetadataReader) Track() int {
	if m.tagData == nil {
		return 0
	}
	t, _ := m.tagData.Track()
	return t
}

func (m *mediaMetadataReader) Genre() string {
	if m.tagData == nil {
		return unknownGenre
	}
	genre := m.tagData.Genre()
	if genre != "" {
		return genre
	}
	return unknownGenre
}

func (m *mediaMetadataReader) Year() int {
	if m.tagData == nil {
		return 0
	}
	return m.tagData.Year()
}

func (m *mediaMetadataReader) Modified() time.Time {
	if m.info == nil {
		return time.Time{}
	}
	return m.info.ModTime()
}
