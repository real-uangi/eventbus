/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/8/26 11:01
 */

// Package eventbus

package eventbus

import "errors"

var NilHandler = errors.New("nil handler")

type Context struct {
	Topic       string      `json:"topic"`
	Payload     interface{} `json:"payload"`
	handler     SubscribeHandler
	handlerName string
}

func (context *Context) Clone(handler SubscribeHandler, name string) *Context {
	return &Context{
		Topic:       context.Topic,
		Payload:     context.Payload,
		handler:     handler,
		handlerName: name,
	}
}

func (context *Context) Execute() error {
	if context.handler == nil {
		return NilHandler
	}
	return context.handler(context.Payload)
}
