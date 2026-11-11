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
* **关闭与排空**：`Close(ctx)` 停止新发布、唤醒等待入队的发布者，并等待已入队事件完成处理。
* **易扩展性**：事件处理函数采用 `func(data interface{})`，支持任意类型数据和一对多订阅者模式。

## 发布、顺序与关闭

`Publish` 成功只表示入队，不代表订阅者已执行或外部接收者已收到。事件和队列只在内存中，不提供持久化或自动重试。多分发器或执行器允许并发，不保证处理顺序；两者均设为 `1` 时按发布顺序处理，同一事件多个订阅者的执行次序仍不固定。

`Close(ctx)` 可重复、并发调用。它停止新发布，等待发布者退出，再依次排空分发队列和执行队列。关闭与发布并发时，未完成入队的发布者可能成功入队，也可能收到 `ErrClosed`；成功入队的事件都会排空。超时仅结束本次等待并返回 `ctx.Err()`，后台继续排空，可再次等待；运行中的处理器不会被强制中断，也不应同步等待其所在总线关闭。

关闭后，`Publish` 返回 `ErrClosed`，`ShouldPublish` 记录错误，`MustPublish` panic `ErrClosed`。正常运行时，`Publish` 继续最多等待五秒，`MustPublish` 不设置入队超时。

订阅者按运行时函数名标识，同一 topic 下同名函数、方法或同源闭包会互相替换。分发使用锁内取得的订阅快照，取消订阅不会撤销已取得快照的执行。发布后的数据由调用方保证不再修改，尤其是切片、map 和指针指向的内容。

## 安装

```bash
go get github.com/real-uangi/eventbus
```

## 使用示例

```go
package main

import (
  "context"
  "errors"
  "fmt"
  "github.com/real-uangi/eventbus"
)

func main() {
  bus := eventbus.NewDefault()
  defer bus.Close(context.Background())

  bus.Subscribe("A", subA1)
  bus.Subscribe("A", subA2)
  bus.Subscribe("B", subB1)
  bus.Subscribe("B", subB2)

  bus.Publish("A", "Hello Event A")
  bus.Publish("B", 123)

}

func subA1(data interface{}) error {
  fmt.Println("subA1", data.(string))
  return nil
}

func subA2(data interface{}) error {
  fmt.Println("subA2", data.(string))
  panic(errors.New("test panic"))
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

