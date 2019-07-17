本项目 fork 自 github.com/philchia/agollo

[english version can be found here](./README_EN.md)

# apollo 客户端 🚀 [![CircleCI](https://circleci.com/gh/ZhengHe-MD/agollo.svg?style=svg)](https://circleci.com/gh/ZhengHe-MD/agollo)

[![Go Report Card](https://goreportcard.com/badge/github.com/ZhengHe-MD/agollo)](https://goreportcard.com/report/github.com/ZhengHe-MD/agollo)
[![Coverage Status](https://coveralls.io/repos/github/ZhengHe-MD/agollo/badge.svg?branch=master)](https://coveralls.io/github/ZhengHe-MD/agollo?branch=master)
[![golang](https://img.shields.io/badge/Language-Go-green.svg?style=flat)](https://golang.org)
[![GoDoc](https://godoc.org/github.com/ZhengHe-MD/agollo?status.svg)](https://godoc.org/github.com/ZhengHe-MD/agollo)
![GitHub release](https://img.shields.io/github/release/ZhengHe-MD/agollo.svg)

## 主要变化

##### 1. 依照 go 习惯重新设计 api

原项目暴露的 api 沿用了 Java 的设计习惯：

```go
val := agollo.GetString(key, defaultVal)
```

这种设计的问题在于：

* 我们必须在调用时提供默认值，但在 go 语言中，我们有零值 (zero value)，而无需考虑 null
* 我们无法确定 key (如 groupA.item) 是否存在。假如想要在 apollo 中设置 fallback 值，比如 groupDefault.item，我们将因为无法判断 key 是否存在而无法决定是否使用 fallback 值

因此，我们修改这种设计：

```go
val, ok := agollo.GetString(key)
```

##### 2. 多实例支持

原项目使用了单例模式，即整个进程中只有一个唯一的 agollo 客户端实例（defaultClient），所有请求都必须通过这个实例来发送。然而，有时候我们需要同时访问多个 app 的配置信息，如 middleware 和 serviceA，而我们不希望 serviceA 的开发者可以控制 middleware 的配置，这时候就需要多实例支持：

```go
// this will use a different client instance
ag := agollo.NewAgollo(conf)
if err := ag.Start(); err != nil {
  // ...
}
ag.GetString(key)
```

##### 3. 利用 observer pattern 支持配置更新监听

原项目提供 WatchUpdate 方法，调用它返回一个只读的配置变化事件 channel，应用可以从这个 channel 消费到配置变化事件，从而实现热更新。但问题在于，这个 channel 里的每个事件只会被消费一次，如果有多个 goroutines 在消费它，那么很可能出现错过重要更新的问题。于是，我们决定在这里利用 observer pattern，每个 goroutine 都可以通过订阅的方式来监听所有配置变化事件：

```go
type simpleObserver struct {}
func (s *simpleObserver) HandleChangeEvent(event *ChangeEvent) {
  // consume the event
}
ag.RegisterObserver(&simpleObserver{})
ag.StartWatchUpdate()
```

##### 4. 支持定制化 Logger

当我们想要在已有的基础设施中融合 agollo 时，有时候需要看到 agollo 内部的日志信息，并按已有的方式打印、记录日志，这时候，你的 Logger 只需要实现下面的接口：

```go
type AgolloLogger interface {
	Printf(format string, v ...interface{})
}
```

你就可以通过 SetLogger 来配置 Logger

```go
agollo.SetLogger(logger)
```

##### 5. 更多的 config getters 支持

我们增加了更多的 getters:

```go
GetString(key)
GetInt(key)
GetBool(key)
GetFloat64(key)
```

## 功能

* 多 namespace 支持
* 容错，本地缓存
* 零依赖
* 配置变化事件订阅
* 自定义 Logger
* 符合 go 习惯的 api
* 多实例支持

## 依赖

**go 1.9** 或更新

## 安装

```sh
$ go get -u github.com/ZhengHe-MD/agollo/v4
```

## 使用

#### Hello world 例子

```go
import "github.com/ZhengHe-MD/agollo/v4"

func main() {
  conf := &agollo.Conf{
    AppID:          "SampleApp",
    Cluster:        "default",
    NameSpaceNames: []string{"application"},
    CacheDir:       "/tmp/agollo",
    IP:             "localhost:8080", 
  }
  err := agollo.StartWithConf(conf)
  if err != nil {
    log.Println(err)
  }
  
  stringVal, ok := agollo.GetString("k1")
  if !ok {
    sv = "defaultV1"
  }
  
  intVal, ok := agollo.GetInt("k2")
  boolVal, ok := agollo.GetBool("k3")
}
```

#### 查询不同的 Namespaces

```go
import "github.com/ZhengHe-MD/agollo/v4"

func main() {
  conf := &agollo.Conf{
    AppID:          "SampleApp",
    Cluster:        "default",
    NameSpaceNames: []string{"application", "middleware"},
    CacheDir:       "/tmp/agollo",
    IP:             "localhost:8080", 
  }
  
  err := agollo.StartWithConf(conf)
  // ...
  stringVal, ok := agollo.GetStringWithNamespace("middleware", "k1")
  // ...
}
```

#### 监听配置更新

```go
import "github.com/ZhengHe-MD/agollo/v4"

type observer struct {}
func (m *observer) HandleChangeEvent(ce *ChangeEvent) {
    // deal with change event
}

func main() {
  // ... start agollo
  recall := agollo.Register(&observer{})
  // this will unregister the observer
  defer recall()
}
```

#### 获取配置

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

#### 订阅新的 namespace 配置

```golang
agollo.SubscribeToNamespaces("newNamespace1", "newNamespace2")
```

#### 自定义 logger

```golang
agollo.SetLogger(logger)
```

## 许可

agollo 使用 MIT 许可
