Note: This is a fork of github.com/philchia/agollo

# agollo is a golang client for apollo 🚀 [![CircleCI](https://circleci.com/gh/ZhengHe-MD/agollo.svg?style=svg)](https://circleci.com/gh/ZhengHe-MD/agollo)

[![Go Report Card](https://goreportcard.com/badge/github.com/ZhengHe-MD/agollo)](https://goreportcard.com/report/github.com/ZhengHe-MD/agollo)
[![Coverage Status](https://coveralls.io/repos/github/ZhengHe-MD/agollo/badge.svg?branch=master)](https://coveralls.io/github/ZhengHe-MD/agollo?branch=master)
[![golang](https://img.shields.io/badge/Language-Go-green.svg?style=flat)](https://golang.org)
[![GoDoc](https://godoc.org/github.com/ZhengHe-MD/agollo?status.svg)](https://godoc.org/github.com/ZhengHe-MD/agollo)
![GitHub release](https://img.shields.io/github/release/ZhengHe-MD/agollo.svg)

## Simple chinese

[简体中文](./README_CN.md)

## Main difference of this fork

##### 1. redesign the api in gopher's way

before, the agollo module followed the Java's way of api design:

```go
val := agollo.GetString(key, defaultVal)
```

the problem is that:

1. we're forced to provide a default value, which is awkward when using golang, we have default value instead of null
2. we can't decide whether the value exists or not. Let's say we have fallback config in apollo, it's impossible to decide whether or not to use fallback config.

so it's necessary to follow the gopher's way:

```go
val, ok := agollo.GetString(key)
```

##### 2. multiple instances support

before, the agollo module implements a singleton agollo client, called **defaultClient**, all subsequent requests are sent throught this client. however, sometimes we need to visit different apps' configs in the same process, for example, the **middleware** app and the **serviceA** app, since we don't want the developers  of **serviceA** to have control over the general settings of **middleware**. therefore, it's necessary to support multiple agollo client instances, while keeping the defaultClient working as before at the same time.

```go
// this will use a different client instance
ag := agollo.NewAgollo(conf)
if err := ag.Start(); err != nil {
  // ...
}
ag.GetString(key)
```

##### 3. support observer pattern for hot config updates

before, the agollo module provides a **WatchUpdate** method that returns a read-only **ChangeEvent** channel for application to listen on. However, the problem is that there can be only one goroutine consuming events, if multiple goroutines simultaneously reading from the same channel, some important updates can be missed. So we decide to implement an observer pattern, to support multiple goroutines consuming every change event, just like subscriptions.

```go
type simpleObserver struct {}
func (s *simpleObserver) HandleChangeEvent(event *ChangeEvent) {
  // consume the event
}
ag.RegisterObserver(&simpleObserver{})
ag.StartWatchUpdate()
```

##### 4. support customized logger

when you want to integrate agollo into a large infrastructure, we may want logs from agollo print in a consistent way, as long as your logger implement the following interface:

```go
type AgolloLogger interface {
	Printf(format string, v ...interface{})
}
```

##### 5. more config getters support

we add some useful config getters to deal with different data types:

```go
GetString(key)
GetInt(key)
GetBool(key)
GetFloat64(key)
```

## Feature

* Multiple namespace support
* Fail tolerant
* Zero dependency
* Realtime change notification with observer pattern
* Customized logger support
* gopher's way of api design
* Multiple instances support

## Required

**go 1.9** or later

## Installation

```sh
$ go get -u github.com/ZhengHe-MD/agollo
```

## Usage

### Start use default app.properties config file

```golang
agollo.Start()
```

### Start use given config file path

```golang
agollo.StartWithConf(&agollo.Conf{
    AppID:          "SampleApp",
    Cluster:        "default",
    NameSpaceNames: []string{"application"},
    CacheDir:       "/tmp/agollo",
    IP:             "localhost:8080", 
})
agollo.StartWithConfFile(name)
```

### Observe Updates

```golang
type observer struct {}
func (m *observer) HandleChangeEvent(ce *ChangeEvent) {
    fmt.Println(ce)
}

recall := agollo.Register(&observer{})
defer recall()

// ...
```

### Get apollo values

```golang
agollo.GetString(key)
agollo.GetStringWithNamespace(namespace, key)
agollo.GetInt(key)
agollo.GetIntWithNamespace(namespace, key)
agollo.GetBool(key)
agollo.GetBoolWithNamespace(namespace, key)
agollo.GetFloat64(key)
agollo.GetFloat64WithNamespace(namespace, key)
```

### Get namespace file contents

```golang
agollo.GetNamespaceContent(namespace)
```

### Get all keys

```golang
agollo.GetAllKeys(namespace)
```

### Subscribe to new namespaces

```golang
agollo.SubscribeToNamespaces("newNamespace1", "newNamespace2")
```

### Set Logger

any logger that satisfies AgolloLogger interface

```golang
type AgolloLogger interface {
	Printf(format string, v ...interface{})
}
```

can be used inside agollo

```golang
agollo.SetLogger(logger)
```

## License

agollo is released under MIT license
