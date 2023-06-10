package util

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
)

// ReadFile reads the content of a file and returns it as a byte slice.
// It takes a file path as string and returns the content and an error if any.
func ReadFile(file string) (conf []byte, err error) {
	conf, err = os.ReadFile(file)
	return
}

// SaveFile saves the content as a byte slice to a file with the given filename.
// It takes the filename and content as byte slice and returns an error if any.
func SaveFile(fileName string, content []byte) error {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	if err = file.Truncate(0); err != nil {
		return err
	}

	// write file
	_, err = file.Write(content)
	if err != nil {
		return err
	}
	return nil
}

// Download downloads a file from the given URI and saves it to the given storage path.
// It takes the URI, storage path, and file permission as input and returns the storage path and an error if any.
func Download(uri, storage string, perm os.FileMode) (string, error) {

	dir, name := path.Split(storage)
	if err := checkDir(dir); err != nil {
		return "", fmt.Errorf("check dir error: %s", err)
	}

	if len(name) <= 0 {
		return "", fmt.Errorf("%s, no file name", storage)
	}

	resp, err := http.Get(uri)
	if err != nil {
		return "", fmt.Errorf("download error: %s", err)
	}
	defer resp.Body.Close()

	file, err := os.OpenFile(storage, os.O_CREATE|os.O_RDWR, perm)
	if err != nil {
		return "", fmt.Errorf("open file error: %s", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return storage, fmt.Errorf("download error: %s", err)
	}
	return storage, nil
}

// Update replaces the file at the destination path with the file at the source path.
// It takes the source and destination paths as input and returns an error if any.
func Update(tmp, des string) error {

	// bug need fix:
	dir, name := path.Split(tmp)
	if err := checkDir(dir); err != nil {
		return fmt.Errorf("check tmp dir error: %s", err)
	}
	if len(name) <= 0 {
		return fmt.Errorf("%s, no file name", tmp)
	}

	dir, name = path.Split(des)
	if err := checkDir(dir); err != nil {
		return fmt.Errorf("check des dir error: %s", err)
	}
	if len(name) <= 0 {
		return fmt.Errorf("%s, no file name", des)
	}

	if _, err := os.Stat(des); err == nil {
		if err := os.Rename(des, des+".back"); err != nil {
			return fmt.Errorf("back up file %s error: %s", des, err)
		}
	}

	if err := os.Rename(tmp, des); err != nil {
		return fmt.Errorf("copy to des dir error: %s", err)
	}
	return nil
}

// checkDir checks if the directory exists at the given path,
// creates it if it doesn't exist, and returns an error if any.
func checkDir(dir string) error {
	if info, err := os.Stat(dir); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	} else if !info.IsDir() {
		err := os.Remove(dir)
		if err != nil {
			return fmt.Errorf("can not remove file %s", dir)
		}
	} else {
		return nil
	}
	return os.MkdirAll(dir, 0755)
}
