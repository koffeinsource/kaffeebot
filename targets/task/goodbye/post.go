package goodbye

import (
	"net/http"
	"net/url"

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

	importURL := config.KaffeeshareURL + "/k/share/json/" + dsFeed.Namespace
	importURL += "/?url=" + url.QueryEscape("https://kaffeebot.com/stops/updating/goodbye")
	c.Infof("URL we get %v", importURL)
	_, body, err := extract.GetURL(importURL, r)
	if err != nil {
		c.Errorf("Error while importing url into namespace: %v, %v", importURL, err)
		// we'll ignore the errors for now
		// actually we may want to offload this into another task queue for easier retries
	} else {
		c.Infof("body: %v", string(body))
	}

	w.WriteHeader(http.StatusOK)
}
