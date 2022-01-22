package variable

import (
	"fmt"
	"sync"

	variableTY "github.com/jkandasa/autoeasy/pkg/types/variable"
)

type store struct {
	variables map[string]*variableTY.VariableConfigPre
	mutex     sync.Mutex
}

var varsStore = &store{variables: make(map[string]*variableTY.VariableConfigPre), mutex: sync.Mutex{}}

func add(varCfg *variableTY.VariableConfigPre) error {
	varsStore.mutex.Lock()
	defer varsStore.mutex.Unlock()

	_, ok := varsStore.variables[varCfg.Name]
	if ok {
		return fmt.Errorf("duplicate variables entry, name:%s, filename:%s", varCfg.Name, varCfg.FileName)
	}

	varsStore.variables[varCfg.Name] = varCfg
	return nil
}

func Get(name string) (*variableTY.VariableConfigPre, error) {
	varsStore.mutex.Lock()
	defer varsStore.mutex.Unlock()
	t, ok := varsStore.variables[name]
	if !ok {
		return nil, fmt.Errorf("variables not found, name:%s", name)
	}
	return t, nil
}
