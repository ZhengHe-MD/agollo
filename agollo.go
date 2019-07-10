package agollo

import (
	"log"
	"os"
)

var (
	defaultClient *Client
	defaultLogger AgolloLogger = log.New(os.Stderr, "", log.LstdFlags)
)

// Start agollo
func Start() error {
	return StartWithConfFile(defaultConfName)
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
	defaultClient = NewClient(conf)

	return defaultClient.Start()
}

// Stop sync config
func Stop() error {
	return defaultClient.Stop()
}

// StartWatchUpdate starts an infinite loop reading changeEvent from update channel
//   and calls HandleChangeEvent method of all observers
func StartWatchUpdate() {
	ceChan := defaultClient.WatchUpdate()

	go func(){
		for {
			ce := <-ceChan

			for _, ob := range defaultClient.getObservers() {
				ob.HandleChangeEvent(ce)
			}
		}
	}()
}

// RegisterObserver registers an observer that will be notified when change event happens
func RegisterObserver(observer ChangeEventObserver) (recall func()) {
	defaultClient.registerObserver(observer)
	return func() {
		defaultClient.recallObserver(observer)
	}
}

// SubscribeToNamespaces fetch namespace config to local and subscribe to updates
func SubscribeToNamespaces(namespaces ...string) error {
	return defaultClient.SubscribeToNamespaces(namespaces...)
}

// GetStringWithNamespace get value from given namespace
func GetStringWithNamespace(namespace, key, defaultValue string) string {
	return defaultClient.GetStringWithNamespace(namespace, key, defaultValue)
}

// GetString from default namespace
func GetString(key, defaultValue string) string {
	return GetStringWithNamespace(defaultNamespace, key, defaultValue)
}

func GetIntWithNamespace(namespace, key string, defaultValue int) int {
	return defaultClient.GetIntWithNamespace(namespace, key, defaultValue)
}

func GetInt(key string, defaultValue int) int {
	return defaultClient.GetInt(key, defaultValue)
}

func GetFloat64WithNamespace(namespace, key string, defaultValue float64) float64 {
	return defaultClient.GetFloat64WithNamespace(namespace, key, defaultValue)
}

func GetFloat64(key string, defaultValue float64) float64 {
	return defaultClient.GetFloat64(key, defaultValue)
}

func GetBoolWithNamespace(namespace, key string, defaultValue bool) bool {
	return defaultClient.GetBoolWithNamespace(namespace, key, defaultValue)
}

func GetBool(key string, defaultValue bool) bool {
	return defaultClient.GetBool(key, defaultValue)
}

// GetNamespaceContent get contents of namespace
func GetNameSpaceContent(namespace, defaultValue string) string {
	return defaultClient.GetNamespaceContent(namespace, defaultValue)
}

// GetAllKeys return all config keys in given namespace
func GetAllKeys(namespace string) []string {
	return defaultClient.GetAllKeys(namespace)
}

// GetReleaseKey return release key for namespace
func GetReleaseKey(namespace string) string {
	return defaultClient.GetReleaseKey(namespace)
}

func SetLogger(logger AgolloLogger) {
	defaultLogger = logger
}
