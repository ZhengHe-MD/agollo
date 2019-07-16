package agollo

import (
	"errors"
	"log"
	"os"
)

var (
	defaultAgollo = &Agollo{}
	defaultLogger AgolloLogger = log.New(os.Stderr, "", log.LstdFlags)
)

type Agollo struct {
	Client *Client
}

func NewAgollo(conf *Conf) *Agollo {
	return &Agollo{NewClient(conf)}
}

func (m *Agollo) Start() error {
	return m.Client.Start()
}

func (m *Agollo) StartWithConfFile(name string) error {
	conf, err := NewConf(name)
	if err != nil {
		return err
	}
	return m.StartWithConf(conf)
}

func (m *Agollo) StartWithConf(conf *Conf) error {
	m.Client = NewClient(conf)

	return m.Client.Start()
}

func (m *Agollo) Stop() error {
	return m.Client.Stop()
}

func (m *Agollo) StartWatchUpdate() {
	ceChan := m.Client.WatchUpdate()

	go func() {
		for {
			ce := <-ceChan

			for _, ob := range m.Client.getObservers() {
				ob.HandleChangeEvent(ce)
			}
		}
	}()
}

func (m *Agollo) RegisterObserver(observer ChangeEventObserver) (recall func()) {
	m.Client.registerObserver(observer)
	return func() {
		m.Client.recallObserver(observer)
	}
}

func (m *Agollo) SubscribeToNamespaces(namespaces ...string) error {
	return m.Client.SubscribeToNamespaces(namespaces...)
}

func (m *Agollo) GetStringWithNamespace(namespace, key string) (string, bool) {
	return m.Client.GetStringWithNamespace(namespace, key)
}

func (m *Agollo) GetString(key string) (string, bool) {
	return m.Client.GetString(key)
}

func (m *Agollo) GetIntWithNamespace(namespace, key string) (int, bool) {
	return m.Client.GetIntWithNamespace(namespace, key)
}

func (m *Agollo) GetInt(key string) (int, bool) {
	return m.Client.GetInt(key)
}

func (m *Agollo) GetFloat64WithNamespace(namespace, key string) (float64, bool) {
	return m.Client.GetFloat64WithNamespace(namespace, key)
}

func (m *Agollo) GetFloat64(key string) (float64, bool) {
	return m.Client.GetFloat64(key)
}

func (m *Agollo) GetBoolWithNamespace(namespace, key string) (bool, bool) {
	return m.Client.GetBoolWithNamespace(namespace, key)
}

func (m *Agollo) GetBool(key string) (bool, bool) {
	return m.Client.GetBool(key)
}

func (m *Agollo) GetNameSpaceContent(namespace string) (string, bool) {
	return m.Client.GetNamespaceContent(namespace)
}

func (m *Agollo) GetAllKeys(namespace string) []string {
	return m.Client.GetAllKeys(namespace)
}

func (m *Agollo) GetReleaseKey(namespace string) (string, bool) {
	return m.Client.GetReleaseKey(namespace)
}

// Start agollo [Deprecated]
func Start() error {
	if defaultAgollo.Client == nil {
		return errors.New("please use StartWithConfFile")
	}
	return defaultAgollo.Start()
}

// StartWithConfFile run agollo with conf file
func StartWithConfFile(name string) error {
	conf, err := NewConf(name)
	if err != nil {
		return err
	}
	return StartWithConf(conf)
}

// StartWithConf run agollo with Conf
func StartWithConf(conf *Conf) error {
	return defaultAgollo.StartWithConf(conf)
}

// Stop sync config
func Stop() error {
	return defaultAgollo.Stop()
}

// StartWatchUpdate starts an infinite loop reading changeEvent from update channel
//   and calls HandleChangeEvent method of all observers
func StartWatchUpdate() {
	defaultAgollo.StartWatchUpdate()
}

// RegisterObserver registers an observer that will be notified when change event happens
func RegisterObserver(observer ChangeEventObserver) (recall func()) {
	return defaultAgollo.RegisterObserver(observer)
}

// SubscribeToNamespaces fetch namespace config to local and subscribe to updates
func SubscribeToNamespaces(namespaces ...string) error {
	return defaultAgollo.SubscribeToNamespaces(namespaces...)
}

// GetStringWithNamespace get value from given namespace
func GetStringWithNamespace(namespace, key string) (string, bool) {
	return defaultAgollo.GetStringWithNamespace(namespace, key)
}

// GetString from default namespace
func GetString(key string) (string, bool) {
	return GetStringWithNamespace(defaultNamespace, key)
}

func GetIntWithNamespace(namespace, key string) (int, bool) {
	return defaultAgollo.GetIntWithNamespace(namespace, key)
}

func GetInt(key string) (int, bool) {
	return defaultAgollo.GetInt(key)
}

func GetFloat64WithNamespace(namespace, key string) (float64, bool) {
	return defaultAgollo.GetFloat64WithNamespace(namespace, key)
}

func GetFloat64(key string) (float64, bool) {
	return defaultAgollo.GetFloat64(key)
}

func GetBoolWithNamespace(namespace, key string) (bool, bool) {
	return defaultAgollo.GetBoolWithNamespace(namespace, key)
}

func GetBool(key string) (bool, bool) {
	return defaultAgollo.GetBool(key)
}

// GetNamespaceContent get contents of namespace
func GetNameSpaceContent(namespace string) (string, bool) {
	return defaultAgollo.GetNameSpaceContent(namespace)
}

// GetAllKeys return all config keys in given namespace
func GetAllKeys(namespace string) []string {
	return defaultAgollo.GetAllKeys(namespace)
}

// GetReleaseKey return release key for namespace
func GetReleaseKey(namespace string) (string, bool) {
	return defaultAgollo.GetReleaseKey(namespace)
}

func SetLogger(logger AgolloLogger) {
	defaultLogger = logger
}
