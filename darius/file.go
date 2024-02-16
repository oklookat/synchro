package darius

import (
	"errors"
	"io"
	"os"
	"path/filepath"
)

type File struct {
	// absolute path.
	abs string

	// filename with extension.
	name string
}

// Get absolute path to file.
func (f File) Abs() string {
	return f.abs
}

// Delete file.
//
// If file not exists - do nothing.
func (f File) Delete() error {
	if err := isExists(f.abs); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	return os.Remove(f.abs)
}

func (f File) Write(p []byte) (int, error) {
	file, err := f.openRW()
	if err != nil {
		return 0, err
	}
	defer file.Close()

	return file.Write(p)
}

func (f File) Clean() error {
	file, err := f.openRW()
	if err != nil {
		return err
	}
	defer file.Close()
	file.Truncate(0)
	_, err = file.Seek(0, io.SeekStart)
	return err
}

// Create file if not exists.
//
// fpath: relative path to file like "file.txt"
func (f *File) new(fpath string) error {
	abs, err := _appData.RelToAbs(fpath)
	if err != nil {
		return err
	}
	f.abs = abs
	if err := f.CreateIfNotExists(); err != nil {
		return err
	}

	f.name = filepath.Base(abs)

	return err
}

// Create file if not exists.
func (f File) CreateIfNotExists() error {
	err := isExists(f.abs)
	if err == nil {
		return nil
	}

	if errors.Is(err, os.ErrNotExist) {
		file, err := os.Create(f.abs)
		if err != nil {
			return err
		}
		return file.Close()
	}

	return err
}

// Open file (RW).
func (f File) openRW() (*os.File, error) {
	if err := isExists(f.abs); err != nil {
		return nil, err
	}
	return os.OpenFile(f.abs, os.O_RDWR, _PERM)
}
