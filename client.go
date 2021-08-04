package lazy

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

func NewClient(cc redis.UniversalClient) *Client {
	return &Client{
		cc: cc,
	}
}

type Client struct {
	cc redis.UniversalClient
}

func (c *Client) Do(ctx context.Context, queue string, jobbytes []byte) error {
	if err := c.cc.RPush(ctx, queuePrefix+queue, string(jobbytes)).Err(); err != nil {
		return err
	}
	return nil
}

func (c *Client) DoAfter(ctx context.Context, queue string, jobbytes []byte, delay time.Duration) error {
	if err := c.cc.ZAdd(ctx, delayPrefix+queue, &redis.Z{
		Score:  float64(time.Now().Add(delay).Unix()),
		Member: string(jobbytes),
	}).Err(); err != nil {
		return err
	}

	return nil
}
