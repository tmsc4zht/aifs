package aifs

import "io/fs"

type VirtualDir struct {
	fs.DirEntry
}

func (VirtualDir) IsDir() bool {
	return true
}
