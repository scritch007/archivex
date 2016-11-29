package archivex

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

//Archiver the Archiver struct
type Archiver struct {
	w       *zip.Writer
	Debug   io.Writer //Log what is happening
	Ignore  []string
	f       *os.File
	SubPath string
}

//Create the Archive
func (a *Archiver) Create(name string) error {

	f, err := os.Create(name)
	if err != nil {
		return err
	}
	return a.CreateWithWriter(f)

}

//Close the archiver
func (a *Archiver) Close() {
	if a.f != nil {
		a.Close()
	}
	if a.w != nil {
		a.w.Close()
	}
}

//log log what is happening
func (a *Archiver) log(debug string) {
	if a.Debug != nil {
		a.Debug.Write([]byte(debug))
	}
}

//CreateWithWriter create the archive using the writer
func (a *Archiver) CreateWithWriter(w io.Writer) error {
	a.w = zip.NewWriter(w)
	return nil
}

//Add element to the archive
func (a *Archiver) Add(name string) error {
	name = path.Clean(name)
	if a.w == nil {
		return errors.New("Archive not initialized")
	}
	fileInfo, err := os.Lstat(name)
	if err != nil {
		return err
	}

	if !fileInfo.IsDir() {
		return a.addFile(name)
	}
	// Walk directory.
	filepath.Walk(name, func(path string, info os.FileInfo, err error) error {
		// Remove base path, convert to forward slash.
		zipPath := path[len(a.SubPath):]
		zipPath = strings.TrimLeft(strings.Replace(zipPath, `\`, "/", -1), `/`)
		if info.IsDir() {
			// Create a header based off of the fileinfo
			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}

			// Set the header's name to what we want--it may not include the top folder
			header.Name = zipPath + "/"

			a.log(fmt.Sprintf("Adding folder %s header as %s\n", path, zipPath))
			// Get a writer in the archive based on our header
			_, err = a.w.CreateHeader(header)
			return err
		}
		return a.addFile(path)
	})
	return nil
}

func (a *Archiver) addFile(name string) error {

	zipPath := name[len(a.SubPath):]
	zipPath = strings.TrimLeft(strings.Replace(zipPath, `\`, "/", -1), `/`)
	a.log(fmt.Sprintf("Adding %s to archive as %s\n", name, zipPath))
	ze, err := a.w.Create(zipPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot create zip entry <%s>: %s\n", zipPath, err)
		return err
	}
	file, err := os.Open(name)
	if err != nil {
		return err
	}
	defer file.Close()
	io.Copy(ze, file)
	return nil
}
