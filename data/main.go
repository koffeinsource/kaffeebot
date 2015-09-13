package data

import (
	"strings"
	"time"
)

// A Feed is all the data we store for one feed
type Feed struct {
	URL        string    `json:"-" datastore:"URL,index"`
	Namespace  string    `json:"namespace" datastore:"Namespace,noindex"`
	LastUpdate time.Time `json:"-" datastore:"LastUpdate,index"`
	LastURL    string    `json:"-" datastore:"LastURL,noindex"`
	Fails      int       `json:"-" datastore:"Fails,index"`
}

// NewFeed creates a Feed from an url
func NewFeed(url string) Feed {
	feed := Feed{URL: url, Namespace: url, LastUpdate: time.Now(), Fails: 0}

	// generate namespace
	feed.Namespace = strings.Replace(feed.Namespace, "https://", "", -1)
	feed.Namespace = strings.Replace(feed.Namespace, "http://", "", -1)
	feed.Namespace = strings.Replace(feed.Namespace, ":", "", -1)
	feed.Namespace = strings.Replace(feed.Namespace, "/", "", -1)
	feed.Namespace = strings.Replace(feed.Namespace, "?", "", -1)
	feed.Namespace = strings.Replace(feed.Namespace, "&", "", -1)
	feed.Namespace = strings.Replace(feed.Namespace, "@", "", -1)
	feed.Namespace = strings.Replace(feed.Namespace, ".", "", -1)

	feed.Namespace = "kbotrss" + feed.Namespace

	return feed
}
