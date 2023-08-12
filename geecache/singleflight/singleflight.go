package singleflight

import "sync"

type call struct {
	wa  sync.WaitGroup
	val interface{}
	err error
}

type Group struct {
	mu sync.Mutex
	m  map[string]*call
}
