// +build appengine

package data

import (
	"time"

	"github.com/koffeinsource/kaffeebot/config"

	"appengine"
	"appengine/datastore"
)

// StoreFeed stores a feed in the datastore
func StoreFeed(c appengine.Context, i Feed) error {
	k := datastore.NewKey(c, "Feed", i.URL, 0, nil)
	_, err := datastore.Put(c, k, &i)
	c.Infof("Stored feed %v", i)
	if err != nil {
		c.Errorf("Error while storing feed in datastore. Feed: %v. Error: %v", i, err)
	}

	return err
}

// GetFeeds returns the feeds that must be updated
func GetFeeds(c appengine.Context) ([]Feed, error) {
	q := datastore.NewQuery("Feed").
		Filter("LastUpdate <", time.Now().Truncate(time.Duration(config.UpdateFeedsEveryXMinutes*time.Minute)))

	var fs []Feed
	t := q.Run(c)
	for {
		var f Feed
		_, err := t.Next(&f)
		if err == datastore.Done {
			break
		}

		fs = append(fs, f)
		if err != nil {
			c.Errorf("Error fetching next feed: %v", err)
			return nil, err
		}
	}

	return fs, nil
}

// GetFailedFeeds returns the feeds with an fail count higher than
// config.FeedFailsBeforeBroken
func GetFailedFeeds(c appengine.Context) ([]Feed, []*datastore.Key, error) {
	q := datastore.NewQuery("Feed").
		Filter("Fails >", config.FeedFailsBeforeBroken)

	var fs []Feed
	var ks []*datastore.Key
	t := q.Run(c)
	for {
		var f Feed
		k, err := t.Next(&f)
		if err == datastore.Done {
			break
		}

		fs = append(fs, f)
		ks = append(ks, k)
		if err != nil {
			c.Errorf("Error fetching next feed: %v", err)
			return nil, nil, err
		}
	}

	return fs, ks, nil
}

// FeedStored checks if the feed is already in the datastore
func FeedStored(c appengine.Context, feedURL string) (*Feed, error) {
	c.Infof("Searching for feed %v", feedURL)
	q := datastore.NewQuery("Feed").
		Filter("URL=", feedURL).
		Limit(1)

	var f Feed

	t := q.Run(c)
	_, err := t.Next(&f)
	if err == datastore.Done || &f == nil {
		c.Infof("Feed not found")
		return nil, nil
	}

	c.Infof("Feed found")
	return &f, err
}

// DeleteFeeds deletes feeds in the datastore
func DeleteFeeds(c appengine.Context, ks []*datastore.Key) error {
	c.Infof("Going to delete %v feeds from the DS.", len(ks))

	return datastore.DeleteMulti(c, ks)
}
