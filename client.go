package agollo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
	"sync"
)

// Client for apollo
type Client struct {
	conf *Conf

	updateChan chan *ChangeEvent

	caches         *namespaceCache
	releaseKeyRepo *cache

	longPoller poller
	requester  requester

	ctx    context.Context
	cancel context.CancelFunc

	observers []ChangeEventObserver
	mu sync.RWMutex
}

// result of query config
type result struct {
	// AppID          string            `json:"appId"`
	// Cluster        string            `json:"cluster"`
	NamespaceName  string            `json:"namespaceName"`
	Configurations map[string]string `json:"configurations"`
	ReleaseKey     string            `json:"releaseKey"`
}

// NewClient create client from conf
func NewClient(conf *Conf) *Client {
	client := &Client{
		conf:           conf,
		caches:         newNamespaceCahce(),
		releaseKeyRepo: newCache(),

		requester: newHTTPRequester(&http.Client{Timeout: queryTimeout}),
	}

	client.longPoller = newLongPoller(conf, longPollInterval, client.handleNamespaceUpdate)
	client.ctx, client.cancel = context.WithCancel(context.Background())
	return client
}

// Start sync config
func (c *Client) Start() (err error) {

	// check cache dir
	if err = c.autoCreateCacheDir(); err != nil {
		return err
	}

	// preload all config to local first
	err = c.preload()

	// start fetch update
	go c.longPoller.start()

	return
}

// handleNamespaceUpdate sync config for namespace, delivery changes to subscriber
func (c *Client) handleNamespaceUpdate(namespace string) error {
	change, err := c.sync(namespace)
	if err != nil || change == nil {
		return err
	}

	c.deliveryChangeEvent(change)
	return nil
}

// Stop sync config
func (c *Client) Stop() error {
	c.longPoller.stop()
	c.cancel()
	// close(c.updateChan)
	c.updateChan = nil
	return nil
}

// fetchAllCinfig fetch from remote, if failed load from local file
func (c *Client) preload() error {
	if err := c.longPoller.preload(); err != nil {
		defaultLogger.Printf("[agollo] preload err:%v", err)
		return c.loadLocal(c.getDumpFileName())
	}
	return nil
}

// loadLocal load caches from local file
func (c *Client) loadLocal(name string) error {
	return c.caches.load(name)
}

// dump caches to file
func (c *Client) dump(name string) error {
	return c.caches.dump(name)
}

// WatchUpdate get all updates
func (c *Client) WatchUpdate() <-chan *ChangeEvent {
	if c.updateChan == nil {
		c.updateChan = make(chan *ChangeEvent, 32)
	}
	return c.updateChan
}

func (c *Client) mustGetCache(namespace string) *cache {
	return c.caches.mustGetCache(namespace)
}

// SubscribeToNamespaces fetch namespace config to local and subscribe to updates
func (c *Client) SubscribeToNamespaces(namespaces ...string) error {
	return c.longPoller.addNamespaces(namespaces...)
}

func (c *Client) GetStringWithNamespace(namespace, key string) (string, bool) {
	cache := c.mustGetCache(namespace)
	return cache.get(key)
}

func (c *Client) GetString(key string) (string, bool) {
	return c.GetStringWithNamespace(defaultNamespace, key)
}

func (c *Client) GetIntWithNamespace(namespace, key string) (int, bool) {
	s, ok := c.GetStringWithNamespace(namespace, key)
	if !ok {
		return 0, false
	}

	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	return v, true
}

func (c *Client) GetInt(key string) (int, bool) {
	return c.GetIntWithNamespace(defaultNamespace, key)
}

