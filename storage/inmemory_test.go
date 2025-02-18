package storage_test

import (
	"testing"

	"github.com/owlmq/owlmq/storage"
)

func TestInMemoryPutAndGet(t *testing.T) {
	sl, err := storage.New(storage.InMemory)
	if err != nil {
		t.Fatalf("Fail to setup test %v", err)
	}
	// test Put
	sl.Put("hello", "world")
	// test Get
	v, err := sl.Get("hello")
	if v != "world" || err != nil {
		t.Fatalf("Storing failed %v", err)
	}
}
