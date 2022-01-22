package template

import (
	"fmt"
	"sync"

	templateTY "github.com/jkandasa/autoeasy/pkg/types/template"
)

type store struct {
	templates map[string]*templateTY.RawTemplate
	mutex     sync.Mutex
}

var templateStore *store

func init() {
	templateStore = &store{templates: make(map[string]*templateTY.RawTemplate), mutex: sync.Mutex{}}
}

func add(tmpl *templateTY.RawTemplate) error {
	templateStore.mutex.Lock()
	defer templateStore.mutex.Unlock()

	_, ok := templateStore.templates[tmpl.Name]
	if ok {
		return fmt.Errorf("duplicate template entry, name:%s, filename:%s", tmpl.Name, tmpl.FileName)
	}

	templateStore.templates[tmpl.Name] = tmpl
	return nil
}

func Get(name string) (*templateTY.RawTemplate, error) {
	templateStore.mutex.Lock()
	defer templateStore.mutex.Unlock()
	t, ok := templateStore.templates[name]
	if !ok {
		return nil, fmt.Errorf("template not found, name:%s", name)
	}
	return t, nil
}
