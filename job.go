package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/paradev-ru/peskar-hub/peskar"
	"github.com/paradev-ru/peskar-index/movie"
)

func SaveAsHTML(j peskar.Job, templatedir, resultdir string) error {
	templateFile := filepath.Join(templatedir, "movie.html")
	if _, err := os.Stat(templateFile); os.IsNotExist(err) {
		return err
	}
	resultDir := path.Join(resultdir, j.Directory())
	if _, err := os.Stat(resultDir); os.IsNotExist(err) {
		if err := os.Mkdir(resultDir, 0755); err != nil {
			return err
		}
	}
	logrus.Infof("%s: Parsing info page '%s'...", j.ID, j.InfoURL)
	m, err := getMovie(j)
	if err != nil {
		return err
	}
	logrus.Infof("%s: Done: %s", j.ID, m.Name)
	logrus.Infof("%s: Creating template...", j.ID)
	s, err := m.Template(templateFile)
	if err != nil {
		return err
	}
	resultFile := filepath.Join(resultDir, "index.html")
	resultF, err := os.OpenFile(resultFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("Could not open file: %v", err)
	}
	defer resultF.Close()
	_, err = resultF.WriteString(s)
	if err != nil {
		return fmt.Errorf("Could not write to file: %s", err)
	}
	logrus.Infof("%s: Saved: /%s/index.html", j.ID, m.Directory)
	return nil
}

func getMovie(j peskar.Job) (*movie.Movie, error) {
	m, err := movie.New(j.InfoURL)
	if err != nil {
		return nil, err
	}
	m.Directory = j.Directory()
	err = m.Parse()
	if err != nil {
		return nil, err
	}
	return m, nil
}
