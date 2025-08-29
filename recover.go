/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/8/26 16:34
 */

// Package eventbus

package eventbus

import "github.com/real-uangi/eventbus/trace"

func (b *bus) printPanic(ctx *Context) {
	if r := recover(); r != nil {
		logger.Warnf("panic on topic[%s] handler[%s]: %v\n%s", ctx.Topic, ctx.handlerName, r, trace.Stack(3))
	}
}
