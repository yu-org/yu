package compress

import (
	"archive/zip"
	"github.com/pkg/errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func ZipFiles(files []string, output string) error {
	zipFile, err := os.Create(output)
	if err != nil {
		return err
	}
	defer zipFile.Close()
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()
	for _, file := range files {
		err = addFileToZip(zipWriter, file)
		if err != nil {
			return err
		}
	}
	return nil
}

func UnzipFile(src, destDir string) ([]string, error) {
	fnames := make([]string, 0)

	r, err := zip.OpenReader(src)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(destDir, f.Name)
		// Check for ZipSlip.
		if !strings.HasPrefix(fpath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return nil, errors.Errorf("Illegal file path: %s", fpath)
		}
		fnames = append(fnames, fpath)

		if f.FileInfo().IsDir() {
			err := os.MkdirAll(fpath, os.ModePerm)
			if err != nil {
				return nil, err
			}
			continue
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return nil, err
		}

		rc, err := f.Open()
		if err != nil {
			return nil, err
		}

		_, err = io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()
		if err != nil {
			return nil, err
		}
	}

	return fnames, nil
}

func addFileToZip(w *zip.Writer, fname string) error {
	fileToZip, err := os.Open(fname)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	header.Name = fname
	header.Method = zip.Deflate

	writer, err := w.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToZip)
	return err
}
