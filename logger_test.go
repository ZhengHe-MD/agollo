package agollo

import (
	"log"
	"os"
	"testing"
	"time"
)

func TestAgolloLogger(t *testing.T) {
	// NOTE: uncomment this if you want to run it
	t.Skip("Skipping this test because it's not a regular test")

	SetLogger(log.New(os.Stderr, "", log.LstdFlags))
	err := StartWithConf(&Conf{
		AppID:          "non-app",
		Cluster:        "default",
		NameSpaceNames: []string{"application"},
		CacheDir:       "/tmp/agollo",
		IP:             "apollo-meta.ibanyu.com:30002",
	})

	if err != nil {
		t.Error(err)
	}

	time.Sleep(90 * time.Second)
}
