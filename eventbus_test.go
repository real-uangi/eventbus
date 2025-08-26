/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/8/26 16:42
 */

// Package eventbus

package eventbus_test

import (
	"fmt"
	"github.com/real-uangi/eventbus"
	"testing"
	"time"
)

var bus eventbus.EventBus

func init() {
	bus = eventbus.NewDefault()
}

func TestBus(t *testing.T) {

	bus.Subscribe("A", subA1)
	bus.Subscribe("A", subA2)
	bus.Subscribe("B", subB1)
	bus.Subscribe("B", subB2)

	if err := bus.Publish("A", "111"); err != nil {
		t.Error(err)
	}
	if err := bus.Publish("B", 222); err != nil {
		t.Error(err)
	}

	time.Sleep(5 * time.Second)
}

func subA1(data interface{}) {
	fmt.Println("subA1", data.(string))
}

func subA2(data interface{}) {
	fmt.Println("subA2", data.(string))
}

func subB1(data interface{}) {
	fmt.Println("subB1", data.(int))
}

func subB2(data interface{}) {
	fmt.Println("subB2", data.(int))
}
