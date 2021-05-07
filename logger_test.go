package agollo

import (
	"log"
	"os"
	"testing"
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
		IP:             "anyhost:anyip",
	})

	if err != nil {
		t.Error(err)
	}
}
