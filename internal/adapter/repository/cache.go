package repository

import (
	"context"

	"github.com/runsystemid/gocache"
	"gitlab.com/ayaka/config"
)

type Cache struct {
	*gocache.Redis
	Conf *config.Config `inject:"config"`
}

func (c *Cache) Startup() error {
	RedisConfig := gocache.RedisConfig{
		Mode:     c.Conf.Redis.Mode,
		Address:  c.Conf.Redis.Address,
		Port:     c.Conf.Redis.Port,
		Password: c.Conf.Redis.Password,
	}

	c.Redis = gocache.New(RedisConfig)

	return c.Ping(context.Background())
}

func (c *Cache) Shutdown() error {
	return c.Close()
}
