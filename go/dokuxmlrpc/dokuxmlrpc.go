// Package dokuxmlrpc is a DokuWiki XML-RPC API client.
package dokuxmlrpc

import (
	"alexejk.io/go-xmlrpc"
)

type Client struct {
	*xmlrpc.Client
}

func NewClient(url string) (*Client, error) {
	c, err := xmlrpc.NewClient(url + "/lib/exe/xmlrpc.php")
	if err != nil {
		return nil, err
	}
	return &Client{c}, nil
}

// Version returns the DokuWiki version.
func (c *Client) Version() (version string, err error) {
	resp := &struct {
		Version string
	}{}
	if err := c.Call("dokuwiki.getVersion", nil, resp); err != nil {
		return "", err
	}
	return resp.Version, nil
}

// Title returns the title of the wiki.
func (c *Client) Title() (title string, err error) {
	resp := &struct {
		Title string
	}{}
	if err := c.Call("dokuwiki.getTitle", nil, resp); err != nil {
		return "", err
	}
	return resp.Title, nil
}
