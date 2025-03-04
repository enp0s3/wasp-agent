package crio

import (
	"io"
	"os"
	"path/filepath"
)

type FileOperationsImpl struct{}

func (fops FileOperationsImpl) Open(fileName string) (*os.File, error) {
	return os.Open(fileName)
}

func (fops FileOperationsImpl) Create(fileName string) (*os.File, error) {
	return os.Create(fileName)
}

func (fops FileOperationsImpl) Copy(dst io.Writer, src io.Reader) error {
	var err error
	_, err = io.Copy(dst, src)
	return err
}

func (fops FileOperationsImpl) Chmod(fileName string, mode os.FileMode) error {
	return os.Chmod(fileName, mode)
}

func (fops FileOperationsImpl) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (fops FileOperationsImpl) Walk(root string, fn filepath.WalkFunc) error {
	return filepath.Walk(root, fn)
}
func (fops FileOperationsImpl) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (fops FileOperationsImpl) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (fops FileOperationsImpl) Symlink(oldname, newname string) error {
	return os.Symlink(oldname, newname)
}

func (fops FileOperationsImpl) Dir(path string) string {
	return filepath.Dir(path)
}
