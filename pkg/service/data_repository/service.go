package data_repository

import (
	"encoding/json"
	"fmt"
	"sync"

	templateTY "github.com/jkandasa/autoeasy/pkg/types/template"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"go.uber.org/zap"
)

const (
	actionAdd    = "add"
	actionDelete = "delete"
)

var (
	store = make(map[string]interface{})
	mutex = sync.RWMutex{}
)

func Add(key string, value interface{}) {
	mutex.Lock()
	defer mutex.Unlock()

	update(actionAdd, key, value)
}

func Delete(key string) {
	mutex.Lock()
	defer mutex.Unlock()
	update(actionDelete, key, nil)
}

func Get(key string) interface{} {
	mutex.RLock()
	defer mutex.RUnlock()

	jsonBytes, err := json.Marshal(store)
	if err != nil {
		zap.L().Error("error on marshalling store value", zap.Error(err))
	}
	result := gjson.GetBytes(jsonBytes, key)
	return result.Value()
}

func update(action, key string, value interface{}) {
	jsonBytes, err := json.Marshal(store)
	if err != nil {
		zap.L().Error("error on marshalling store value", zap.Error(err))
	}

	switch action {
	case actionAdd:
		jsonBytesUpdated, err := sjson.SetBytes(jsonBytes, key, value)
		if err != nil {
			zap.L().Error("error on updating value", zap.String("key", key), zap.Error(err))
		}
		jsonBytes = jsonBytesUpdated

	case actionDelete:
		jsonBytesUpdated, err := sjson.DeleteBytes(jsonBytes, key)
		if err != nil {
			zap.L().Error("error on deleting value", zap.String("key", key), zap.Error(err))
		}
		jsonBytes = jsonBytesUpdated

	}

	err = json.Unmarshal(jsonBytes, &store)
	if err != nil {
		zap.L().Error("error on unmarshalling store value", zap.Error(err))
	}
}

func GetValue(key string, input interface{}) interface{} {
	jsonBytes, err := json.Marshal(input)
	if err != nil {
		zap.L().Error("error on marshalling input value", zap.String("key", key), zap.Error(err))
	}
	result := gjson.GetBytes(jsonBytes, key)
	return result.Value()
}

func AddWithStore(store templateTY.Store, value interface{}) {
	if store.Key == "" {
		return
	}

	newValue := value
	if store.Query != "" {
		newValue = GetValue(store.Query, value)
	}
	if store.Format != "" {
		newValue = fmt.Sprintf(store.Format, newValue)
	}
	Add(store.Key, newValue)
}
