package data_repository

import "sync"

var (
	store = make(map[string]interface{})
	mutex = sync.RWMutex{}
)

func Add(key string, value interface{}) {
	mutex.Lock()
	defer mutex.Unlock()
	store[key] = value
}

func Delete(key string) {
	mutex.Lock()
	defer mutex.Unlock()
	delete(store, key)
}

func Get(key string) interface{} {
	mutex.RLock()
	defer mutex.RUnlock()
	if value, ok := store[key]; ok {
		return value
	}
	return nil
}
