// vim: foldmethod=marker

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Command mastobot is a Mastodon bot that posts random text generated by Markov
// chain algorithm.
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/mattn/go-mastodon"
)

func main() {
	log.SetFlags(0)
	rand.Seed(time.Now().UnixNano())

	var (
		numWords  = flag.Int("words", 100, "Maximum number of words to generate.")
		prefixLen = flag.Int("prefix", 2, "Prefix length in words.")
		file      = flag.String("file", "", "File to build chains from.")
		server    = flag.String("server", "", "Mastodon server `URL`.")
		token     = flag.String("token", "", "Mastodon access token.")
	)
	flag.Parse()

	if *file == "" || *server == "" || *token == "" {
		log.Fatal("-file, -server and -token flags must be set.")
	}

	f, err := os.Open(*file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// Build a Markov chain in memory.
	c := newChain(*prefixLen)
	c.build(f)

	mc := mastodon.NewClient(&mastodon.Config{
		AccessToken: *token,
		Server:      *server,
	})
	s, err := mc.PostStatus(context.Background(), &mastodon.Toot{
		Status: c.generate(*numWords),
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Successfully shitposted: %v", s)
}

// Markov chain algorithm (adapted from go.dev/doc/codewalk/markov/) {{{

// prefix is a Markov chain prefix of one or more words.
type prefix []string

// String returns the prefix as a string (for use as a map key) and implements
// the fmt.Stringer interface.
func (p prefix) String() string {
	return strings.Join(p, " ")
}

// shift removes the first word from the prefix and appends the given word.
func (p prefix) shift(word string) {
	copy(p, p[1:])
	p[len(p)-1] = word
}

// chain contains a map ("chain") of prefixes to a list of suffixes. A prefix is
// a string of prefixLen words joined with spaces. A suffix is a single word. A
// prefix can have multiple suffixes.
type chain struct {
	chain     map[string][]string
	prefixLen int
}

// newChain returns a new chain with prefixes of prefixLen words.
func newChain(prefixLen int) *chain {
	return &chain{make(map[string][]string), prefixLen}
}

// build reads text from the provided io.Reader and parses it into prefixes and
// suffixes that are stored in chain.
func (c *chain) build(r io.Reader) {
	br := bufio.NewReader(r)
	p := make(prefix, c.prefixLen)
	for {
		var s string
		if _, err := fmt.Fscan(br, &s); err != nil {
			break
		}
		key := p.String()
		c.chain[key] = append(c.chain[key], s)
		p.shift(s)
	}
}

// generate returns a string of at most n words generated from chain.
func (c *chain) generate(n int) string {
	p := make(prefix, c.prefixLen)
	var words []string
	for i := 0; i < n; i++ {
		choices := c.chain[p.String()]
		if len(choices) == 0 {
			break
		}
		next := choices[rand.Intn(len(choices))]
		words = append(words, next)
		p.shift(next)
	}
	return strings.Join(words, " ")
}

// }}}
