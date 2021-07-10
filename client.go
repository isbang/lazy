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

func (c *Client) Do(ctx context.Context, queue string, jobstr string) error {
	if err := c.cc.RPush(ctx, queuePrefix+queue, jobstr).Err(); err != nil {
		return err
	}
	return nil
}

func (c *Client) DoAfter(ctx context.Context, queue string, jobstr string, delay time.Duration) error {
	if err := c.cc.ZAdd(ctx, delayPrefix+queue, &redis.Z{
		Score:  float64(time.Now().Add(delay).Unix()),
		Member: jobstr,
	}).Err(); err != nil {
		return err
	}

	return nil
}
