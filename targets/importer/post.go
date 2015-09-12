package importer

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/koffeinsource/kaffeebot/data"

	"appengine"
	"appengine/taskqueue"
)

// DispatchPOST import the url passed via POST as a feed to the database
// and triggers an update if it was not in the DS before
func DispatchPOST(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	// get URL
	r.ParseForm()
	rssurl := r.Form.Get("rssurl")

	if rssurl == "" {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !(strings.HasPrefix(rssurl, "http://") || strings.HasPrefix(rssurl, "https://")) {
		rssurl = "https://" + rssurl
	}

	// check if URL is valid
	if !govalidator.IsRequestURL(rssurl) {
		c.Errorf("Invalid URL. URL: %v", rssurl)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// check if URL is already in the datastore
	if f, err := data.FeedStored(c, rssurl); err != nil {
		c.Errorf("Error while checking if url is in DS. URL: %v, error: %v", rssurl, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if f != nil {
		c.Infof("Feed is already in the DS %v", f)
		// yes it is, lets marshal our reply and return it
		b, err := json.Marshal(*f)
		if err != nil {
			c.Errorf("Error marshalling json: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(b)
		return
	}

	// no it is not in the DS
	// add the url to the datastore
	feed := data.NewFeed(rssurl)

	if err := data.StoreFeed(c, feed); err != nil {
		c.Errorf("Error storing url in DS: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// trigger import for that url
	task := taskqueue.NewPOSTTask("/task/updater/post/", map[string][]string{"url": {feed.URL}})
	if _, err := taskqueue.Add(c, task, "rss-feed-update"); err != nil {
		c.Errorf("Error while triggering the url update: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// return the data
	b, err := json.Marshal(feed)
	if err != nil {
		c.Errorf("Error marshalling json: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(b)
}
