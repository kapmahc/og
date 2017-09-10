package nut

import (
	"time"

	"golang.org/x/tools/blog/atom"
)

var _rss []*atom.Entry

// AddRssEntry add rss entry
func AddRssEntry(link, title, summary, author string, updated time.Time) {
	_rss = append(_rss, &atom.Entry{
		Title: title,
		ID:    link,
		Link: []atom.Link{
			{Href: link},
		},
		Summary: &atom.Text{Body: summary},
		Author: &atom.Person{
			Name: author,
		},
	})
}