func (c *Client) GetFloat64WithNamespace(namespace, key string) (float64, bool) {
	s, ok := c.GetStringWithNamespace(namespace, key)
	if !ok {
		return 0, false
	}

	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

func (c *Client) GetFloat64(key string) (float64, bool) {
	return c.GetFloat64WithNamespace(defaultNamespace, key)
}

func (c *Client) GetBoolWithNamespace(namespace, key string) (bool, bool) {
	s, ok := c.GetStringWithNamespace(namespace, key)
	if !ok {
		return false, false
	}

	b, err := strconv.ParseBool(s)
	if err != nil {
		return false, false
	}
	return b, true
}

func (c *Client) GetBool(key string) (bool, bool) {
	return c.GetBoolWithNamespace(defaultNamespace, key)
}

func (c *Client) GetNamespaceContent(namespace string) (string, bool) {
	return c.GetStringWithNamespace(namespace, "content")
}

// GetAllKeys return all config keys in given namespace
func (c *Client) GetAllKeys(namespace string) []string {
	var keys []string
	cache := c.mustGetCache(namespace)
	cache.kv.Range(func(key, value interface{}) bool {
		str, ok := key.(string)
		if ok {
			keys = append(keys, str)
		}
		return true
	})
	return keys
}

// sync namespace config
func (c *Client) sync(namesapce string) (*ChangeEvent, error) {
	releaseKey, _ := c.GetReleaseKey(namesapce)
	url := configURL(c.conf, namesapce, releaseKey)
	bts, err := c.requester.request(url)
	if err != nil || len(bts) == 0 {
		return nil, err
	}
	var result result
	if err := json.Unmarshal(bts, &result); err != nil {
		return nil, err
	}

	return c.handleResult(&result), nil
}

// deliveryChangeEvent push change to subscriber
func (c *Client) deliveryChangeEvent(change *ChangeEvent) {
	if c.updateChan == nil {
		return
	}
	select {
	case <-c.ctx.Done():
	case c.updateChan <- change:
	}
}

// handleResult generate changes from query result, and update local cache
func (c *Client) handleResult(result *result) *ChangeEvent {
	var ret = ChangeEvent{
		Namespace: result.NamespaceName,
		Changes:   map[string]*Change{},
	}

	cache := c.mustGetCache(result.NamespaceName)
	kv := cache.dump()

	for k, v := range kv {
		if _, ok := result.Configurations[k]; !ok {
			cache.delete(k)
			ret.Changes[k] = makeDeleteChange(k, v)
		}
	}

	for k, v := range result.Configurations {
		cache.set(k, v)
		old, ok := kv[k]
		if !ok {
			ret.Changes[k] = makeAddChange(k, v)
			continue
		}
		if old != v {
			ret.Changes[k] = makeModifyChange(k, old, v)
		}
	}

	c.setReleaseKey(result.NamespaceName, result.ReleaseKey)

	// dump caches to file
	c.dump(c.getDumpFileName())

	if len(ret.Changes) == 0 {
		return nil
	}

	return &ret
}

func (c *Client) getDumpFileName() string {
	cacheDir := c.conf.CacheDir
	fileName := fmt.Sprintf(".%s_%s", c.conf.AppID, c.conf.Cluster)
	return path.Join(cacheDir, fileName)
}

// GetReleaseKey return release key for namespace
func (c *Client) GetReleaseKey(namespace string) (string, bool) {
	return c.releaseKeyRepo.get(namespace)
}

func (c *Client) setReleaseKey(namespace, releaseKey string) {
	c.releaseKeyRepo.set(namespace, releaseKey)
}

// autoCreateCacheDir autoCreateCacheDir
func (c *Client) autoCreateCacheDir() error {
	fs, err := os.Stat(c.conf.CacheDir)
	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(c.conf.CacheDir, os.ModePerm)
		}

		return err
	}

	if !fs.IsDir() {
		return fmt.Errorf("conf.CacheDir is not a dir")
	}

	return nil
}

func (c *Client) registerObserver(observer ChangeEventObserver) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.observers = append(c.observers, observer)
}

func (c *Client) recallObserver(ob ChangeEventObserver) {
	var newObservers []ChangeEventObserver
	for _, observer := range c.getObservers() {
		if observer != ob {
			newObservers = append(newObservers, observer)
		}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.observers = newObservers
}

func (c *Client) getObservers() []ChangeEventObserver {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.observers
}