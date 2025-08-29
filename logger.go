/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/8/29 10:50
 */

// Package eventbus

package eventbus

import (
	"fmt"
	"log"
)

type Logger interface {
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
}

var logger Logger = DefaultLogger()

type defaultLogger struct{}

func (l *defaultLogger) Infof(format string, args ...interface{}) {
	log.Println(fmt.Sprintf(format, args...))
}

func (l *defaultLogger) Warnf(format string, args ...interface{}) {
	log.Println(fmt.Sprintf(format, args...))
}

func DefaultLogger() Logger {
	return &defaultLogger{}
}
