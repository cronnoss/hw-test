package hw04lrucache

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	t.Run("empty cache", func(t *testing.T) {
		c := NewCache(10)

		_, ok := c.Get("aaa")
		require.False(t, ok)

		_, ok = c.Get("bbb")
		require.False(t, ok)
	})

	t.Run("simple", func(t *testing.T) {
		c := NewCache(5)

		wasInCache := c.Set("aaa", 100)
		require.False(t, wasInCache)

		wasInCache = c.Set("bbb", 200)
		require.False(t, wasInCache)

		val, ok := c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 100, val)

		val, ok = c.Get("bbb")
		require.True(t, ok)
		require.Equal(t, 200, val)

		wasInCache = c.Set("aaa", 300)
		require.True(t, wasInCache)

		val, ok = c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 300, val)

		val, ok = c.Get("ccc")
		require.False(t, ok)
		require.Nil(t, val)
	})

	t.Run("purge logic", func(t *testing.T) {
		c := NewCache(3)

		c.Set("aaa", 100)
		c.Set("bbb", 200)
		c.Set("ccc", 300)

		val, ok := c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 100, val)
		val1, ok := c.Get("bbb")
		require.True(t, ok)
		require.Equal(t, 200, val1)
		val2, ok := c.Get("ccc")
		require.True(t, ok)
		require.Equal(t, 300, val2)

		c.Clear()

		val, ok = c.Get("aaa")
		require.False(t, ok)
		require.Nil(t, val)
		val1, ok = c.Get("bbb")
		require.False(t, ok)
		require.Nil(t, val1)
		val2, ok = c.Get("ccc")
		require.False(t, ok)
		require.Nil(t, val2)
	})
}

func TestCacheMultithreading(t *testing.T) {
	t.Skip() // Remove me if task with asterisk completed.

	c := NewCache(10)
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Set(Key(strconv.Itoa(i)), i)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Get(Key(strconv.Itoa(rand.Intn(1_000_000))))
		}
	}()

	wg.Wait()
}

func TestUpdateValue(t *testing.T) {
	c := NewCache(3)

	c.Set("aaa", 100)
	c.Set("bbb", 200)
	c.Set("ccc", 300)

	c.Set("bbb", 400)

	val, ok := c.Get("bbb")
	require.True(t, ok)
	require.Equal(t, 400, val)
}

func TestNewCache(t *testing.T) {
	if NewCache(5) == nil {
		t.Error("cache should not be nil")
	}
}

func TestPushingOutFirstElementDueToQueueSize(t *testing.T) {
	c := NewCache(3)

	c.Set("key1", 100)
	c.Set("key2", 200)
	c.Set("key3", 300)

	c.Set("key4", 400)

	val, ok := c.Get("key1")
	require.False(t, ok)
	require.Nil(t, val)
}

func TestPushingOutMostOldElement(t *testing.T) {
	c := NewCache(3)

	c.Set("key1", 100)
	c.Set("key2", 200)
	c.Set("key3", 300)

	c.Get("key2")
	c.Get("key1")
	c.Get("key3")
	c.Set("key3", 500)

	c.Set("key4", 400)

	// key2 should be removed, because it was the most old element in cache.
	val, ok := c.Get("key2")
	require.False(t, ok)
	require.Nil(t, val)

	val1, ok := c.Get("key1")
	require.True(t, ok)
	require.Equal(t, 100, val1)
	val2, ok := c.Get("key3")
	require.True(t, ok)
	require.Equal(t, 500, val2)
	val3, ok := c.Get("key4")
	require.True(t, ok)
	require.Equal(t, 400, val3)
}
