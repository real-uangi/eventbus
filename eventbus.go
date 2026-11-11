/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/8/26 10:58
 */

// Package eventbus

package eventbus

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrClosed = errors.New("event bus is closed")

type EventBus interface {
	Subscribe(topic string, handler SubscribeHandler)
	Unsubscribe(topic string, handler SubscribeHandler)
	Publish(topic string, data interface{}) error
	ShouldPublish(topic string, data interface{})
	MustPublish(topic string, data interface{})
	Close(context.Context) error
}

type bus struct {
	config *Config

	dispatchersQueue chan *Context
	executorQueue    chan *Context

	handlerGroups map[string]*handlerGroup

	mu sync.Mutex

	stateMu     sync.Mutex
	closed      bool
	publishers  sync.WaitGroup
	dispatchers sync.WaitGroup
	executors   sync.WaitGroup
	closing     chan struct{}
	done        chan struct{}
	closeOnce   sync.Once
}

type SubscribeHandler func(data interface{}) error

func New(config *Config) EventBus {
	config.WithDefaults()
	eb := &bus{
		config:           config,
		dispatchersQueue: make(chan *Context, config.DispatcherQueueSize),
		executorQueue:    make(chan *Context, config.ExecutorQueueSize),
		handlerGroups:    make(map[string]*handlerGroup),
		closing:          make(chan struct{}),
		done:             make(chan struct{}),
	}
	eb.initExecutors()
	eb.initDispatchers()
	return eb
}

func NewDefault() EventBus {
	return New(&Config{})
}

func WithLogger(customLogger Logger) {
	logger = customLogger
}

func (b *bus) getGroup(topic string) *handlerGroup {
	b.mu.Lock()
	defer b.mu.Unlock()
	group, ok := b.handlerGroups[topic]
	if !ok {
		group = newHandlerGroup()
		b.handlerGroups[topic] = group
	}
	return group
}

func (b *bus) Subscribe(topic string, handler SubscribeHandler) {
	b.getGroup(topic).Add(handler)
}

func (b *bus) Unsubscribe(topic string, handler SubscribeHandler) {
	b.getGroup(topic).Remove(handler)
}

func (b *bus) Publish(topic string, data interface{}) error {
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	return b.publish(topic, data, timer.C)
}

func (b *bus) publish(topic string, data interface{}, timeout <-chan time.Time) error {
	// Admission and the shutdown boundary share a lock. Close can safely wait
	// for every admitted publisher before closing the dispatcher queue.
	b.stateMu.Lock()
	if b.closed {
		b.stateMu.Unlock()
		return ErrClosed
	}
	b.publishers.Add(1)
	b.stateMu.Unlock()
	defer b.publishers.Done()
	ctx := &Context{
		Topic:   topic,
		Payload: data,
	}
	select {
	case b.dispatchersQueue <- ctx:
		return nil
	case <-b.closing:
		return ErrClosed
	case <-timeout:
		return errors.New("timeout")
	}
}

func (b *bus) ShouldPublish(topic string, data interface{}) {
	err := b.Publish(topic, data)
	if err != nil {
		logger.Warnf("failed to publish event [%s]: %v", topic, err)
	}
}

func (b *bus) MustPublish(topic string, data interface{}) {
	if err := b.publish(topic, data, nil); err != nil {
		panic(err)
	}
}

// Close rejects new publications and drains accepted events. A timeout only
// bounds this caller's wait: running handlers finish and draining continues.
func (b *bus) Close(ctx context.Context) error {
	b.closeOnce.Do(func() {
		b.stateMu.Lock()
		b.closed = true
		close(b.closing)
		b.stateMu.Unlock()
		go func() {
			b.publishers.Wait()
			close(b.dispatchersQueue)
			b.dispatchers.Wait()
			close(b.executorQueue)
			b.executors.Wait()
			close(b.done)
		}()
	})
	select {
	case <-b.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (b *bus) initDispatchers() {
	for range b.config.Dispatchers {
		b.dispatchers.Add(1)
		go b.newDispatcher()
	}
	logger.Infof("initialized %d dispatchers with queue size %d", b.config.Dispatchers, b.config.DispatcherQueueSize)
}

func (b *bus) newDispatcher() {
	defer b.dispatchers.Done()
	for ctx := range b.dispatchersQueue {
		b.getGroup(ctx.Topic).Dispatch(ctx, b.executorQueue)
	}
}

func (b *bus) initExecutors() {
	for range b.config.Executors {
		b.executors.Add(1)
		go b.newExecutor()
	}
	logger.Infof("initialized %d executors with queue size %d", b.config.Executors, b.config.ExecutorQueueSize)
}

func (b *bus) newExecutor() {
	defer b.executors.Done()
	for ctx := range b.executorQueue {
		b.handlerExecute(ctx)
	}
}

func (b *bus) handlerExecute(ctx *Context) {
	defer b.printPanic(ctx)
	err := ctx.Execute()
	if err != nil {
		logger.Warnf("error occurs on topic[%s] handler[%s]: %v\n%s", ctx.Topic, ctx.handlerName, err)
	}
}
