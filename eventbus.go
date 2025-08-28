/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/8/26 10:58
 */

// Package eventbus

package eventbus

import (
	"errors"
	"sync"
	"time"
)

type EventBus interface {
	Subscribe(topic string, handler SubscribeHandler)
	Unsubscribe(topic string, handler SubscribeHandler)
	Publish(topic string, data interface{}) error
}

type bus struct {
	config *Config

	dispatchersQueue chan *Context
	executorQueue    chan *Context

	handlerGroups map[string]*handlerGroup

	mu sync.Mutex
}

type SubscribeHandler func(data interface{})

func New(config *Config) EventBus {
	config.WithDefaults()
	eb := &bus{
		config:           config,
		dispatchersQueue: make(chan *Context, config.DispatcherQueueSize),
		executorQueue:    make(chan *Context, config.ExecutorQueueSize),
		handlerGroups:    make(map[string]*handlerGroup),
	}
	eb.initExecutors()
	eb.initDispatchers()
	return eb
}

func NewDefault() EventBus {
	return New(&Config{})
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
	ctx := &Context{
		Topic:   topic,
		Payload: data,
	}
	select {
	case b.dispatchersQueue <- ctx:
		return nil
	case <-time.After(5 * time.Second):
		return errors.New("timeout")
	}
}

func (b *bus) initDispatchers() {
	for range b.config.Dispatchers {
		go b.newDispatcher()
	}
}

func (b *bus) newDispatcher() {
	for ctx := range b.dispatchersQueue {
		b.getGroup(ctx.Topic).Dispatch(ctx, b.executorQueue)
	}
}

func (b *bus) initExecutors() {
	for range b.config.Executors {
		go b.autoRestartExecutor()
	}
}

func (b *bus) autoRestartExecutor() {
	for {
		b.newExecutor()
	}
}

func (b *bus) newExecutor() {
	defer printPanic()
	for ctx := range b.executorQueue {
		ctx.Execute()
	}
}
