package main

import (
	"time"

	"github.com/gorilla/feeds"
)

var recentPosts []*Post

const feedNum = 10

func addPostToFeed(p *Post) {
	for i, rp := range recentPosts {
		if p.Time.After(rp.Time) {
			recentPosts = append(recentPosts, nil)
			copy(recentPosts[i+1:], recentPosts[i:])
			recentPosts[i] = p
			return
		}
	}
	recentPosts = append(recentPosts, p)
}

func feed() *feeds.Feed {
	now := time.Now()
	feed := &feeds.Feed{
		Title:       "硬盘在歌唱",
		Link:        &feeds.Link{Href: "http://disksing.com"},
		Description: "我编程三日，两耳不闻人声，只有硬盘在歌唱。",
		Author:      &feeds.Author{"disksing", "menglong@outlook.com"},
		Created:     now,
	}

	for i := 0; i < feedNum && i < len(recentPosts); i++ {
		p := recentPosts[i]
		feed.Items = append(feed.Items, &feeds.Item{
			Title:       p.Title,
			Link:        &feeds.Link{Href: "http://disksing.com/" + p.Name},
			Description: p.Content,
			Created:     p.Time,
		})
	}

	return feed
}
