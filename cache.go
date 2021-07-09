package agollo

import (
	"encoding/gob"
	"os"
	"sync"
)

type namespaceCache struct {
	lock   sync.Mutex
	caches map[string]*cache
}

func newNamespaceCahce() *namespaceCache {
	return &namespaceCache{
		caches: map[string]*cache{},
	}
}

func (n *namespaceCache) mustGetCache(namespace string) *cache {
	n.lock.Lock()
	defer n.lock.Unlock()

	if ret, ok := n.caches[namespace]; ok {
		return ret
	}

	cache := newCache()
	n.caches[namespace] = cache
	return cache
}

func (n *namespaceCache) drain() {
	n.lock.Lock()
	defer n.lock.Unlock()

	for namespace := range n.caches {
		delete(n.caches, namespace)
	}
}

func (n *namespaceCache) dump(name string) error {
	n.lock.Lock()
	defer n.lock.Unlock()

	var dumps = make(map[string]map[string]interface{})

	for namespace, cache := range n.caches {
		dumps[namespace] = cache.dump()
	}

	f, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	return gob.NewEncoder(f).Encode(&dumps)
}

func (n *namespaceCache) load(name string) error {
	n.drain()

	f, err := os.OpenFile(name, os.O_RDONLY, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	var dumps = make(map[string]map[string]interface{})

	if err := gob.NewDecoder(f).Decode(&dumps); err != nil {
		return err
	}

	for namespace, kv := range dumps {
		cache := n.mustGetCache(namespace)
		for k, v := range kv {
			cache.set(k, v)
		}
	}

	return nil
}

type cache struct {
	kv sync.Map
}

func newCache() *cache {
	return &cache{
		kv: sync.Map{},
	}
}

func (c *cache) set(key string, val interface{}) {
	c.kv.Store(key, val)
}

func (c *cache) get(key string) (interface{}, bool) {
	if val, ok := c.kv.Load(key); ok {
		return val, true
	}
	return "", false
}

func (c *cache) delete(key string) {
	c.kv.Delete(key)
}

func (c *cache) dump() map[string]interface{} {
	var ret = make(map[string]interface{})
	c.kv.Range(func(key, val interface{}) bool {
		k, _ := key.(string)
		ret[k] = val

		return true
	})
	return ret
}
