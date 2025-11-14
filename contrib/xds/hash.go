// Copyright 2022 The imkuqin-zw Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package xds

import (
	"fmt"
	"hash/fnv"
	"sort"
	"sync"

	"github.com/imkuqin-zw/yggdrasil/pkg/resolver"
)

// MaglevHash Maglev一致性哈希实现
type MaglevHash struct {
	mu          sync.RWMutex
	lookup      []uint64
	size        uint64
	backends    int
	permutation [][]uint64
}

// NewMaglevHash 创建Maglev哈希
func NewMaglevHash(backends int) *MaglevHash {
	m := &MaglevHash{
		backends: backends,
		size:     65537, // 默认大素数
	}

	// 生成排列
	m.generatePermutation()

	// 构建查找表
	m.buildLookup()

	return m
}

// generatePermutation 生成Maglev排列
func (m *MaglevHash) generatePermutation() {
	m.permutation = make([][]uint64, m.backends)
	for i := 0; i < m.backends; i++ {
		m.permutation[i] = make([]uint64, m.size)

		// 生成两个偏移量
		offset1 := hash2uint64([]byte{byte(i)}) % m.size
		offset2 := hash2uint64([]byte{byte(i + 1)}) % m.size
		skip := hash2uint64([]byte{byte(i + 2)})%(m.size-1) + 1

		// 生成排列
		var (
			j uint64
			n uint64
		)
		for ; n < m.size; n++ {
			m.permutation[i][n] = (offset1 + offset2*j + n) % m.size
			j += skip
			if j >= m.size {
				j %= m.size
			}
		}
	}
}

// buildLookup 构建查找表
func (m *MaglevHash) buildLookup() {
	m.lookup = make([]uint64, m.size)

	entry := make([]bool, m.size)
	next := make([]uint64, m.backends)

	// 填充查找表
	for j := uint64(0); j < m.size; j++ {
		for i := 0; i < m.backends; i++ {
			candidate := m.permutation[i][next[i]]
			next[i]++
			for entry[candidate] {
				candidate = m.permutation[i][next[i]]
				next[i]++
			}
			entry[candidate] = true
			m.lookup[candidate] = uint64(i)
		}
	}
}

// Lookup 查找后端
func (m *MaglevHash) Lookup(key string) uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.lookup) == 0 {
		return 0
	}

	hash := hash2uint64([]byte(key))
	return m.lookup[hash%uint64(len(m.lookup))]
}

// RingHash 环形哈希实现
type RingHash struct {
	mu       sync.RWMutex
	keys     []uint64
	hash     map[uint64]resolver.Endpoint
	replicas int
}

// NewRingHash 创建环形哈希
func NewRingHash(endpoints []resolver.Endpoint) *RingHash {
	r := &RingHash{
		hash:     make(map[uint64]resolver.Endpoint),
		replicas: 150, // 每个节点150个虚拟节点
	}

	// 添加节点
	for _, endpoint := range endpoints {
		r.addNode(endpoint)
	}

	// 排序键
	r.sortKeys()

	return r
}

// addNode 添加节点
func (r *RingHash) addNode(endpoint resolver.Endpoint) {
	key := endpoint.GetAddress()

	// 创建虚拟节点
	for i := 0; i < r.replicas; i++ {
		hash := hash2uint64([]byte(fmt.Sprintf("%s:%d", key, i)))
		r.keys = append(r.keys, hash)
		r.hash[hash] = endpoint
	}
}

// sortKeys 排序哈希键
func (r *RingHash) sortKeys() {
	sort.Slice(r.keys, func(i, j int) bool {
		return r.keys[i] < r.keys[j]
	})
}

// Lookup 查找节点
func (r *RingHash) Lookup(key string) resolver.Endpoint {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.keys) == 0 {
		return nil
	}

	hash := hash2uint64([]byte(key))

	// 在环形上查找第一个大于等于hash的节点
	idx := sort.Search(len(r.keys), func(i int) bool {
		return r.keys[i] >= hash
	})

	// 如果没找到，返回第一个节点
	if idx == len(r.keys) {
		idx = 0
	}

	return r.hash[r.keys[idx]]
}

// hash2uint64 字符串转uint64哈希
func hash2uint64(data []byte) uint64 {
	h := fnv.New64a()
	_, _ = h.Write(data)
	return h.Sum64()
}

// JumpConsistentHash Jump一致性哈希
type JumpConsistentHash struct {
	buckets int
}

// NewJumpConsistentHash 创建Jump一致性哈希
func NewJumpConsistentHash(buckets int) *JumpConsistentHash {
	return &JumpConsistentHash{
		buckets: buckets,
	}
}

// Lookup Jump哈希查找
func (j *JumpConsistentHash) Lookup(key string) uint64 {
	hash := hash2uint64([]byte(key))
	var b int64 = -1
	var j2 int64

	for {
		b = int64(j.bucket(hash, j2))
		if b >= int64(j.buckets) {
			break
		}
		j2++
	}

	return uint64(b)
}

// bucket Jump哈希桶计算
func (j *JumpConsistentHash) bucket(key uint64, i int64) int64 {
	b := int64(-1)
	var jVar float64
	for jVar < float64(i) {
		b = i
		key = key*2862933555777941757 + 1
		jVar = float64(float64(uint64((1<<31)-1)) * float64(key&0xffffffff))
	}
	return b + 1
}
