/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/8/26 14:45
 */

// Package eventbus

package eventbus

import (
	"reflect"
	"runtime"
	"sync"
)

type handlerGroup struct {
	handlers map[string]SubscribeHandler
	mu       sync.Mutex
}

func newHandlerGroup() *handlerGroup {
	return &handlerGroup{
		handlers: make(map[string]SubscribeHandler),
	}
}

func (g *handlerGroup) Add(handler SubscribeHandler) {
	key := funcKey(handler)
	g.mu.Lock()
	defer g.mu.Unlock()
	g.handlers[key] = handler
}

func (g *handlerGroup) Remove(handler SubscribeHandler) {
	removeKey := funcKey(handler)
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.handlers, removeKey)
}

func (g *handlerGroup) Dispatch(context *Context, executorQueue chan *Context) {
	for name, handler := range g.getHandlers() {
		executorQueue <- context.Clone(handler, name)
	}
}

func (g *handlerGroup) getHandlers() map[string]SubscribeHandler {
	g.mu.Lock()
	defer g.mu.Unlock()
	handlers := make(map[string]SubscribeHandler)
	for name, handler := range g.handlers {
		handlers[name] = handler
	}
	return handlers
}

func funcKey(fn interface{}) string {
	v := reflect.ValueOf(fn)
	pc := v.Pointer()
	f := runtime.FuncForPC(pc)
	if f == nil {
		return ""
	}
	return f.Name()
}
