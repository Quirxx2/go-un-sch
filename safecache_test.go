package golangunitedschoolcerts

import (
	"math"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
)

// Check that struct implements interface
var _ Cache[int, testValue] = &SafeCache[int, testValue]{}

func concurrentlyAdd(t *testing.T, capacity int, onEviction EvictionCallback[int, testValue], vNum int, pauseAfter int) (cache *SafeCache[int, testValue], complete *sync.WaitGroup, pause *sync.WaitGroup, keys []int) {
	c, err := NewLRUCache(capacity, onEviction)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	cache = NewSafeCache[int, testValue](c)
	if vNum < pauseAfter {
		assert.FailNow(t, "can't pause after non existent key")
	}
	complete = new(sync.WaitGroup)
	complete.Add(vNum)
	pause = new(sync.WaitGroup)
	if pauseAfter != -1 {
		pause.Add(pauseAfter)
	}
	keys = make([]int, 0)
	for i := 0; i < vNum; i++ {
		keys = append(keys, i)
	}
	go func() {
		for i := 0; i < vNum; i++ {
			go func(key int) {
				defer complete.Done()
				if pauseAfter != -1 && key >= pauseAfter {
					pause.Wait()
				}
				cache.Add(key, testValue{key})
				if pauseAfter != -1 && key < pauseAfter {
					defer pause.Done()
				}
			}(i)
		}
	}()
	return
}

func testSafeCacheCallback(t *testing.T, expKeys []int, expValues []int) EvictionCallback[int, testValue] {
	return func(key *int, value *testValue) {
		assert.True(t, slices.Contains(expKeys, *key))
		assert.True(t, slices.Contains(expValues, value.v))
	}
}

func Test_SafeCache_Add(t *testing.T) {
	t.Run("Adding elements from many goroutines", func(t *testing.T) {
		vNum := 1000
		cache, complete, _, expKeys := concurrentlyAdd(t, vNum*testValue{}.Size(), nil, vNum, -1)
		complete.Wait()
		assert.Equal(t, vNum, cache.Len())
		assert.Equal(t, vNum*testValue{}.Size(), cache.Size())
		assert.ElementsMatch(t, expKeys, cache.Keys())
		for i := 0; i < vNum; i++ {
			got, ok := cache.Get(i)
			assert.True(t, ok)
			assert.Equal(t, &testValue{i}, got)
		}
	})
	t.Run("Replacing existing keys from many goroutines", func(t *testing.T) {
		vNum := 1000
		existing := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
		newV := testValue{math.MaxInt}
		cb := func(key *int, value *testValue) {
			assert.True(t, slices.Contains(existing, *key))
			assert.Equal(t, *key, value.v)
		}
		cache, complete, pause, expKeys := concurrentlyAdd(t, vNum*testValue{}.Size(), cb, vNum, len(existing)*2)
		pause.Wait()
		wg := new(sync.WaitGroup)
		for i := 0; i < len(existing); i++ {
			wg.Add(1)
			go func(key int) {
				defer wg.Done()
				cache.Add(key, newV)
				got, ok := cache.Get(key)
				assert.True(t, ok)
				assert.Equal(t, &newV, got)
				assert.False(t, slices.Contains(cache.Keys()[:len(existing)], key))
			}(existing[i])
		}
		wg.Wait()
		complete.Wait()
		assert.Equal(t, vNum, cache.Len())
		assert.Equal(t, vNum*testValue{}.Size(), cache.Size())
		assert.ElementsMatch(t, expKeys, cache.Keys())
	})
	t.Run("Overflowing capacity while adding", func(t *testing.T) {
		vNum := 1000
		cb := func(key *int, value *testValue) {
			assert.True(t, *key < vNum/2, value.v < vNum/2)
		}
		cache, complete, pause, _ := concurrentlyAdd(t, vNum, cb, vNum, vNum/2)
		pause.Wait()
		complete.Wait()
		assert.Equal(t, vNum/2, cache.Len())
		assert.Equal(t, vNum/2*testValue{}.Size(), cache.Size())
		for i := vNum / 2; i < vNum; i++ {
			got, ok := cache.Get(i)
			assert.True(t, ok)
			assert.Equal(t, &testValue{i}, got)
		}
	})
}

