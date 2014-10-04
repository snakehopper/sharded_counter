package sharded_counter

import (
	"appengine"
	"appengine/datastore"
	"appengine/memcache"
	"fmt"
	"math/rand"
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
	configKind    = "GeneralCounterShardConfig"
	shardKind     = "GeneralCounterShard"
)

func memcacheKey(name string) string {
	return shardKind + ":" + name
}

// Count retrieves the value of the named counter.
func Count(c appengine.Context, name string) (int, error) {
	total := 0
	mkey := memcacheKey(name)
	if _, err := memcache.JSON.Get(c, mkey, &total); err == nil {
		return total, nil
	}
	pKey := datastore.NewKey(c, configKind, name, 0, nil)
	q := datastore.NewQuery(shardKind).Ancestor(pKey).Filter("Name =", name)
	for t := q.Run(c); ; {
		var s shard
		_, err := t.Next(&s)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return total, err
		}
		total += s.Count
	}
	memcache.JSON.Set(c, &memcache.Item{
		Key:    mkey,
		Object: &total,
	})
	return total, nil
}

// Increment increments the named counter.
func Increment(c appengine.Context, name string) error {
	// Get counter config.
	var cfg counterConfig
	ckey := datastore.NewKey(c, configKind, name, 0, nil)
	err := datastore.RunInTransaction(c, func(c appengine.Context) error {
		err := datastore.Get(c, ckey, &cfg)
		if err == datastore.ErrNoSuchEntity {
			cfg.Shards = defaultShards
			_, err = datastore.Put(c, ckey, &cfg)
		}
		return err
	}, nil)
	if err != nil {
		return err
	}

	err = datastore.RunInTransaction(c, func(c appengine.Context) error {
		shardName := fmt.Sprintf("shard%d", rand.Intn(cfg.Shards))
		key := datastore.NewKey(c, shardKind, shardName, 0, ckey)
		var s shard
		err := datastore.Get(c, key, &s)
		// A missing entity and a present entity will both work.
		if err != nil && err != datastore.ErrNoSuchEntity {
			return err
		}
		s.Count++
		s.Name = name
		_, err = datastore.Put(c, key, &s)
		return err
	}, nil)
	if err != nil {
		return err
	}

	memcache.IncrementExisting(c, memcacheKey(name), 1)
	return nil
}

// IncreaseShards increases the number of shards for the named counter to n.
// It will never decrease the number of shards.
func IncreaseShards(c appengine.Context, name string, n int) error {
	ckey := datastore.NewKey(c, configKind, name, 0, nil)
	return datastore.RunInTransaction(c, func(c appengine.Context) error {
		var cfg counterConfig
		mod := false
		err := datastore.Get(c, ckey, &cfg)
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
			_, err = datastore.Put(c, ckey, &cfg)
		}
		return err
	}, nil)
}
