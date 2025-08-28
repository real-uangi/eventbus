/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/8/26 15:55
 */

// Package eventbus

package eventbus

import "runtime"

type Config struct {
	Dispatchers         int
	DispatcherQueueSize int
	Executors           int
	ExecutorQueueSize   int
}

func (config *Config) WithDefaults() {
	if config.Dispatchers == 0 {
		config.Dispatchers = runtime.NumCPU()
	}
	if config.DispatcherQueueSize == 0 {
		config.DispatcherQueueSize = 8 * config.Dispatchers
	}
	if config.Executors == 0 {
		config.Executors = 2 * runtime.NumCPU()
	}
	if config.ExecutorQueueSize == 0 {
		config.ExecutorQueueSize = 8 * config.Executors
	}
}
