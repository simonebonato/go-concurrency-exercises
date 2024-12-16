// Personal comment:
// this exercise is quite simple, but it shows that it can be really important to use 
// a MUTEX to prevent multiple goroutines to read and/or write the same elements at the 
// same time.
// I was trying to use a more complicated approach with the RWMutex, but that did not work
// simply because it could happen that when a routine was reading the map, another one could 
// change the value at the same time

//////////////////////////////////////////////////////////////////////
//
// Given is some code to cache key-value pairs from a database into
// the main memory (to reduce access time). Note that golang's map are
// not entirely thread safe. Multiple readers are fine, but multiple
// writers are not. Change the code to make this thread safe.
//

package main

import (
	"container/list"
	"sync"
	"testing"
)

// CacheSize determines how big the cache can grow
const CacheSize = 100

// KeyStoreCacheLoader is an interface for the KeyStoreCache
type KeyStoreCacheLoader interface {
	// Load implements a function where the cache should gets it's content from
	Load(string) string
}

type page struct {
	Key   string
	Value string
}

// KeyStoreCache is a LRU cache for string key-value pairs
type KeyStoreCache struct {
	cache map[string]*list.Element
	pages list.List
	load  func(string) string
	mu    sync.Mutex
}

// New creates a new KeyStoreCache
func New(load KeyStoreCacheLoader) *KeyStoreCache {
	return &KeyStoreCache{
		load:  load.Load,
		cache: make(map[string]*list.Element),
	}
}

// Get gets the key from cache, loads it from the source if needed
func (k *KeyStoreCache) Get(key string) string {
	k.mu.Lock()
	defer k.mu.Unlock()

	if e, ok := k.cache[key]; ok { // PROBLEMATIC LINE!!!
		k.pages.MoveToFront(e)
		return e.Value.(page).Value
	}
	// Miss - load from database and save it in cache
	p := page{key, k.load(key)}
	// if cache is full remove the least used item
	if len(k.cache) >= CacheSize {
		end := k.pages.Back()
		// remove from map
		delete(k.cache, end.Value.(page).Key) // PROBLEMATIC LINE!!!
		// remove from list
		k.pages.Remove(end) // PROBLEMATIC LINE!!!
	}
	k.pages.PushFront(p)           // PROBLEMATIC LINE!!!
	k.cache[key] = k.pages.Front() // PROBLEMATIC LINE!!!
	return p.Value
}

// Loader implements KeyStoreLoader
type Loader struct {
	DB *MockDB
}

// Load gets the data from the database
func (l *Loader) Load(key string) string {
	val, err := l.DB.Get(key)
	if err != nil {
		panic(err)
	}

	return val
}

func run(t *testing.T) (*KeyStoreCache, *MockDB) {
	loader := Loader{
		DB: GetMockDB(),
	}
	cache := New(&loader)

	RunMockServer(cache, t)

	return cache, loader.DB
}

func main() {
	run(nil)
}
