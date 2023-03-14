package golangunitedschoolcerts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Check that struct implements interface
var _ Cache[int, testValue] = &LRUCache[int, testValue]{}

type testValue struct {
	v int
}

func (testValue) Size() int {
	return 2
}

func cacheRangeOfInts(t *testing.T, count int, onEviction EvictionCallback[int, testValue]) (Cache[int, testValue], []int) {
	capacity := count * testValue{}.Size()
	cache, err := NewLRUCache(capacity, onEviction)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	keys := make([]int, 0)
	for i := 0; i < count; i++ {
		v := testValue{i}
		cache.Add(i, v)
		keys = append(keys, i)
	}
	return cache, keys
}

func testCallback(t *testing.T, expKey int, expValue testValue) EvictionCallback[int, testValue] {
	return func(key *int, value *testValue) {
		assert.Equal(t, expKey, *key)
		assert.Equal(t, expValue, *value)
	}
}

func collectCallback(keys *[]int, values *[]int) EvictionCallback[int, testValue] {
	return func(key *int, value *testValue) {
		*keys = append(*keys, *key)
		*values = append(*values, value.v)
	}
}

func Test_LRUCache_Add(t *testing.T) {
	vNum := 10
	expSize := vNum * testValue{}.Size()
	t.Run("Add values with different keys", func(t *testing.T) {
		cache, keys := cacheRangeOfInts(t, vNum, nil)
		assert.Equal(t, vNum, cache.Len())
		assert.Equal(t, expSize, cache.Size())
		assert.Equal(t, keys, cache.Keys())
	})
	t.Run("Replacing existing key", func(t *testing.T) {
		key := vNum / 2
		newV := testValue{999}
		cache, keys := cacheRangeOfInts(t, vNum, testCallback(t, key, testValue{key}))
		cache.Add(key, newV)
		assert.Equal(t, vNum, cache.Len())
		assert.Equal(t, expSize, cache.Size())
		var changedKeys []int
		changedKeys = append(changedKeys, keys[:key]...)
		changedKeys = append(changedKeys, keys[key+1:]...)
		changedKeys = append(changedKeys, key)
		assert.Equal(t, changedKeys, cache.Keys())
		got, ok := cache.Get(key)
		assert.True(t, ok)
		assert.Equal(t, &newV, got)
	})
	t.Run("Overflowing capacity while adding", func(t *testing.T) {
		evictedKeys := make([]int, 0)
		evictedValues := make([]int, 0)
		cache, keys := cacheRangeOfInts(t, vNum, collectCallback(&evictedKeys, &evictedValues))
		expKeys := make([]int, 0)
		for i := vNum; i < vNum*2; i++ {
			expKeys = append(expKeys, i)
			cache.Add(i, testValue{i})
		}
		assert.Equal(t, len(expKeys), cache.Len())
		assert.Equal(t, len(expKeys)*testValue{}.Size(), cache.Size())
		assert.Equal(t, expKeys, cache.Keys())
		assert.Equal(t, keys, evictedKeys)
		assert.Equal(t, keys, evictedValues)
	})
}

func Test_LRUCache_Contains(t *testing.T) {
	vNum := 10
	cache, _ := cacheRangeOfInts(t, vNum, nil)
	tData := map[string]struct {
		key int
		exp bool
	}{
		"Existing key":    {vNum / 2, true},
		"Nonexisting key": {-vNum, false},
	}

	for name, tCase := range tData {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tCase.exp, cache.Contains(tCase.key))
		})
	}
}

func Test_LRUCache_Get(t *testing.T) {
	vNum := 4
	tData := map[string]struct {
		key     int
		expV    *testValue
		expOk   bool
		expKeys []int
	}{
		"Existing key":    {2, &testValue{2}, true, []int{0, 1, 3, 2}},
		"Nonexisting key": {-vNum, nil, false, []int{0, 1, 2, 3}},
	}

	for name, tCase := range tData {
		t.Run(name, func(t *testing.T) {
			cache, _ := cacheRangeOfInts(t, vNum, nil)
			got, ok := cache.Get(tCase.key)
			assert.Equal(t, tCase.expV, got)
			assert.Equal(t, tCase.expOk, ok)
			assert.Equal(t, tCase.expKeys, cache.Keys())
		})
	}
}

