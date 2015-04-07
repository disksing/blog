package main

import (
	"bytes"
	"net/http"
	"os"
	"path/filepath"
	"text/template"

	"github.com/huangml/log"
)

var t *template.Template
var pages = make(map[string]bytes.Buffer)

func main() {
	build()
	serv()
}

func build() {
	var err error
	t, err = template.ParseGlob("templates/*.t")
	if err != nil {
		log.Fatal(err)
		return
	}

	parseCategories()

	filepath.Walk("posts", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && filepath.Ext(path) == ".md" {
			p, err := parsePost(path)
			if err != nil {
				log.Error(err)
				return nil
			}

			var b bytes.Buffer
			err = t.ExecuteTemplate(&b, "post.t", p)
			if err != nil {
				log.Error(err)
				return nil
			}

			pages["/"+p.Name] = b
			addPostToCategory(p)
			addPostToFeed(p)
		}
		return nil
	})

	log.Info("create home page")

	var b bytes.Buffer
	err = t.ExecuteTemplate(&b, "index.t", categories)
	log.FatalOnError(err)
	pages["/"] = b

	log.Info("create feeds")
	feed := feed()

	atom, err := feed.ToAtom()
	log.FatalOnError(err)
	pages["/feed"] = *bytes.NewBufferString(atom)

	rss, err := feed.ToRss()
	log.FatalOnError(err)
	pages["/rss"] = *bytes.NewBufferString(rss)
}

func serv() {

	os.Mkdir("log", 0644)
	log.SetStatFilePath("log")

	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if b, ok := pages[r.RequestURI]; ok {
			w.Write(b.Bytes())
		} else {
			http.NotFound(w, r)
		}
	})

	http.ListenAndServe(":80", nil)
}
