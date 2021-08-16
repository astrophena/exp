package dokuxmlrpc

import (
	"log"
	"os"
	"testing"
)

var c *Client

func TestMain(m *testing.M) {
	var err error
	c, err = NewClient("https://wiki.internal.astrophena.name")
	if err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}
	os.Exit(m.Run())
}

func TestVersion(t *testing.T) {
	v, err := c.Version()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("DokuWiki version: " + v)
}

func TestTitle(t *testing.T) {
	title, err := c.Title()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Wiki title: " + title)
}
