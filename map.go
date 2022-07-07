package main

import (
	"sync"
	"time"
)

type item struct {
	value string
	hold  int64
}

type TTLMap struct {
	m map[string]*item
	l sync.Mutex
}

func NewTTLMap(ln int, maxTTL int) (m *TTLMap) {
	m = &TTLMap{m: make(map[string]*item, ln)}
	go func() {
		for now := range time.Tick(time.Second) {
			m.l.Lock()
			for k, v := range m.m {
				if now.Unix()-v.hold > int64(maxTTL) {
					delete(m.m, k)
				}
			}
			m.l.Unlock()
		}
	}()
	return
}

func (m *TTLMap) Len() int {
	return len(m.m)
}

func (m *TTLMap) Put(k, v string) {
	m.l.Lock()
	it, ok := m.m[k]
	if !ok {
		it = &item{value: v}
		m.m[k] = it
	}
	it.hold = time.Now().Unix()
	m.l.Unlock()
}

func (m *TTLMap) Get(k string) (v string, find bool) {
	if it, ok := m.m[k]; ok {
		v = it.value
		return v, true
	}
	return "", false
}

func (m *TTLMap) Have(k string) (find bool) {
	if _, ok := m.m[k]; ok {
		return true
	}
	return false
}
