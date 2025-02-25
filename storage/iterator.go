package storage

// Iterator iteriert über eine map[string]string
type MapIterator struct {
	keys   []string
	values map[string]string
	index  int
}

// NewMapIterator erstellt einen neuen Iterator für map[string]string
func newMapIterator(items map[string]string) *MapIterator {
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	return &MapIterator{keys: keys, values: items, index: -1}
}

// Next bewegt den Iterator zum nächsten Element und gibt zurück, ob ein weiteres Element existiert
func (it *MapIterator) Next() bool {
	if it.index+1 < len(it.keys) {
		it.index++
		return true
	}
	return false
}

// Value gibt das aktuelle Schlüssel-Wert-Paar zurück
func (it *MapIterator) Value() (string, string) {
	key := it.keys[it.index]
	return key, it.values[key]
}
