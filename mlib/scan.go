package mlib

import (
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/dhowden/tag"
)

type WalkFunc func(dir *FileMeta, file *FileMeta) error

type Files struct {
	Roots []FileMeta
}

func (f *Files) Walk(walkFn WalkFunc) error {
	for _, root := range f.Roots {
		if err := root.WalkChildren(walkFn); err != nil {
			return err
		}
	}

	return nil
}

type FileMeta struct {
	Path     string
	TagMeta  tag.Metadata
	Children []FileMeta
}

func (f *FileMeta) IsDir() bool {
	return len(f.Children) > 0
}

func (f *FileMeta) WalkChildren(walkFn WalkFunc) error {
	for _, child := range f.Children {
		if child.IsDir() {
			if err := child.WalkChildren(walkFn); err != nil {
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
	var rootMetas []FileMeta
	for _, root := range roots {
		meta, err := scanDir(root)
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

func scanDir(dir string) (*FileMeta, error) {
	meta := &FileMeta{
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
			child, err := scanDir(path.Join(dir, file.Name()))
			if err != nil {
				return nil, err
			}
			if child != nil {
				meta.Children = append(meta.Children, *child)
			}

			continue
		}

		if !file.Mode().IsRegular() {
			continue
		}

		// try to read tags from media files
		filePath := path.Join(dir, file.Name())
		child, err := readFile(filePath)
		if err != nil {
			return nil, err
		}

		meta.Children = append(meta.Children, *child)
	}

	return meta, nil
}

func readFile(filePath string) (*FileMeta, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tagMeta, err := tag.ReadFrom(f)
	if err != nil && err != tag.ErrNoTagsFound {
		log.Printf("failed to read tag from %s: %v\n", filePath, err)
	}

	return &FileMeta{
		Path:    filePath,
		TagMeta: tagMeta,
	}, nil
}
