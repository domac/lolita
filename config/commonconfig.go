package config

import (
	"github.com/codegangsta/inject"
	injectutil "github.com/domac/lolita/util"
	"reflect"
)

type TypeConfig interface {
	SetInjector(inj inject.Injector)
	GetType() string
	Invoke(f interface{}) (refvs []reflect.Value, err error)
}

type CommonConfig struct {
	inject.Injector `json:"-"`
	Type            string `json:"type"`
}

func (t *CommonConfig) SetInjector(inj inject.Injector) {
	t.Injector = inj
}

func (t *CommonConfig) GetType() string {
	return t.Type
}

func (t *CommonConfig) Invoke(f interface{}) (refvs []reflect.Value, err error) {
	return injectutil.Invoke(t.Injector, f)
}

type ConfigRaw map[string]interface{}