func Test_LRUCache_Remove(t *testing.T) {
	vNum := 4
	tData := map[string]struct {
		key      int
		expCount int
		expKeys  []int
		callback func(*testing.T, int, testValue) EvictionCallback[int, testValue]
	}{
		"Existing key, not nil callback":    {2, 3, []int{0, 1, 3}, testCallback},
		"Existing key, nil callback":        {3, 3, []int{0, 1, 2}, nil},
		"Nonexisting key, not nil callback": {vNum * 2, 4, []int{0, 1, 2, 3}, testCallback},
		"Nonexisting key, nil callback":     {-vNum, 4, []int{0, 1, 2, 3}, nil},
	}

	for name, tCase := range tData {
		t.Run(name, func(t *testing.T) {
			var cb EvictionCallback[int, testValue] = nil
			if tCase.callback != nil {
				cb = tCase.callback(t, tCase.key, testValue{tCase.key})
			}
			cache, _ := cacheRangeOfInts(t, vNum, cb)
			cache.Remove(tCase.key)
			assert.Equal(t, tCase.expCount, cache.Len())
			assert.Equal(t, tCase.expCount*testValue{}.Size(), cache.Size())
			assert.Equal(t, tCase.expKeys, cache.Keys())
		})
	}
}

func Test_LRUCache_RemoveOldest(t *testing.T) {
	tData := map[string]struct {
		vNum     int
		expKeys  []int
		callback func(*testing.T, int, testValue) EvictionCallback[int, testValue]
	}{
		"Not empty cache, not nil callback": {4, []int{1, 2, 3}, testCallback},
		"Not empty cache, nil callback":     {4, []int{1, 2, 3}, nil},
		"Empty cache, not nil callback":     {0, []int{}, testCallback},
		"Empty cache, nil callback":         {0, []int{}, nil},
	}

	for name, tCase := range tData {
		t.Run(name, func(t *testing.T) {
			var cb EvictionCallback[int, testValue] = nil
			if tCase.callback != nil {
				cb = tCase.callback(t, 0, testValue{0})
			}
			cache, _ := cacheRangeOfInts(t, tCase.vNum, cb)
			cache.RemoveOldest()
			assert.Equal(t, tCase.expKeys, cache.Keys())
		})
	}
}

func Test_LRUCache_Keys(t *testing.T) {
	tData := map[string]struct {
		vNum    int
		expKeys []int
	}{
		"Not empty cache": {4, []int{0, 1, 2, 3}},
		"Empty cache":     {0, []int{}},
	}

	for name, tCase := range tData {
		t.Run(name, func(t *testing.T) {
			cache, _ := cacheRangeOfInts(t, tCase.vNum, nil)
			assert.Equal(t, tCase.expKeys, cache.Keys())
		})
	}
}

func Test_LRUCache_Size(t *testing.T) {
	tData := map[string]struct {
		vNum int
	}{
		"Not empty cache": {100},
		"Empty cache":     {0},
	}

	for name, tCase := range tData {
		t.Run(name, func(t *testing.T) {
			cache, _ := cacheRangeOfInts(t, tCase.vNum, nil)
			assert.Equal(t, tCase.vNum*testValue{}.Size(), cache.Size())
		})
	}
}

func Test_LRUCache_Len(t *testing.T) {
	tData := map[string]struct {
		vNum int
	}{
		"Not empty cache": {100},
		"Empty cache":     {0},
	}

	for name, tCase := range tData {
		t.Run(name, func(t *testing.T) {
			cache, _ := cacheRangeOfInts(t, tCase.vNum, nil)
			assert.Equal(t, tCase.vNum, cache.Len())
		})
	}
}

func Test_LRUCache_Capacity(t *testing.T) {
	tData := map[string]struct {
		vNum int
	}{
		"Not empty cache": {100},
		"Empty cache":     {0},
	}

	for name, tCase := range tData {
		t.Run(name, func(t *testing.T) {
			cache, _ := cacheRangeOfInts(t, tCase.vNum, nil)
			assert.Equal(t, tCase.vNum*testValue{}.Size(), cache.Capacity())
		})
	}
}

