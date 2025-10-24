package render

import (
	"container/list"
	"sync"

	"golang.org/x/image/font"
)

// lruEntry represents a single cache entry containing a font face and its associated key.
type lruEntry struct {
	key  string
	face font.Face
}

// fontLRU implements a thread-safe Least Recently Used (LRU) cache for font.Face objects.
//
// The cache maintains insertion order using a doubly linked list from the `container/list` package.
// When the capacity is exceeded, the least recently used item is evicted.
// Each font.Face that implements `Close()` is properly closed when evicted or cleared.
type fontLRU struct {
	mu       sync.Mutex               // Protects access to the cache
	capacity int                      // Maximum number of cached items
	items    map[string]*list.Element // Map for O(1) access to list elements
	order    *list.List               // Tracks item usage order (oldest → newest)
}

// newFontLRU creates and returns a new LRU cache for font.Face objects with the specified capacity.
// A minimum capacity of 1 is enforced.
func newFontLRU(capacity int) *fontLRU {
	if capacity < 1 {
		capacity = 1
	}
	return &fontLRU{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		order:    list.New(),
	}
}

// get retrieves a font.Face from the cache by key.
// If the key exists, the entry is marked as recently used and returned along with a true flag.
// Otherwise, (nil, false) is returned.
func (c *fontLRU) get(key string) (font.Face, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if el, ok := c.items[key]; ok {
		c.order.MoveToBack(el)
		return el.Value.(*lruEntry).face, true
	}
	return nil, false
}

// put inserts or updates a font.Face in the cache under the specified key.
// If the key already exists, the entry is updated and marked as recently used.
// When the cache exceeds capacity, the least recently used entry is evicted.
// If the evicted font.Face implements Close(), it will be closed.
func (c *fontLRU) put(key string, face font.Face) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Update existing entry
	if el, ok := c.items[key]; ok {
		c.order.MoveToBack(el)
		el.Value.(*lruEntry).face = face
		return
	}

	// Evict the oldest entry if at capacity
	if c.order.Len() >= c.capacity {
		oldest := c.order.Front()
		if oldest != nil {
			ent := oldest.Value.(*lruEntry)
			if closer, ok := ent.face.(interface{ Close() error }); ok {
				_ = closer.Close()
			}
			delete(c.items, ent.key)
			c.order.Remove(oldest)
		}
	}

	// Add new entry
	el := c.order.PushBack(&lruEntry{key: key, face: face})
	c.items[key] = el
}

// clear removes all entries from the cache and closes any font.Face
// that implements the Close() method.
func (c *fontLRU) clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, el := range c.items {
		ent := el.Value.(*lruEntry)
		if closer, ok := ent.face.(interface{ Close() error }); ok {
			_ = closer.Close()
		}
	}

	c.items = make(map[string]*list.Element)
	c.order.Init()
}
