package config

const (
	// KaffeeshareURL is the URL of the kshare instance you want to put the feeds
	KaffeeshareURL = "https://kaffeeshare.appspot.com"

	// FeedFailsBeforeBroken is the number of times a feed may return an error
	// before we delete it from the database
	FeedFailsBeforeBroken = 100

	// UpdateFeedsEveryXMinutes tells how often a feed should be checked for new
	// entries
	UpdateFeedsEveryXMinutes = 30
)
