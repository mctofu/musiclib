package mlib

import (
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/dhowden/tag"
)

var tagExts = map[string]struct{}{
	".mp3":  {},
	".m4a":  {},
	".m4p":  {},
	".flac": {},
	".ogg":  {},
	".wav":  {},
}

type WalkFunc func(dir *PathMeta, file *PathMeta) error

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
	Name     string
	Path     string
	TagMeta  tag.Metadata
	Parent   *PathMeta
	Children []PathMeta
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
			if child != nil {
				child.Parent = meta
				meta.Children = append(meta.Children, *child)
			}

			continue
		}

		if !file.Mode().IsRegular() {
			continue
		}

		fileExt := path.Ext(file.Name())
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

	return &PathMeta{
		Path:    filePath,
		TagMeta: tagMeta,
	}, nil
}
