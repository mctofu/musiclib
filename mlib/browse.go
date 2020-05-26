package mlib

type BrowseOptions struct {
	TextFilter string
}

type BrowseItem struct {
	Name   string
	URI    string
	Folder bool
}
