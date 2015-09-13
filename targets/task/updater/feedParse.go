package updater

import (
	"encoding/xml"
	"errors"
	"html/template"
	"net/http"

	"appengine"

	"github.com/koffeinsource/kaffeebot/data"
	"github.com/koffeinsource/kaffeeshare/extract"
)

type rss2 struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	// Required
	Title       string `xml:"channel>title"`
	Link        string `xml:"channel>link"`
	Description string `xml:"channel>description"`
	// Optional
	PubDate  string     `xml:"channel>pubDate"`
	ItemList []rss2Item `xml:"channel>item"`
}

type rss2Item struct {
	// Required
	Title       string        `xml:"title"`
	Link        string        `xml:"link"`
	Description template.HTML `xml:"description"`
	// Optional
	Content  template.HTML `xml:"encoded"`
	PubDate  string        `xml:"pubDate"`
	Comments string        `xml:"comments"`
}

type atom1 struct {
	XMLName   xml.Name     `xml:"http://www.w3.org/2005/Atom feed"`
	Title     string       `xml:"title"`
	Subtitle  string       `xml:"subtitle"`
	ID        string       `xml:"id"`
	Updated   string       `xml:"updated"`
	Rights    string       `xml:"rights"`
	Link      link         `xml:"link"`
	Author    author       `xml:"author"`
	EntryList []atom1Entry `xml:"entry"`
}

type link struct {
	Href string `xml:"href,attr"`
}

type author struct {
	Name  string `xml:"name"`
	Email string `xml:"email"`
}

type atom1Entry struct {
	Title   string `xml:"title"`
	Summary string `xml:"summary"`
	Content string `xml:"content"`
	ID      string `xml:"id"`
	Updated string `xml:"updated"`
	Link    link   `xml:"link"`
	Author  author `xml:"author"`
}

// convert atom1 feed entry to rss2 feed
// we only deal with rss2 feeds
func atom1ToRss2(atom atom1) *rss2 {
	r := rss2{
		Title:       atom.Title,
		Link:        atom.Link.Href,
		Description: atom.Subtitle,
		PubDate:     atom.Updated,
	}
	r.ItemList = make([]rss2Item, len(atom.EntryList))
	for i, entry := range atom.EntryList {
		r.ItemList[i].Title = entry.Title
		r.ItemList[i].Link = entry.Link.Href
		if entry.Content != "" {
			r.ItemList[i].Description = template.HTML(entry.Content)
		} else {
			r.ItemList[i].Description = template.HTML(entry.Summary)
		}
	}
	return &r
}

func parseAtom(content []byte) (*rss2, error) {
	var a atom1
	err := xml.Unmarshal(content, &a)
	if err != nil {
		return nil, err
	}
	return atom1ToRss2(a), nil
}

func parseFeedContent(content []byte) (*rss2, error) {
	// we first try to parse it as rss2 feed
	// if that will return an error we try atom
	var v rss2
	err := xml.Unmarshal(content, &v)
	if err != nil {
		// ok that is not nice, but works for now
		if err.Error() == "expected element type <rss> but have <feed>" {
			// try Atom 1.0
			return parseAtom(content)
		}
		return nil, err
	}

	if v.Version == "2.0" {
		// RSS 2.0
		for i := range v.ItemList {
			if v.ItemList[i].Content != "" {
				v.ItemList[i].Description = v.ItemList[i].Content
			}
		}
		return &v, nil
	}

	return &v, errors.New("not rss 2.0")
}

// extracts the URL from an Feed. If the feed has publication date, we only
// add the URL since the last feed update
func getLinksFromFeed(dsFeed *data.Feed, r *http.Request, c appengine.Context) ([]string, error) {
	c.Infof("DS Feed @ importer: %v", dsFeed)
	_, body, err := extract.GetURL(dsFeed.URL, r)
	if err != nil {
		return nil, err
	}

	rss, err := parseFeedContent(body)
	if err != nil {
		return nil, err
	}

	var urls []string
	stopURL := dsFeed.LastURL
	for index, i := range rss.ItemList {
		if index == 0 {
			dsFeed.LastURL = i.Link
		}
		if i.Link == stopURL {
			break
		}
		urls = append(urls, i.Link)
	}

	return urls, nil
}