func Test_SafeCache_Contains(t *testing.T) {
	vNum := 1000
	tData := map[string]struct {
		keys []int
		exp  bool
	}{
		"Existing keys from many goroutines":    {[]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, true},
		"Nonexisting keys from many goroutines": {[]int{-vNum, vNum, vNum * 10, -vNum / 2, vNum * 100, vNum * 1000}, false},
	}

	for name, tCase := range tData {
		t.Run(name, func(t *testing.T) {
			cache, complete, pause, expKeys := concurrentlyAdd(t, vNum*testValue{}.Size(), nil, vNum, len(tCase.keys))
			pause.Wait()
			wg := new(sync.WaitGroup)
			for i := 0; i < len(tCase.keys); i++ {
				wg.Add(1)
				go func(key int) {
					defer wg.Done()
					assert.Equal(t, tCase.exp, cache.Contains(key))
				}(tCase.keys[i])
			}
			wg.Wait()
			complete.Wait()
			assert.Equal(t, vNum, cache.Len())
			assert.Equal(t, vNum*testValue{}.Size(), cache.Size())
			assert.ElementsMatch(t, expKeys, cache.Keys())
		})
	}
}

func Test_SafeCache_Get(t *testing.T) {
	vNum := 1000
	tData := map[string]struct {
		keys []int
		exp  bool
	}{
		"Existing keys from many goroutines":    {[]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, true},
		"Nonexisting keys from many goroutines": {[]int{-vNum, vNum, vNum * 10, -vNum / 2, vNum * 100, vNum * 1000}, false},
	}

	for name, tCase := range tData {
		t.Run(name, func(t *testing.T) {
			cache, complete, pause, expKeys := concurrentlyAdd(t, vNum*testValue{}.Size(), nil, vNum, len(tCase.keys))
			pause.Wait()
			wg := new(sync.WaitGroup)
			for i := 0; i < len(tCase.keys); i++ {
				wg.Add(1)
				go func(key int) {
					defer wg.Done()
					for cache.Len() < len(tCase.keys)*2 {
						time.Sleep(1 * time.Nanosecond)
					}
					got, ok := cache.Get(key)
					assert.Equal(t, tCase.exp, ok)
					if tCase.exp {
						assert.Equal(t, &testValue{key}, got)
					}
					assert.False(t, slices.Contains(cache.Keys()[:len(tCase.keys)], key))
				}(tCase.keys[i])
			}
			wg.Wait()
			complete.Wait()
			assert.Equal(t, vNum, cache.Len())
			assert.Equal(t, vNum*testValue{}.Size(), cache.Size())
			assert.ElementsMatch(t, expKeys, cache.Keys())
		})
	}
}

func Test_SafeCache_Remove(t *testing.T) {
	vNum := 1000
	tData := map[string]struct {
		keys     []int
		expCount int
		callback bool
	}{
		"Existing keys, not nil callback":    {[]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, vNum - 11, true},
		"Existing keys, nil callback":        {[]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, vNum - 11, false},
		"Nonexisting keys, not nil callback": {[]int{-vNum, vNum, vNum * 10, -vNum / 2, vNum * 100, vNum * 1000}, vNum, true},
		"Nonexisting keys, nil callback":     {[]int{-vNum, vNum, vNum * 10, -vNum / 2, vNum * 100, vNum * 1000}, vNum, false},
	}

	for name, tCase := range tData {
		t.Run(name, func(t *testing.T) {
			var cb EvictionCallback[int, testValue] = nil
			if tCase.callback {
				cb = testSafeCacheCallback(t, tCase.keys, tCase.keys)
			}
			cache, complete, pause, _ := concurrentlyAdd(t, vNum*testValue{}.Size(), cb, vNum, len(tCase.keys))
			pause.Wait()
			wg := new(sync.WaitGroup)
			for _, k := range tCase.keys {
				wg.Add(1)
				go func(key int) {
					defer wg.Done()
					cache.Remove(key)
					assert.False(t, slices.Contains(cache.Keys(), key))
					assert.False(t, cache.Contains(key))
				}(k)
			}
			wg.Wait()
			complete.Wait()
			assert.Equal(t, tCase.expCount, cache.Len())
			assert.Equal(t, tCase.expCount*testValue{}.Size(), cache.Size())
		})
	}
}

