package gls

import (
	"reflect"
)

func getg() interface{}

type eface struct {
	_type, elem uintptr
}

//go:nosplit
func runtime_convT2E_hack(_type, elem uintptr) eface {
	return eface{
		_type: _type,
		elem:  elem,
	}
}

func GetGoid() int64 {
	g := getg()
	goid := reflect.ValueOf(g).FieldByName("goid").Int()
	return goid
}
