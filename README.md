# EventBus

轻量级 Go 事件总线（EventBus）库，用于实现一对多事件订阅和发布机制。

## 特性

* **订阅与取消订阅**：通过 `Subscribe(topic, handler)` 和 `Unsubscribe(topic, handler)` 管理事件订阅。
* **事件发布**：通过 `Publish(topic, data)` 发布事件，支持异步队列处理，带 5 秒超时保护。
* **异步处理机制**：

    * `dispatchersQueue` 分发事件给订阅者组。
    * `executorQueue` 执行订阅者处理函数。
    * 利用 Goroutine 并发处理，提高效率。
* **线程安全**：使用 `sync.Mutex` 保护 `handlerGroups`，保证多线程环境下安全。
* **自动重启执行器**：执行器在发生 panic 时自动重启，确保事件处理不中断。
* **可配置性**：通过 `Config` 配置队列大小和并发数量，支持快速初始化默认配置的 `NewDefault()`。
* **易扩展性**：事件处理函数采用 `func(data interface{})`，支持任意类型数据和一对多订阅者模式。

## 安装

```bash
go get github.com/real-uangi/eventbus
```

## 使用示例

```go
package main

import (
  "errors"
  "fmt"
  "github.com/real-uangi/eventbus"
  "time"
)

func main() {
  bus := eventbus.NewDefault()

  bus.Subscribe("A", subA1)
  bus.Subscribe("A", subA2)
  bus.Subscribe("B", subB1)
  bus.Subscribe("B", subB2)

  bus.Publish("A", "Hello Event A")
  bus.Publish("B", 123)

  time.Sleep(2 * time.Second)
}

func subA1(data interface{}) error {
  fmt.Println("subA1", data.(string))
  return nil
}

func subA2(data interface{}) error {
  fmt.Println("subA2", data.(string))
  panic(errors.New("test panic"))
  return nil
}

func subB1(data interface{}) error {
  fmt.Println("subB1", data.(int))
  return nil
}

func subB2(data interface{}) error {
  fmt.Println("subB2", data.(int))
  return errors.New("test error")
}
```

