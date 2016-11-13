package main

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
)

func Untar(mpath, dest string) error {
	fr, err := os.OpenFile(mpath, os.O_RDONLY, 0444)
	if err != nil {
		return err
	}
	defer fr.Close()
	if err != nil {
		return err
	}
	tr := tar.NewReader(fr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		path := hdr.Name
		switch hdr.Typeflag {
		case tar.TypeReg:
			path = filepath.Join(dest, filepath.Base(path))
			ow, err := overwrite(path)
			if err != nil {
				return err
			}
			if _, err := io.Copy(ow, tr); err != nil {
				return err
			}
			if err := ow.Close(); err != nil {
				return err
			}
		}
	}
	return nil
}

func overwrite(mpath string) (*os.File, error) {
	f, err := os.OpenFile(mpath, os.O_RDWR|os.O_TRUNC, 0777)
	if err != nil {
		f, err = os.Create(mpath)
		if err != nil {
			return f, err
		}
	}
	return f, nil
}
