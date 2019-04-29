/*
Copyright 2013 Google Inc.

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

// 一致性哈希算法的实现
package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

// 哈希环数据结构
type Map struct {
	hash     Hash           // 哈希算法
	replicas int            // 为了让服务节点更加分散
	keys     []int          // 哈希值列表
	hashMap  map[int]string // 哈希值对应的服务节点
}

// 创建哈希环数据结构
func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	// 默认使用的哈希算法：crc32.ChecksumIEEE
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// 判断节点个数是否为0
func (m *Map) IsEmpty() bool {
	return len(m.keys) == 0
}

// 增加节点到哈希环
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			// 节点的字符串添加replica，为了哈希值的分散
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	// 将哈希值列表升序便于搜索
	sort.Ints(m.keys)
}

// 获取key哈希值对应的服务节点
func (m *Map) Get(key string) string {
	if m.IsEmpty() {
		return ""
	}

	// 哈希列表中找到比key的哈希值大的第1个值
	hash := int(m.hash([]byte(key)))
	idx := sort.Search(len(m.keys), func(i int) bool { return m.keys[i] >= hash })
	if idx == len(m.keys) {
		idx = 0
	}

	return m.hashMap[m.keys[idx]]
}
