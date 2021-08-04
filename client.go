package lazy

import (
	"context"
	"encoding/json"
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
	jb, err := json.Marshal(baseJob{
		Job:       jobbytes,
		CreatedAT: time.Now(),
		Attempts:  0,
	})
	if err != nil {
		return err
	}

	if err := c.cc.RPush(ctx, queuePrefix+queue, jb).Err(); err != nil {
		return err
	}
	return nil
}

func (c *Client) DoAfter(ctx context.Context, queue string, jobbytes []byte, delay time.Duration) error {
	jb, err := json.Marshal(baseJob{
		Job:       jobbytes,
		CreatedAT: time.Now(),
		Attempts:  0,
	})
	if err != nil {
		return err
	}

	if err := c.cc.ZAdd(ctx, delayPrefix+queue, &redis.Z{
		Score:  float64(time.Now().Add(delay).Unix()),
		Member: jb,
	}).Err(); err != nil {
		return err
	}

	return nil
}
