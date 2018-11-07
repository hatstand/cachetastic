package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

type Cache interface {
	Get(key interface{}) (interface{}, error)
}

type CacheLoader func(interface{}) (interface{}, error)

type CacheImpl struct {
	data         sync.Map
	loader       CacheLoader
	refreshAfter time.Duration
	requestGroup singleflight.Group
}

func (c *CacheImpl) load(key interface{}) (interface{}, error) {
	v, err, shared := c.requestGroup.Do(fmt.Sprintf("%v", key), func() (interface{}, error) {
		return c.loader(key)
	})
	log.Printf("Loaded key %s shared: %v", key, shared)
	return v, err
}

func (c *CacheImpl) refresh(key interface{}) {
	v, err := c.load(key)
	if err != nil {
		log.Printf("Failed to refresh value for key %v: %v\n", key, err)
	} else {
		c.data.Store(key, v)
	}
	time.AfterFunc(c.refreshAfter, func() {
		c.refresh(key)
	})
}

func (c *CacheImpl) Get(key interface{}) (interface{}, error) {
	v, ok := c.data.Load(key)
	if ok {
		return v, nil
	}
	v, err := c.load(key)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch value for key %v: %v", key, err)
	}
	c.data.Store(key, v)

	time.AfterFunc(c.refreshAfter, func() {
		c.refresh(key)
	})
	return v, nil
}

func NewCache(loader CacheLoader, refreshAfter time.Duration) Cache {
	return &CacheImpl{
		loader:       loader,
		refreshAfter: refreshAfter,
	}
}

func main() {
	fmt.Println("vim-go")

	cache := NewCache(func(key interface{}) (interface{}, error) {
		log.Printf("Fetching value for key: %v", key)
		time.Sleep(5 * time.Second)
		return 42, nil
	}, time.Second*20)

	go func() {
		v, _ := cache.Get("foo")
		log.Printf("Fetched %v", v)
	}()
	go func() {
		v, _ := cache.Get("foo")
		log.Printf("Fetched %v", v)
	}()
	go func() {
		v, _ := cache.Get("foo")
		log.Printf("Fetched %v", v)
	}()
	go func() {
		v, _ := cache.Get("foo")
		log.Printf("Fetched %v", v)
	}()

	time.Sleep(10 * time.Second)
}
