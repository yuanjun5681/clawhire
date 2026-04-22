package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

type Client struct {
	raw *mongo.Client
	db  *mongo.Database
}

func NewClient(ctx context.Context, uri, database string) (*Client, error) {
	opt := options.Client().ApplyURI(uri)

	c, err := mongo.Connect(opt)
	if err != nil {
		return nil, fmt.Errorf("mongo connect: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := c.Ping(pingCtx, readpref.Primary()); err != nil {
		_ = c.Disconnect(pingCtx)
		return nil, fmt.Errorf("mongo ping: %w", err)
	}

	return &Client{raw: c, db: c.Database(database)}, nil
}

func (c *Client) DB() *mongo.Database { return c.db }
func (c *Client) Raw() *mongo.Client  { return c.raw }

func (c *Client) Close(ctx context.Context) error {
	return c.raw.Disconnect(ctx)
}

func (c *Client) Ping(ctx context.Context) error {
	return c.raw.Ping(ctx, readpref.Primary())
}
