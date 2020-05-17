package filesys

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/mctofu/music-library-grpc/go/mlibgrpc"
	"github.com/mctofu/musiclib/mlib"
)

type Library struct {
	RootPaths []string
}

func (l *Library) Browse(ctx context.Context, browsePath string, opts mlib.BrowseOptions) ([]*mlibgrpc.BrowseItem, error) {
	if browsePath == "" {
		return l.rootBrowse(opts.TextFilter)
	}

	files, err := ioutil.ReadDir(browsePath)
	if err != nil {
		return nil, err
	}

	parentMatch := l.parentMatch(opts.TextFilter, browsePath)

	items := make([]*mlibgrpc.BrowseItem, 0, len(files))
	for _, file := range files {
		if !parentMatch {
			fullPath := path.Join(browsePath, file.Name())
			match, err := fileMatch(opts.TextFilter, file, fullPath)
			if err != nil {
				return nil, err
			}
			if !match {
				continue
			}
		}

		items = append(items, &mlibgrpc.BrowseItem{
			Name:   file.Name(),
			Folder: file.IsDir(),
			Uri:    path.Join(browsePath, file.Name()),
		})
	}

	return items, nil
}

func (l *Library) rootBrowse(filter string) ([]*mlibgrpc.BrowseItem, error) {
	items := make([]*mlibgrpc.BrowseItem, 0, len(l.RootPaths))
	for _, rootPath := range l.RootPaths {
		file, err := os.Stat(rootPath)
		if err != nil {
			return nil, fmt.Errorf("invalid root path: %v", err)
		}
		match, err := fileMatch(filter, file, rootPath)
		if err != nil {
			return nil, err
		}
		if !match {
			continue
		}

		items = append(items, &mlibgrpc.BrowseItem{
			Name:   file.Name(),
			Folder: true,
			Uri:    rootPath,
		})
	}

	return items, nil
}

func (l *Library) parentMatch(filter string, browsePath string) bool {
	if filter == "" {
		return true
	}
	for _, rootPath := range l.RootPaths {
		if strings.HasPrefix(browsePath, rootPath) {
			rootName := strings.ToLower(path.Base(rootPath))
			if strings.Contains(strings.ToLower(rootName), filter) {
				return true
			}
			if strings.Contains(strings.ToLower(browsePath[len(rootPath):]), filter) {
				return true
			}
			return false
		}
	}
	return false
}

func fileMatch(filter string, file os.FileInfo, fullPath string) (bool, error) {
	if filter == "" {
		return true, nil
	}
	if strings.Contains(strings.ToLower(file.Name()), filter) {
		return true, nil
	}
	if !file.IsDir() {
		return false, nil
	}

	files, err := ioutil.ReadDir(fullPath)
	if err != nil {
		return false, err
	}

	for _, f := range files {
		match, err := fileMatch(filter, f, path.Join(fullPath, f.Name()))
		if err != nil {
			return false, err
		}
		if match {
			return true, nil
		}
	}

	return false, nil
}
