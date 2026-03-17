package render

import "sync"

var (
	fontCacheMu sync.RWMutex
	fontCache   = newFontLRU(32)
)

// getFontCache returns the current font cache instance safely.
func getFontCache() *fontLRU {
	fontCacheMu.RLock()
	c := fontCache
	fontCacheMu.RUnlock()
	return c
}

// SetFontCacheCapacity replaces the global font-face cache with a new one of
// the given capacity. Existing cached faces are discarded.
func SetFontCacheCapacity(capacity int) {
	c := newFontLRU(capacity)
	fontCacheMu.Lock()
	fontCache = c
	fontCacheMu.Unlock()
}

// ClearFontCache releases all cached font.Face objects.
func ClearFontCache() {
	getFontCache().clear()
}
