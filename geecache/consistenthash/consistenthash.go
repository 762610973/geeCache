package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash 定义函数类型hash，采取依赖注入的方式，允许用于替换成自定义的hash函数，也方便测试时替换
//默认为crc32.ChecksumIEEE算法
type Hash func(data []byte) uint32

// Map 一致性哈希算法的主数据结构，
type Map struct {
	hash     Hash
	replicas int            //虚拟节点倍数
	keys     []int          // Sorted 虚拟节点的哈希值，哈喜欢keys，
	hashMap  map[int]string //虚拟节点和真实节点的映射表hashMap，键是虚拟节点的哈希值，值是真实节点的名称
}

// New creates a Map instance 允许自定义虚拟节点倍数和hash函数
func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add 添加真实节点/机器
func (m *Map) Add(keys ...string) {
	//允许传入0个或者多个真实节点的名称
	for _, key := range keys {
		//对每一个真实节点key，对应创建m.replicas个虚拟节点
		// 虚拟节点的名称是strconv.Itoa(i) + key
		for i := 0; i < m.replicas; i++ {
			//m.Hash计算虚拟节点的哈希值
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			//添加到环上
			m.keys = append(m.keys, hash)
			//增加虚拟节点和真实节点的映射关系
			m.hashMap[hash] = key
		}
	}
	//环上的哈希值排序
	sort.Ints(m.keys)
}

// Get gets the closest item in the hash to the provided key.
func (m *Map) Get(key string) string {
	// 计算key的哈希值，顺时针找到第一个陪陪的虚拟节点的下标idx，从m.keys中获取对应的哈希值，
	if len(m.keys) == 0 {
		return ""
	}
	//如果idx==len(m.keys)，说明应选择m.keys[0]，因为是一个m.keys环状结构，用取余的方式来处理这种情况
	hash := int(m.hash([]byte(key)))
	// Binary search for appropriate replica.
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	return m.hashMap[m.keys[idx%len(m.keys)]]
}
