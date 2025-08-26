/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/8/26 15:55
 */

// Package eventbus

package eventbus

type Config struct {
	Dispatchers         int
	DispatcherQueueSize int
	Executors           int
	ExecutorQueueSize   int
}

func (config *Config) WithDefaults() {
	if config.Dispatchers == 0 {
		config.Dispatchers = 2
	}
	if config.DispatcherQueueSize == 0 {
		config.DispatcherQueueSize = 16
	}
	if config.Executors == 0 {
		config.Executors = 8
	}
	if config.ExecutorQueueSize == 0 {
		config.ExecutorQueueSize = config.Executors * 8
	}
}
