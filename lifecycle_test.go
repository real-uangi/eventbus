package eventbus

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func closeTestBus(t *testing.T, b EventBus) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := b.Close(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestCloseDrainsAcceptedEventsInOrder(t *testing.T) {
	b := New(&Config{Dispatchers: 1, Executors: 1, DispatcherQueueSize: 2, ExecutorQueueSize: 2})
	t.Cleanup(func() { closeTestBus(t, b) })
	var got []int
	b.Subscribe("ordered", func(value interface{}) error {
		got = append(got, value.(int))
		return nil
	})
	for i := range 20 {
		if err := b.Publish("ordered", i); err != nil {
			t.Fatal(err)
		}
	}
	closeTestBus(t, b)
	if len(got) != 20 {
		t.Fatalf("handled %d accepted events, want 20", len(got))
	}
	for i, value := range got {
		if value != i {
			t.Fatalf("event %d = %d", i, value)
		}
	}
	if err := b.Publish("ordered", 21); !errors.Is(err, ErrClosed) {
		t.Fatalf("publish after close: %v", err)
	}
	b.ShouldPublish("ordered", 22)
	defer func() {
		if err, ok := recover().(error); !ok || !errors.Is(err, ErrClosed) {
			t.Fatalf("MustPublish after close must panic with ErrClosed")
		}
	}()
	b.MustPublish("ordered", 23)
}

func TestCloseWakesPublisherAndContinuesAfterTimeout(t *testing.T) {
	b := New(&Config{Dispatchers: 1, Executors: 1, DispatcherQueueSize: 1, ExecutorQueueSize: 1}).(*bus)
	started, release := make(chan struct{}), make(chan struct{})
	var count atomic.Int64
	b.Subscribe("blocked", func(interface{}) error {
		if count.Add(1) == 1 {
			close(started)
			<-release
		}
		return nil
	})
	t.Cleanup(func() { closeTestBus(t, b) })
	t.Cleanup(func() { close(release) })
	if err := b.Publish("blocked", 1); err != nil {
		t.Fatal(err)
	}
	<-started
	for _, n := range []int{2, 3, 4} {
		if err := b.Publish("blocked", n); err != nil {
			t.Fatal(err)
		}
	}
	// One executing handler, one queued execution, one blocked dispatcher,
	// and one queued dispatch leave no capacity for another publication.
	deadline := time.Now().Add(3 * time.Second)
	for len(b.dispatchersQueue) != 1 || len(b.executorQueue) != 1 {
		if time.Now().After(deadline) {
			t.Fatal("queues did not saturate")
		}
		runtime.Gosched()
	}
	result := make(chan error, 1)
	go func() { result <- b.Publish("blocked", 5) }()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := b.Close(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("close while handler is blocked: %v", err)
	}
	select {
	case err := <-result:
		if !errors.Is(err, ErrClosed) {
			t.Fatalf("blocked publication: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("shutdown did not wake blocked publisher")
	}
	// Cleanup releases the running handler before waiting for final draining.
}

func TestConcurrentSubscriptionAndPublication(t *testing.T) {
	b := New(&Config{Dispatchers: 2, Executors: 2})
	t.Cleanup(func() { closeTestBus(t, b) })
	handler := func(interface{}) error { return nil }
	var workers sync.WaitGroup
	for range 4 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for range 200 {
				b.Subscribe("concurrent", handler)
				b.MustPublish("concurrent", nil)
				b.Unsubscribe("concurrent", handler)
			}
		}()
	}
	workers.Wait()
}

func TestCloseRacesWithPublishWithoutLosingAcceptedEvents(t *testing.T) {
	b := New(&Config{Dispatchers: 2, Executors: 2})
	t.Cleanup(func() { closeTestBus(t, b) })
	var accepted, handled atomic.Int64
	started := make(chan struct{}, 1)
	b.Subscribe("race-close", func(interface{}) error {
		handled.Add(1)
		select {
		case started <- struct{}{}:
		default:
		}
		return nil
	})
	var workers sync.WaitGroup
	for range 8 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for range 100 {
				err := b.Publish("race-close", nil)
				if errors.Is(err, ErrClosed) {
					return
				}
				if err != nil {
					t.Error(err)
					return
				}
				accepted.Add(1)
			}
		}()
	}
	<-started
	var closers sync.WaitGroup
	for range 4 {
		closers.Add(1)
		go func() {
			defer closers.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			if err := b.Close(ctx); err != nil {
				t.Error(err)
			}
		}()
	}
	closers.Wait()
	workers.Wait()
	if accepted.Load() != handled.Load() {
		t.Fatalf("accepted=%d handled=%d", accepted.Load(), handled.Load())
	}
}
