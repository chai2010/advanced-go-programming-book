package gls

import "sync"

var gls struct {
	m map[int64]map[interface{}]interface{}
	sync.Mutex
}

func init() {
	gls.m = make(map[int64]map[interface{}]interface{})
}

func getMap() map[interface{}]interface{} {
	gls.Lock()
	defer gls.Unlock()

	goid := GetGoid()
	if m, _ := gls.m[goid]; m != nil {
		return m
	}

	m := make(map[interface{}]interface{})
	gls.m[goid] = m
	return m
}

func Get(key interface{}) interface{} {
	return getMap()[key]
}
func Put(key interface{}, v interface{}) {
	getMap()[key] = v
}
func Delete(key interface{}) {
	delete(getMap(), key)
}

func Clean() {
	gls.Lock()
	defer gls.Unlock()

	delete(gls.m, GetGoid())
}
