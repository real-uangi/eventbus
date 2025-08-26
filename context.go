/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/8/26 11:01
 */

// Package eventbus

package eventbus

type Context struct {
	Topic   string      `json:"topic"`
	Payload interface{} `json:"payload"`
	handler SubscribeHandler
}

func (context *Context) Clone(handler SubscribeHandler) *Context {
	return &Context{
		Topic:   context.Topic,
		Payload: context.Payload,
		handler: handler,
	}
}

func (context *Context) Execute() {
	if context.handler == nil {
		return
	}
	context.handler(context.Payload)
}