func Test_SafeCache_RemoveOldest(t *testing.T) {
	n := 1000
	tData := map[string]struct {
		vNum     int
		expKeys  []int
		expCount int
		callback bool
	}{
		"Existing keys, not nil callback": {n, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, n - 11, true},
		"Existing keys, nil callback":     {n, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, n - 11, false},
		"Empty cache, not nil callback":   {0, []int{}, 0, true},
		"Empty cache, nil callback":       {0, []int{}, 0, false},
	}

	for name, tCase := range tData {
		t.Run(name, func(t *testing.T) {
			var cb EvictionCallback[int, testValue] = nil
			if tCase.callback {
				cb = testSafeCacheCallback(t, tCase.expKeys, tCase.expKeys)
			}
			cache, complete, pause, _ := concurrentlyAdd(t, tCase.vNum*testValue{}.Size(), cb, tCase.vNum, len(tCase.expKeys))
			pause.Wait()
			wg := new(sync.WaitGroup)
			for _, k := range tCase.expKeys {
				wg.Add(1)
				go func(key int) {
					defer wg.Done()
					cache.RemoveOldest()
				}(k)
			}
			wg.Wait()
			complete.Wait()
			left := cache.Keys()
			for _, k := range tCase.expKeys {
				assert.False(t, slices.Contains(left, k))
			}
			assert.Equal(t, tCase.expCount, cache.Len())
			assert.Equal(t, tCase.expCount*testValue{}.Size(), cache.Size())
		})
	}
}

func Test_SafeCache_Keys(t *testing.T) {
	tData := map[string]struct {
		vNum int
	}{
		"Not empty cache": {1000},
		"Empty cache":     {0},
	}

	for name, tCase := range tData {
		t.Run(name, func(t *testing.T) {
			cache, complete, _, expKeys := concurrentlyAdd(t, tCase.vNum*testValue{}.Size(), nil, tCase.vNum, -1)
			complete.Wait()
			assert.ElementsMatch(t, expKeys, cache.Keys())
		})
	}

}

func Test_SafeCache_Size(t *testing.T) {
	tData := map[string]struct {
		vNum     int
		capacity int
		expSize  int
	}{
		"Not empty cache, no eviction":  {1000, 1000 * testValue{}.Size(), 1000 * testValue{}.Size()},
		"Not empty cache, half evicted": {1000, 1000 * testValue{}.Size() / 2, 1000},
		"Empty cache":                   {0, 0, 0},
	}

	for name, tCase := range tData {
		t.Run(name, func(t *testing.T) {
			cache, complete, _, _ := concurrentlyAdd(t, tCase.capacity, nil, tCase.vNum, -1)
			complete.Wait()
			assert.Equal(t, tCase.expSize, cache.Size())
		})
	}
}

func Test_SafeCache_Len(t *testing.T) {
	tData := map[string]struct {
		vNum     int
		capacity int
		expLen   int
	}{
		"Not empty cache, no eviction":  {1000, 1000 * testValue{}.Size(), 1000},
		"Not empty cache, half evicted": {1000, 1000 * testValue{}.Size() / 2, 500},
		"Empty cache":                   {0, 0, 0},
	}

	for name, tCase := range tData {
		t.Run(name, func(t *testing.T) {
			cache, complete, _, _ := concurrentlyAdd(t, tCase.capacity, nil, tCase.vNum, -1)
			complete.Wait()
			assert.Equal(t, tCase.expLen, cache.Len())
		})
	}
}

func Test_SafeCache_Capacity(t *testing.T) {
	tData := map[string]struct {
		vNum     int
		capacity int
	}{
		"Not empty cache": {1000, 1000 * testValue{}.Size()},
		"Empty cache":     {0, 0},
	}

	for name, tCase := range tData {
		t.Run(name, func(t *testing.T) {
			cache, complete, _, _ := concurrentlyAdd(t, tCase.capacity, nil, tCase.vNum, -1)
			complete.Wait()
			assert.Equal(t, tCase.capacity, cache.Capacity())
		})
	}
}

func Test_SafeCache_Resize(t *testing.T) {
	tData := map[string]struct {
		vNum     int
		capacity int
		callback bool
	}{
		"Expand cache, not nil callback":  {1000, 1000 * 2 * testValue{}.Size(), true},
		"Expand cache, nil callback":      {1000, 1000 * 2 * testValue{}.Size(), false},
		"Shrink cache, not nil callback":  {1000, 1000 / 2 * testValue{}.Size(), true},
		"Shrink cache, nil callback":      {1000, 1000 / 2 * testValue{}.Size(), false},
		"Same capacity, not nil callback": {1000, 1000 * testValue{}.Size(), true},
		"Same capacity, nil callback":     {1000, 1000 * testValue{}.Size(), false},
	}

	for name, tCase := range tData {
		t.Run(name, func(t *testing.T) {
			evicted := int32(0)
			pauseAfter := tCase.vNum / 2
			var countEvictions EvictionCallback[int, testValue] = nil
			if tCase.callback {
				countEvictions = func(key *int, value *testValue) {
					atomic.AddInt32(&evicted, 1)
					assert.True(t, *key < pauseAfter)
				}
			}
			capacity := tCase.vNum * testValue{}.Size()
			cache, complete, pause, _ := concurrentlyAdd(t, capacity, countEvictions, tCase.vNum, pauseAfter)
			pause.Wait()
			cache.Resize(tCase.capacity)
			complete.Wait()
			assert.Equal(t, tCase.capacity, cache.Capacity())
			diff := capacity - tCase.capacity
			if diff > 0 && tCase.callback {
				assert.Equal(t, int(evicted), diff/testValue{}.Size())
			} else {
				assert.Zero(t, evicted)
			}
		})
	}
}

