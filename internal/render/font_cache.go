package render

var fontCache = newFontLRU(32)

// SetFontCacheCapacity changes the max number of cached font faces.
func SetFontCacheCapacity(capacity int) {
	fontCache = newFontLRU(capacity)
}

// ClearFontCache releases all cached font.Face objects.
func ClearFontCache() {
	fontCache.clear()
}
