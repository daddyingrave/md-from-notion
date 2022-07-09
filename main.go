package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const CreatedDateFormat = "January 2, 2006 3:04 PM"

var (
	UUID     = regexp.MustCompile(`\s[\da-f]{32}`)
	FileName = regexp.MustCompile(`/[^/]+$`)
	// CreatedDate specific to my notes
	CreatedDate = regexp.MustCompile(`Created: (.+)`)
)

func main() {
	conf, err := readConf()
	if err != nil {
		log.Fatal(err)
		return
	}

	importFrom := conf.Import.Path
	exportTo := conf.Export.Path

	err = createDirIfNotExist(exportTo)
	if err != nil {
		return
	}

	err = filepath.Walk(importFrom, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return nil
		}

		if strings.Contains(path, "Untitled Database") || !strings.HasSuffix(path, ".md") {
			return nil
		}

		if !info.IsDir() {
			var content, err = os.ReadFile(path)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			newPath := strings.ReplaceAll(UUID.ReplaceAllString(path, ""), importFrom, exportTo)

			err = createDirIfNotExist(FileName.ReplaceAllString(newPath, ""))
			if err != nil {
				return err
			}

			err = os.WriteFile(newPath, content, 0666)
			if err != nil {
				return err
			}

			err = os.Chtimes(newPath, time.Now(), extractDate(content))

			if err != nil {
				fmt.Println(err)
			}
		}
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
}

func createDirIfNotExist(dir string) error {
	_, err := os.ReadDir(dir)
	if err != nil {
		err := os.MkdirAll(dir, 0777)
		if err != nil {
			log.Fatal(err)
			return err
		}
	}
	return nil
}

type Conf struct {
	Import struct {
		Path string
	}
	Export struct {
		Path string
	}
}

func readConf() (*Conf, error) {
	conf, err := os.ReadFile("import-conf.yaml")
	if err != nil {
		log.Fatal(err, "import-conf.yaml required to be exist")
		return nil, err
	}

	c := Conf{}

	err = yaml.Unmarshal(conf, &c)
	if err != nil {
		log.Fatal(err)
	}

	if c.Import.Path == "" || c.Export.Path == "" {
		return nil, errors.New("both Import and Export paths should be non-empty")
	}

	return &c, nil
}

func extractDate(noteContent []byte) time.Time {
	// custom fields usually listed in the beginning of the note
	var contentLength = 0
	if len(noteContent) < 300 {
		contentLength = len(noteContent)
	} else {
		contentLength = 300
	}

	matches := CreatedDate.FindSubmatch(noteContent[:contentLength])
	if len(matches) < 1 {
		return time.Now()
	} else {
		date, err := time.Parse(CreatedDateFormat, string(matches[1]))
		if err != nil {
			log.Fatal(err)
		}

		log.Println(date)

		return date
	}
}
