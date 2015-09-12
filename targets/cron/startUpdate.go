package cron

import (
	"net/http"

	"github.com/koffeinsource/kaffeebot/data"

	"appengine"
	"appengine/taskqueue"
)

// DispatchStartUpdate is cronjob. It get's all URLs from the DS and puts
// an update task for all urls in a queue
func DispatchStartUpdate(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	fs, err := data.GetFeeds(c)
	if err != nil {
		c.Errorf("Error getting the urls from the DS: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for _, f := range fs {
		task := taskqueue.NewPOSTTask("/task/updater/post/", map[string][]string{"url": {f.URL}})
		if _, err := taskqueue.Add(c, task, "rss-feed-update"); err != nil {
			c.Errorf("Error while triggering the url update: %v", err)
		} else {
			c.Infof("Added %v to update queue", f.URL)
		}
	}
	w.WriteHeader(http.StatusOK)
}
