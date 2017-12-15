package sharding_counter

import (
	"fmt"
	"math/rand"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
)

type counterConfig struct {
	Shards int
}

type shard struct {
	Name  string
	Count int
}

const (
	defaultShards = 20
	configKind    = "NamedCounterShardConfig"
	shardKind     = "NamedCounterShard"
)

func memcacheKey(name string) string {
	return shardKind + ":" + name
}

// Count retrieves the value of the named counter.
func Count(ctx context.Context, name string) (int, error) {
	total := 0
	q := datastore.NewQuery(shardKind).Filter("Name =", name)
	var shards []shard
	_, err := q.GetAll(ctx, &shards)
	if err != nil {
		return total, err
	}
	for _, s := range shards {
		total += s.Count
	}
	return total, nil
}

// Increment increments the named counter.
func Increment(ctx context.Context, name string) error {
	// Get counter config.
	var cfg counterConfig
	ckey := datastore.NewKey(ctx, configKind, name, 0, nil)
	err := datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		err := datastore.Get(ctx, ckey, &cfg)
		if err == datastore.ErrNoSuchEntity {
			cfg.Shards = defaultShards
			_, err = datastore.Put(ctx, ckey, &cfg)
		}
		return err
	}, nil)
	if err != nil {
		return err
	}
	var s shard
	err = datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		shardName := fmt.Sprintf("%s-shard%d", name, rand.Intn(cfg.Shards))
		key := datastore.NewKey(ctx, shardKind, shardName, 0, nil)
		err := datastore.Get(ctx, key, &s)
		// A missing entity and a present entity will both work.
		if err != nil && err != datastore.ErrNoSuchEntity {
			return err
		}
		s.Name = name
		s.Count++
		_, err = datastore.Put(ctx, key, &s)
		return err
	}, nil)
	if err != nil {
		return err
	}
	return nil
}

// IncreaseShards increases the number of shards for the named counter to n.
// It will never decrease the number of shards.
func IncreaseShards(ctx context.Context, name string, n int) error {
	ckey := datastore.NewKey(ctx, configKind, name, 0, nil)
	return datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		var cfg counterConfig
		mod := false
		err := datastore.Get(ctx, ckey, &cfg)
		if err == datastore.ErrNoSuchEntity {
			cfg.Shards = defaultShards
			mod = true
		} else if err != nil {
			return err
		}
		if cfg.Shards < n {
			cfg.Shards = n
			mod = true
		}
		if mod {
			_, err = datastore.Put(ctx, ckey, &cfg)
		}
		return err
	}, nil)
}

func Reset(ctx context.Context, name string) error {
	var shards []shard
	q := datastore.NewQuery(shardKind)

	keys, _ := q.GetAll(ctx, &shards)
	for i, key := range keys {
		s := shards[i]
		s.Count = 0
		err := datastore.RunInTransaction(ctx, func(ctx context.Context) error {
			_, err := datastore.Put(ctx, key, &s)
			return err
		}, nil)
		if err != nil {
			return err
		}
	}
	return nil
}
