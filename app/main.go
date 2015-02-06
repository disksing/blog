package main

import (
	"bytes"
	"fmt"
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

			pages[p.Name] = b
			addPostToCategory(p)
		}
		return nil
	})

	var b bytes.Buffer
	err = t.ExecuteTemplate(&b, "index.t", categories)
	log.FatalOnError(err)
	pages["index"] = b
}

func serv() {

	os.Mkdir("log", 0644)
	log.SetStatFilePath("log")

	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		b := func() *bytes.Buffer {
			if b, ok := pages[r.RequestURI[1:]]; ok {
				return &b
			}

			if b, ok := pages["index"]; ok {
				return &b
			}

			return nil
		}()

		if b != nil {
			//w.Header().Add("Cache-Control", "max-age=3600")
			w.Write(b.Bytes())
			log.Stat(fmt.Sprintf("pv %s %s", r.RequestURI, r.Referer()))
		} else {
			http.NotFound(w, r)
		}
	})

	http.ListenAndServe(":80", nil)
}
