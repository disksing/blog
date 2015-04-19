package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/huangml/log"
	"github.com/russross/blackfriday"
)

type Post struct {
	Name        string
	Title       string
	Description string
	Time        time.Time
	Category    string
	Content     string
	HTML        string
}

func parsePost(path string) (*Post, error) {
	log.Info("Parse:", path)

	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	p := &Post{
		Name: strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
	}

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		s := scanner.Text()

		if strings.HasPrefix(s, "+++") {
			break
		}

		sp := strings.SplitN(s, ":", 2)
		if len(sp) != 2 {
			return nil, fmt.Errorf("invalid header: %s", s)
		}

		k, v := strings.TrimSpace(sp[0]), strings.TrimSpace(sp[1])
		switch k {
		case "title":
			p.Title = v
		case "description":
			p.Description = v
		case "time":
			p.Time, err = time.Parse("2006/01/02 15:04", v)
			if err != nil {
				return nil, err
			}
		case "category":
			p.Category = v
		case "html":
			p.HTML = v
		default:
			return nil, fmt.Errorf("invalid header: %s", s)
		}
	}

	if p.HTML != "" {
		b, _ := ioutil.ReadFile(p.HTML)
		p.Content = string(b)
	} else {
		var content bytes.Buffer
		for scanner.Scan() {
			content.Write(scanner.Bytes())
			content.WriteString("\n")
		}

		p.Content = string(blackfriday.MarkdownCommon(content.Bytes()))
	}

	return p, nil
}
