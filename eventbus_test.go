/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/8/26 16:42
 */

// Package eventbus

package eventbus_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/real-uangi/eventbus"
	"runtime"
	"testing"
)

func TestBus(t *testing.T) {
	bus := eventbus.NewDefault()
	t.Cleanup(func() { _ = bus.Close(context.Background()) })

	bus.Subscribe("A", subA1)
	bus.Subscribe("A", subA2)
	bus.Subscribe("B", subB1)
	bus.Subscribe("B", subB2)

	t.Log(runtime.NumGoroutine())

	if err := bus.Publish("A", "111"); err != nil {
		t.Error(err)
	}

	t.Log(runtime.NumGoroutine())

	if err := bus.Publish("B", 222); err != nil {
		t.Error(err)
	}

	if err := bus.Publish("C", 555); err != nil {
		t.Error(err)
	}

	t.Log(runtime.NumGoroutine())

	if err := bus.Close(context.Background()); err != nil {
		t.Fatal(err)
	}

	t.Log(runtime.NumGoroutine())
}

func subA1(data interface{}) error {
	fmt.Println("subA1", data.(string))
	return nil
}

func subA2(data interface{}) error {
	fmt.Println("subA2", data.(string))
	panic(errors.New("test panic"))
}

func subB1(data interface{}) error {
	fmt.Println("subB1", data.(int))
	return nil
}

func subB2(data interface{}) error {
	fmt.Println("subB2", data.(int))
	return errors.New("test error")
}
