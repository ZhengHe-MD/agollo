Note: This is a fork of github.com/philchia/agollo

# agollo is a golang client for apollo 🚀 [![CircleCI](https://circleci.com/gh/ZhengHe-MD/agollo.svg?style=svg)](https://circleci.com/gh/ZhengHe-MD/agollo)

[![Go Report Card](https://goreportcard.com/badge/github.com/ZhengHe-MD/agollo)](https://goreportcard.com/report/github.com/ZhengHe-MD/agollo)
[![Coverage Status](https://coveralls.io/repos/github/ZhengHe-MD/agollo/badge.svg?branch=master)](https://coveralls.io/github/ZhengHe-MD/agollo?branch=master)
[![golang](https://img.shields.io/badge/Language-Go-green.svg?style=flat)](https://golang.org)
[![GoDoc](https://godoc.org/github.com/ZhengHe-MD/agollo?status.svg)](https://godoc.org/github.com/ZhengHe-MD/agollo)
![GitHub release](https://img.shields.io/github/release/ZhengHe-MD/agollo.svg)

## Simple chinese

[简体中文](./README_CN.md)

## Feature

* Multiple namespace support
* Fail tolerant
* Zero dependency
* Realtime change notification with observer pattern
* Customize logger
* Api redesigned in gopher's way -- `val, ok := agollo.GetXXX(key)`

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
