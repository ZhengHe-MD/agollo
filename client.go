package agollo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/ZhengHe-MD/agollo/v4/parse"
)

type namespaceTyp string

const (
	emptyNamespaceTyp      namespaceTyp = ""
	ymlNamespaceTyp        namespaceTyp = "yml"
	yamlNamespaceTyp       namespaceTyp = "yaml"
	jsonNamespaceTyp       namespaceTyp = "json"
	propertiesNamespaceTyp namespaceTyp = "properties"
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
	mu        sync.RWMutex
}

// result of query config
type result struct {
	// AppID          string            `json:"appId"`
	// Cluster        string            `json:"cluster"`
	NamespaceName  string                 `json:"namespaceName"`
	Configurations map[string]interface{} `json:"configurations"`
	ReleaseKey     string                 `json:"releaseKey"`
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
	var err error
	for _, v := range c.conf.NameSpaceNames {
		if _, e := c.sync(v); e != nil {
			defaultLogger.Printf("module:agollo method:preload namespace:%v, err:%v", v, err)
			if e1 := c.loadLocal(c.getDumpFileName()); e1 != nil {
				err = e1
			}
		}
	}

	return err
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
	val, ok := cache.get(key)
	if !ok {
		return "", false
	}
	strVal, ok := val.(string)
	if !ok {
		return "", false
	}
	return strVal, true
}

func (c *Client) GetString(key string) (string, bool) {
	return c.GetStringWithNamespace(defaultNamespace, key)
}

func (c *Client) GetIntWithNamespace(namespace, key string) (int, bool) {
	cache := c.mustGetCache(namespace)
	val, ok := cache.get(key)
	if !ok {
		return 0, false
	}
	IntVal, ok := val.(int)
	if !ok {
		return 0, false
	}
	return IntVal, true
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

func (c *Client) GetIntSliceWithNamespace(namespace, key string) ([]int, bool) {
	cache := c.mustGetCache(namespace)
	val, ok := cache.get(key)
	if !ok {
		return []int{}, false
	}
	intSliceIfVal, ok := val.([]interface{})
	if !ok {
		return []int{}, false
	}
	intSlices := make([]int, len(intSliceIfVal))
	for idx, intIfVal := range intSliceIfVal {
		intData, ok := intIfVal.(int)
		if !ok {
			defaultLogger.Printf("module:agollo method:GetStringSliceWithNamespace assertion failed")
			return []int{}, false
		}
		intSlices[idx] = intData
	}
	return intSlices, true
}

func (c *Client) GetIntSlice(key string) ([]int, bool) {
	return c.GetIntSliceWithNamespace(defaultNamespace, key)
}

func (c *Client) GetStringSliceWithNamespace(namespace, key string) ([]string, bool) {
	cache := c.mustGetCache(namespace)
	val, ok := cache.get(key)
	if !ok {
		return []string{}, false
	}
	stringSliceIfVal, ok := val.([]interface{})
	if !ok {
		return []string{}, false
	}
	stringSlices := make([]string, len(stringSliceIfVal))
	for idx, stringIfVal := range stringSliceIfVal {
		stringData, ok := stringIfVal.(string)
		if !ok {
			defaultLogger.Printf("module:agollo method:GetStringSliceWithNamespace assertion failed")
			return []string{}, false
		}
		stringSlices[idx] = stringData
	}
	return stringSlices, true
}

func (c *Client) GetStringSlice(key string) ([]string, bool) {
	return c.GetStringSliceWithNamespace(defaultNamespace, key)
}

func (c *Client) GetNamespaceContent(namespace string) (string, bool) {
	namespaceTyp := c.getNameSpaceTyp(namespace)
	return c.GetStringWithNamespace(namespace, string(namespaceTyp)+"content")
}

// 只有文件类型配置可以 Unmarshal, 类似： properties 这种配置类型是 key , value 结构,没有所谓 content 字段，不适合 Unmarshal
func (c *Client) GetNamespaceVal(namespace string, val interface{}) error {
	namespaceTyp := c.getNameSpaceTyp(namespace)
	parser := parse.GetParser(string(namespaceTyp))
	content, ok := c.GetStringWithNamespace(namespace, string(namespaceTyp)+"content")
	if !ok {
		return nil
	}
	if err := parser.Unmarshal([]byte(content), val); err != nil {
		return err
	}
	return nil
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
	defaultLogger.Printf("module:agollo method:Client.sync url:%s data:%s err:%v", url, bts, err)
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
	parser := parse.GetParser(string(c.getNameSpaceTyp(result.NamespaceName)))
	cache := c.mustGetCache(result.NamespaceName)
	kv := cache.dump()

	newConfigurations := c.getConfigurations(parser, result.Configurations)

	for k, v := range kv {
		if _, ok := newConfigurations[k]; !ok {
			cache.delete(k)
			ret.Changes[k] = makeDeleteChange(k, v)
		}
	}

	for k, v := range newConfigurations {
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
	val, ok := c.releaseKeyRepo.get(namespace)
	if !ok {
		return "", false
	}
	strVal, ok := val.(string)
	if !ok {
		return "", false
	}
	return strVal, true
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

func (c *Client) getNameSpaceTyp(namespace string) namespaceTyp {
	if strings.HasSuffix(namespace, ".yml") {
		return ymlNamespaceTyp
	}
	if strings.HasSuffix(namespace, ".yaml") {
		return yamlNamespaceTyp
	}
	if strings.HasSuffix(namespace, ".json") {
		return jsonNamespaceTyp
	}
	if strings.HasSuffix(namespace, ".properties") {
		return propertiesNamespaceTyp
	}
	return emptyNamespaceTyp
}

func (c *Client) getConfigurations(parser parse.ContentParser, configurations map[string]interface{}) map[string]interface{} {
	newConfigurations := make(map[string]interface{})
	for key, val := range configurations {
		tempConfigurations, err := parser.Parse(val)
		if err != nil {
			continue
		}
		if tempConfigurations == nil {
			newConfigurations[key] = val
			continue
		}
		for k, v := range tempConfigurations {
			newConfigurations[k] = v
		}
	}
	content := parser.GetParserType()
	if val, ok := configurations["content"]; ok {
		newConfigurations[content+"content"] = val
	}
	return newConfigurations
}