func Test_LRUCache_Purge(t *testing.T) {
	tData := map[string]struct {
		vNum    int
		collect func(*[]int, *[]int) EvictionCallback[int, testValue]
	}{
		"Not empty cache, not nil callback": {10, collectCallback},
		"Not empty cache, nil callback":     {100, nil},
		"Empty cache, not nill callback":    {0, collectCallback},
		"Empty cache, nill callback":        {0, nil},
	}

	for name, tCase := range tData {
		t.Run(name, func(t *testing.T) {
			var cb EvictionCallback[int, testValue]
			evictedKeys := make([]int, 0)
			evictedValues := make([]int, 0)
			if tCase.collect != nil {
				cb = tCase.collect(&evictedKeys, &evictedValues)
			}
			cache, keys := cacheRangeOfInts(t, tCase.vNum, cb)
			cache.Purge()
			if tCase.collect != nil {
				assert.Equal(t, keys, evictedKeys)
				assert.Equal(t, keys, evictedValues)
			}
			assert.Zero(t, cache.Size())
			assert.Zero(t, cache.Len())
			assert.Empty(t, cache.Keys())
		})
	}
}

func Test_LRUCache_Resize(t *testing.T) {
	tData := map[string]struct {
		vNum     int
		capacity int
		collect  func(*[]int, *[]int) EvictionCallback[int, testValue]
	}{
		"Expand cache, not nil callback":  {10, 20 * testValue{}.Size(), collectCallback},
		"Expand cache, nil callback":      {10, 20 * testValue{}.Size(), nil},
		"Shrink cache, not nil callback":  {10, 5 * testValue{}.Size(), collectCallback},
		"Shrink cache, nil callback":      {10, 5 * testValue{}.Size(), nil},
		"Same capacity, not nil callback": {10, 10 * testValue{}.Size(), collectCallback},
		"Same capacity, nil callback":     {10, 10 * testValue{}.Size(), nil},
	}

	for name, tCase := range tData {
		t.Run(name, func(t *testing.T) {
			var cb EvictionCallback[int, testValue]
			evictedKeys := make([]int, 0)
			evictedValues := make([]int, 0)
			if tCase.collect != nil {
				cb = tCase.collect(&evictedKeys, &evictedValues)
			}
			cache, keys := cacheRangeOfInts(t, tCase.vNum, cb)
			diff := cache.Capacity() - tCase.capacity
			if diff > 0 && tCase.collect != nil {
				keys = keys[:diff/testValue{}.Size()]
			} else {
				keys = []int{}
			}
			cache.Resize(tCase.capacity)
			assert.Equal(t, tCase.capacity, cache.Capacity())
			assert.Equal(t, keys, evictedKeys)
			assert.Equal(t, keys, evictedValues)
		})
	}
}

func Test_LRUCache_Peek(t *testing.T) {
	vNum := 4
	tData := map[string]struct {
		key     int
		expV    *testValue
		expOk   bool
		expKeys []int
	}{
		"Existing key":    {2, &testValue{2}, true, []int{0, 1, 2, 3}},
		"Nonexisting key": {-vNum, nil, false, []int{0, 1, 2, 3}},
	}

	for name, tCase := range tData {
		t.Run(name, func(t *testing.T) {
			cache, _ := cacheRangeOfInts(t, vNum, nil)
			got, ok := cache.Peek(tCase.key)
			assert.Equal(t, tCase.expV, got)
			assert.Equal(t, tCase.expOk, ok)
			assert.Equal(t, tCase.expKeys, cache.Keys())
		})
	}
}

func Test_LRUCache_Touch(t *testing.T) {
	vNum := 4
	tData := map[string]struct {
		key     int
		expKeys []int
	}{
		"Existing key":    {2, []int{0, 1, 3, 2}},
		"Nonexisting key": {-vNum, []int{0, 1, 2, 3}},
	}

	for name, tCase := range tData {
		t.Run(name, func(t *testing.T) {
			cache, _ := cacheRangeOfInts(t, vNum, nil)
			cache.Touch(tCase.key)
			assert.Equal(t, tCase.expKeys, cache.Keys())
		})
	}
}
