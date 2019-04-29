/*
Copyright 2012 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// 控制重复的请求只执行1次
package singleflight

import "sync"

// 执行中或者执行完成的结果
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

// Group代表重复请求的一组操作
type Group struct {
	mu sync.Mutex
	m  map[string]*call
}

// 保证对同一个key的请求不会出现并发重复操作
// 如果存在重复请求，等待上一个操作完成返回相同响应
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	// 加锁操作
	g.mu.Lock()

	// 延迟初始化
	if g.m == nil {
		g.m = make(map[string]*call)
	}

	// 如果存在重复请求，阻塞，等待WaitGroup Done，返回响应和错误
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}

	// 如果不存在重复请求，创建Call结构和WaitGroup
	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	// 执行请求操作，完成之后删除对应的哈希表记录
	c.val, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}
