package atomic

import (
	"bytes"
	"crypto/sha256"
	"io"
	"os"
	"path"
	"reflect"
)

// WriteFile writes data to a file named by filename. If the contents of the file
// are the same as the data, the file is not written.
func WriteFile(filename string, data []byte, perm os.FileMode) error {

	h := sha256.New()
	h.Write(data)
	newHash := h.Sum(nil)

	// get the sha256 of the file contents if it exists
	existingHash, _ := getSha256(filename)

	if reflect.DeepEqual(newHash, existingHash) {
		return nil
	}

	// open temp file in same directory as target file
	tempFile, err := os.CreateTemp(path.Dir(filename), path.Base(filename))
	if err != nil {
		return err
	}
	tname := tempFile.Name()
	defer os.Remove(tname)

	if _, err := io.Copy(tempFile, io.NopCloser(bytes.NewReader(data))); err != nil {
		return err
	}
	if err := tempFile.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tname, perm); err != nil {
		return err
	}

	if err := os.Rename(tname, filename); err != nil {
		return err
	}

	return nil
}

// getSha256 returns the sha256 of the file contents if it exists
func getSha256(filename string) ([]byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}
