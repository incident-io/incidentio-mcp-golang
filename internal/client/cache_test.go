package client

import (
	"testing"
	"time"
)

func TestCache_SetAndGet(t *testing.T) {
	cache := NewCache(1 * time.Second)

	// Test setting and getting a value
	cache.Set("key1", "value1")
	value, found := cache.Get("key1")
	if !found {
		t.Error("Expected to find key1 in cache")
	}
	if value != "value1" {
		t.Errorf("Expected 'value1', got %v", value)
	}
}

func TestCache_Expiration(t *testing.T) {
	cache := NewCache(100 * time.Millisecond)

	// Set a value
	cache.Set("key1", "value1")

	// Should be found immediately
	_, found := cache.Get("key1")
	if !found {
		t.Error("Expected to find key1 in cache")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should not be found after expiration
	_, found = cache.Get("key1")
	if found {
		t.Error("Expected key1 to be expired")
	}
}

func TestCache_Delete(t *testing.T) {
	cache := NewCache(1 * time.Second)

	cache.Set("key1", "value1")
	cache.Delete("key1")

	_, found := cache.Get("key1")
	if found {
		t.Error("Expected key1 to be deleted")
	}
}

func TestCache_Clear(t *testing.T) {
	cache := NewCache(1 * time.Second)

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Clear()

	_, found1 := cache.Get("key1")
	_, found2 := cache.Get("key2")

	if found1 || found2 {
		t.Error("Expected all keys to be cleared")
	}
}

func TestCache_CleanExpired(t *testing.T) {
	cache := NewCache(100 * time.Millisecond)

	// Set multiple values
	cache.Set("key1", "value1")
	time.Sleep(50 * time.Millisecond)
	cache.Set("key2", "value2")

	// Wait for first key to expire
	time.Sleep(60 * time.Millisecond)

	// Clean expired entries
	cache.CleanExpired()

	// key1 should be gone, key2 should still exist
	_, found1 := cache.Get("key1")
	_, found2 := cache.Get("key2")

	if found1 {
		t.Error("Expected key1 to be cleaned")
	}
	if !found2 {
		t.Error("Expected key2 to still exist")
	}
}

func TestCache_ConcurrentAccess(t *testing.T) {
	cache := NewCache(1 * time.Second)

	// Test concurrent writes and reads
	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			cache.Set("key", i)
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			cache.Get("key")
		}
		done <- true
	}()

	// Wait for both to complete
	<-done
	<-done

	// Should not panic or race
}

func TestCache_DifferentTypes(t *testing.T) {
	cache := NewCache(1 * time.Second)

	// Test with different types
	cache.Set("string", "value")
	cache.Set("int", 42)
	cache.Set("struct", struct{ Name string }{"test"})

	str, _ := cache.Get("string")
	if str != "value" {
		t.Error("String value mismatch")
	}

	num, _ := cache.Get("int")
	if num != 42 {
		t.Error("Int value mismatch")
	}

	s, _ := cache.Get("struct")
	if s.(struct{ Name string }).Name != "test" {
		t.Error("Struct value mismatch")
	}
}

// Made with Bob
