// Command mailcheck is a simple mail checker.
//
// This code is merged into go.astrophena.name/infra/cmd/tgbotd.
package main

import (
	"flag"
	"log"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

func main() {
	log.SetFlags(0)

	var (
		addr     = flag.String("addr", "imap.gmail.com:993", "connect to `host:port`")
		username = flag.String("username", "", "authenticate with `username`")
		password = flag.String("password", "", "authenticate with `password`")
	)
	flag.Parse()

	// TODO: on tgbotd, protect c with mutex.

	c, err := client.DialTLS(*addr, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Logout()

	if err := c.Login(*username, *password); err != nil {
		log.Fatal(err)
	}

	// Open read-only, so we can't accidentially delete or modify
	// messages.
	_, err = c.Select("INBOX", true)
	if err != nil {
		log.Fatal(err)
	}

	// Fetch unread messages, if any.
	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{imap.SeenFlag}
	ids, err := c.Search(criteria)
	if err != nil {
		log.Fatal(err)
	}

	if len(ids) > 0 {
		seqset := new(imap.SeqSet)
		seqset.AddNum(ids...)

		messages := make(chan *imap.Message, len(ids))
		items := []imap.FetchItem{imap.FetchEnvelope}

		if err := c.Fetch(seqset, items, messages); err != nil {
			log.Fatal(err)
		}

		log.Println("Unread messages:")
		for msg := range messages {
			log.Printf("%s — %v", msg.Envelope.Subject, msg.Envelope.From)
		}

		return
	}
	log.Println("No unread messages.")
}
