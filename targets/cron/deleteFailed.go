package cron

import (
	"net/http"

	"github.com/koffeinsource/kaffeebot/data"

	"appengine"
	"appengine/taskqueue"
)

// DispatchDeleteFailed is cronjob. It get's all URLs that failed too often.
func DispatchDeleteFailed(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	fs, ks, err := data.GetFailedFeeds(c)
	if err != nil {
		c.Errorf("Error getting the urls from the DS: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for _, f := range fs {
		task := taskqueue.NewPOSTTask("/task/goodbye/post/", map[string][]string{"url": {f.URL}})
		if _, err := taskqueue.Add(c, task, "rss-feed-update"); err != nil {
			c.Errorf("Error while triggering the url update: %v", err)
		} else {
			c.Infof("Added %v to update queue", f.URL)
		}
	}

	err = data.DeleteFeeds(c, ks)
	if err != nil {
		c.Errorf("Error delete the feeds from the DS: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