func Test_SafeCache_Purge(t *testing.T) {
	tData := map[string]struct {
		vNum     int
		callback bool
	}{
		"Not empty cache, not nil callback": {1000, true},
		"Not empty cache, nil callback":     {1000, false},
		"Empty cache, not nill callback":    {0, true},
		"Empty cache, nill callback":        {0, false},
	}

	for name, tCase := range tData {
		t.Run(name, func(t *testing.T) {
			evicted := int32(0)
			var countEvictions EvictionCallback[int, testValue] = nil
			if tCase.callback {
				countEvictions = func(key *int, value *testValue) {
					atomic.AddInt32(&evicted, 1)
				}
			}
			capacity := tCase.vNum * testValue{}.Size()
			cache, complete, _, _ := concurrentlyAdd(t, capacity, countEvictions, tCase.vNum, -1)
			complete.Wait()
			cache.Purge()
			assert.Zero(t, cache.Len())
			assert.Zero(t, cache.Size())
			assert.Empty(t, cache.Keys())
			if tCase.callback {
				assert.Equal(t, int(evicted), tCase.vNum)
			} else {
				assert.Zero(t, evicted)
			}
		})
	}
}

func Test_SafeCache_Peek(t *testing.T) {
	vNum := 1000
	tData := map[string]struct {
		keys []int
		exp  bool
	}{
		"Existing keys from many goroutines":    {[]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, true},
		"Nonexisting keys from many goroutines": {[]int{-vNum, vNum, vNum * 10, -vNum / 2, vNum * 100, vNum * 1000}, false},
	}

	for name, tCase := range tData {
		t.Run(name, func(t *testing.T) {
			cache, complete, pause, expKeys := concurrentlyAdd(t, vNum*testValue{}.Size(), nil, vNum, len(tCase.keys))
			pause.Wait()
			wg := new(sync.WaitGroup)
			for i := 0; i < len(tCase.keys); i++ {
				wg.Add(1)
				go func(key int) {
					defer wg.Done()
					got, ok := cache.Peek(key)
					assert.Equal(t, tCase.exp, ok)
					if tCase.exp {
						assert.Equal(t, &testValue{key}, got)
						assert.True(t, slices.Contains(cache.Keys()[:len(tCase.keys)], key))
					}
				}(tCase.keys[i])
			}
			wg.Wait()
			complete.Wait()
			assert.Equal(t, vNum, cache.Len())
			assert.Equal(t, vNum*testValue{}.Size(), cache.Size())
			assert.ElementsMatch(t, expKeys, cache.Keys())
		})
	}
}

func Test_SafeCache_Touch(t *testing.T) {
	vNum := 1000
	tData := map[string]struct {
		keys []int
	}{
		"Existing keys from many goroutines":    {[]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
		"Nonexisting keys from many goroutines": {[]int{-vNum, vNum, vNum * 10, -vNum / 2, vNum * 100, vNum * 1000}},
	}

	for name, tCase := range tData {
		t.Run(name, func(t *testing.T) {
			cache, complete, pause, expKeys := concurrentlyAdd(t, vNum*testValue{}.Size(), nil, vNum, len(tCase.keys))
			pause.Wait()
			wg := new(sync.WaitGroup)
			for i := 0; i < len(tCase.keys); i++ {
				wg.Add(1)
				go func(key int) {
					defer wg.Done()
					for cache.Len() < len(tCase.keys)*2 {
						time.Sleep(1 * time.Nanosecond)
					}
					cache.Touch(key)
					assert.False(t, slices.Contains(cache.Keys()[:len(tCase.keys)], key))
				}(tCase.keys[i])
			}
			wg.Wait()
			complete.Wait()
			assert.Equal(t, vNum, cache.Len())
			assert.Equal(t, vNum*testValue{}.Size(), cache.Size())
			assert.ElementsMatch(t, expKeys, cache.Keys())
		})
	}
}
