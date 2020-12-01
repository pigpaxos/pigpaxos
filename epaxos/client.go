package epaxos

import (
	"github.com/pigpaxos/pigpaxos"
	"github.com/pigpaxos/pigpaxos/log"
)

// Client overwrites read operation for Paxos
type Client struct {
	*paxi.HTTPClient
	ballot paxi.Ballot
}

func NewClient(id paxi.ID) *Client {
	log.Debugf("Starting EPaxos Client with id %v", id)
	return &Client{
		HTTPClient: paxi.NewHTTPClient(id),
	}
}

func (c *Client) Get(key paxi.Key) (paxi.Value, error) {
	c.CID++
	v, _, err := c.RESTGet(0, key)
	return v, err
}

func (c *Client) Put(key paxi.Key, value paxi.Value) error {
	c.CID++
	_, _, err := c.RESTPut(0, key, value)
	return err
}