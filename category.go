package main

import (
	"bufio"
	"os"
	"strings"

	"github.com/huangml/log"
)

type Category struct {
	Name        string
	Title       string
	Description string
	Posts       []*Post
}

var categories []*Category

func parseCategories() {
	log.Info("Parse Categories")

	f, err := os.OpenFile("posts/category", os.O_RDONLY, 0)
	log.FatalOnError(err)
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		sp := strings.SplitN(scanner.Text(), ":", 3)
		c := &Category{}
		if len(sp) > 0 {
			c.Name = sp[0]
		}
		if len(sp) > 1 {
			c.Title = sp[1]
		}
		if len(sp) > 2 {
			c.Description = sp[2]
		}

		categories = append(categories, c)
	}
}

func addPostToCategory(p *Post) {
	for _, c := range categories {
		if p.Category == c.Name || c.Name == "others" {
			for i, cp := range c.Posts {
				if p.Time.Before(cp.Time) {
					c.Posts = append(c.Posts, nil)
					copy(c.Posts[i+1:], c.Posts[i:])
					c.Posts[i] = p
					return
				}
			}
			c.Posts = append(c.Posts, p)
			return
		}
	}
}
