package singleflight

import (
	"sync"
)

type call struct { //表示正在进行中或已经结束的请求
	wg  sync.WaitGroup
	val interface{}
	err error
}

type Group struct { //管理不同key的请求
	mu sync.Mutex
	m  map[string]*call
}

func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok { //如果map中有这个call说明有goroutine访问过这个节点了
		g.mu.Unlock()
		c.wg.Wait()         //如果请求正在进行中则等待
		return c.val, c.err //请求结束，返回结果
	}

	c := new(call)
	c.wg.Add(1)  //发起请求前加锁
	g.m[key] = c //添加到g.m，表明key已经有对应的请求在处理
	g.mu.Unlock()

	c.val, c.err = fn() //调用fn发起请求
	c.wg.Done()         //请求结束

	g.mu.Lock()
	delete(g.m, key) //更新g.m
	g.mu.Unlock()

	return c.val, c.err
}
