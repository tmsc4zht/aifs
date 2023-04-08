package aifs

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"

	"github.com/h2non/filetype"
	"github.com/h2non/filetype/matchers"
	"github.com/h2non/filetype/types"
)

type FS struct {
	fsys fs.FS
}

func New(fsys fs.FS) FS {
	return FS{fsys: fsys}
}

func (aifs FS) Open(name string) (fs.File, error) {
	return aifs.open(aifs.fsys, sepFilePath(filepath.Clean(name)))
}

func (aifs FS) open(currentFS fs.FS, paths []string) (fs.File, error) {
	// len(paths) == 1
	// => 終端に達したということ
	// フォルダなら開いて渡す
	// archiveならファイルシステムを開いて渡す
	// それ以外は通常ファイルとして渡す
	if len(paths) == 1 {
		isDir, err := getIsDir(currentFS, paths[0])
		if err != nil {
			return nil, err
		}
		if isDir {
			return currentFS.Open(paths[0])
		}
		kind, err := getKind(currentFS, paths[0])
		if err != nil {
			return nil, err
		}
		switch kind {
		case matchers.TypeZip:
			zr, err := zipNewReader(currentFS, paths[0])
			if err != nil {
				return nil, fmt.Errorf("could not zip as reader: %v", err)
			}
			return zr.Open(".")
		}
		return currentFS.Open(paths[0])
	}

	// 今がディレクトリなら後ろのを結合して次に渡す
	if isdir, err := getIsDir(currentFS, paths[0]); isdir || err != nil {
		if err != nil {
			return nil, err
		}

		if isdir {
			paths[1] = paths[0] + `/` + paths[1]
			return aifs.open(currentFS, paths[1:])
		}
	}

	// 今がファイル
	kind, err := getKind(currentFS, paths[0])
	if err != nil {
		return nil, fmt.Errorf("could not find kind %v: %v", paths[0], err)
	}

	// zipだったら
	switch kind {
	case matchers.TypeZip:
		zr, err := zipNewReader(currentFS, paths[0])
		if err != nil {
			return nil, fmt.Errorf("could not zip as reader: %v", err)
		}
		return aifs.open(zr, paths[1:])
	}

	return nil, fmt.Errorf("could not open %v as directory", paths[0])
}

func (aifs FS) ReadDir(name string) ([]fs.DirEntry, error) {
	return aifs.readDir(aifs.fsys, sepFilePath(filepath.Clean(name)))
}

func (aifs FS) readDir(currentFS fs.FS, paths []string) ([]fs.DirEntry, error) {
	// len(paths) == 1
	// => 終端に達したということ
	// フォルダなら開いて渡す
	// archiveならファイルシステムを開いて渡す
	// それ以外は通常ファイルとして渡す
	if len(paths) == 1 {
		isDir, err := getIsDir(currentFS, paths[0])
		if err != nil {
			return nil, err
		}
		if isDir {
			return readDirWrapper(currentFS, paths[0])
		}
		kind, err := getKind(currentFS, paths[0])
		if err != nil {
			return nil, err
		}
		switch kind {
		case matchers.TypeZip:
			zr, err := zipNewReader(currentFS, paths[0])
			if err != nil {
				return nil, fmt.Errorf("could not zip as reader: %v", err)
			}
			return readDirWrapper(zr, ".")
		}
		return readDirWrapper(currentFS, paths[0])
	}

	// 今がディレクトリなら後ろのを結合して次に渡す
	if isdir, err := getIsDir(currentFS, paths[0]); isdir || err != nil {
		if err != nil {
			return nil, err
		}

		if isdir {
			paths[1] = paths[0] + `/` + paths[1]
			return aifs.readDir(currentFS, paths[1:])
		}
	}

	// 今がファイル
	kind, err := getKind(currentFS, paths[0])
	if err != nil {
		return nil, fmt.Errorf("could not find kind %v: %v", paths[0], err)
	}

	// zipだったら
	switch kind {
	case matchers.TypeZip:
		zr, err := zipNewReader(currentFS, paths[0])
		if err != nil {
			return nil, fmt.Errorf("could not zip as reader: %v", err)
		}
		return aifs.readDir(zr, paths[1:])
	}

	return nil, fmt.Errorf("could not open %v as directory", paths[0])

}

func readDirWrapper(fsys fs.FS, name string) ([]fs.DirEntry, error) {
	entries, err := fs.ReadDir(fsys, name)
	if err != nil {
		return nil, err
	}
	for i, entry := range entries {
		if entry.IsDir() {
			continue
		}
		kind, _ := getKind(fsys, name+`/`+entry.Name())
		switch kind {
		case matchers.TypeZip:
			entries[i] = VirtualDir{entry}
		}
	}
	return entries, nil
}

// helper functions
func zipNewReader(fsys fs.FS, name string) (fs.FS, error) {
	f, err := fsys.Open(name)
	if err != nil {
		return nil, err
	}
	s, err := f.Stat()
	if err != nil {
		return nil, err
	}

	r, ok := f.(io.ReaderAt)
	if !ok {
		return nil, fmt.Errorf("could not %s as io.ReaderAt", name)
	}

	return zip.NewReader(r, s.Size())
}

func getIsDir(fsys fs.FS, name string) (bool, error) {
	f, err := fsys.Open(name)
	if err != nil {
		return false, fmt.Errorf("could not open %v: %v", name, err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return false, fmt.Errorf("could not get stat %v: %v", name, err)
	}

	return info.IsDir(), nil
}

func getKind(fsys fs.FS, name string) (types.Type, error) {
	f, err := fsys.Open(name)
	if err != nil {
		return types.Unknown, fmt.Errorf("could not open %v: %v", name, err)
	}
	defer f.Close()

	head := make([]byte, 261)
	f.Read(head)
	return filetype.Match(head)
}

// name must be cleaned by filepath.Clean
func sepFilePath(name string) []string {
	ret := []string{}

	for {
		base := filepath.Base(name)
		name = filepath.Dir(name)

		if base == "." {
			break
		}
		ret = append(ret, base)

	}

	if len(ret) == 0 {
		ret = append(ret, ".")
	}

	// 逆順にする
	for i, j := 0, len(ret)-1; i < j; i, j = i+1, j-1 {
		ret[i], ret[j] = ret[j], ret[i]
	}

	return ret
}
