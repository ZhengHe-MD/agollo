package agollo

import (
	"time"
)

const (
	defaultConfName  = "app.properties"
	defaultNamespace = "application"

	longPollInterval      = time.Second * 2
	// NOTE: apollo will return 304 after 60 secs when querying config for a
	//       unknown app, so longPollTimeout should be larger than 60 secs
	longPollTimeout       = time.Second * 90
	queryTimeout          = time.Second * 2
	defaultNotificationID = -1
)
