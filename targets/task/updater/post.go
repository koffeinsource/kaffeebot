package updater

import (
	"net/http"
	"net/url"
	"time"

	"github.com/koffeinsource/kaffeebot/config"
	"github.com/koffeinsource/kaffeebot/data"
	"github.com/koffeinsource/kaffeeshare/extract"

	"appengine"
)

// DispatchPOST updates the rss feed passed via post
func DispatchPOST(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	defer r.Body.Close()

	if err := r.ParseForm(); err != nil {
		c.Errorf("Error at in updatePost @ ParseForm. Error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	dsFeed, err := data.FeedStored(c, r.FormValue("url"))

	// get the url from the datastore
	if err != nil {
		c.Errorf("Error getting the url from the DS: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if dsFeed == nil {
		c.Errorf("Got no feed from the DS: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// read the url
	c.Infof("DS Feed @ importer: %v", dsFeed)

	// parse feed
	urls, err := getLinksFromFeed(dsFeed, r, c)
	if err != nil {
		c.Errorf("Error parsing the url: %v", dsFeed.URL)
		dsFeed.Fails++
		err := data.StoreFeed(c, *dsFeed)
		if err != nil {
			c.Errorf("Error while updating url %v in the DS: %v", dsFeed.URL, err)
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.Infof("URLs found in feed: %v", urls)

	dsFeed.LastUpdate = time.Now() // update time of the last update

	// put urls into kshare
	for _, u := range urls {
		importURL := config.KaffeeshareURL + "/k/share/json/" + dsFeed.Namespace //"test"
		importURL += "/?url=" + url.QueryEscape(u)
		c.Infof("URL we get %v", importURL)
		_, body, err := extract.GetURL(importURL, r)
		if err != nil {
			c.Errorf("Error while importing url into namespace: %v, %v", importURL, err)
			// we'll ignore the errors for now
			// actually we may want to offload this into another task queue for easier retries
		} else {
			c.Infof("body: %v", string(body))
		}
	}

	// update feed in DS
	dsFeed.Fails = 0 // reset error count

	err = data.StoreFeed(c, *dsFeed)
	if err != nil {
		c.Errorf("Error while updating url %v in the DS: %v", dsFeed.URL, err)
	}
	w.WriteHeader(http.StatusOK)
}
