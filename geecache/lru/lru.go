package lru

import "container/list"

// Cache lruCache
type Cache struct {
	maxBytes int64 //允许使用的最大内存
	nbytes   int64 //记录当前已经使用的内存
	ll       *list.List
	cache    map[string]*list.Element //键是字符串，值是双向链表中对应节点的指针
	// optional and executed when an entry is purged.
	OnEvicted func(key string, value Value) //某条记录
}

// entry 是双向链表节点的数据类型，在链表中保存每个值对应的key的好处在于，淘汰队首节点时，需要用key从从字典中删除对应的映射
type entry struct {
	key   string
	value Value
}

// Value 允许值是实现了Value接口的任意类型，只包含一个方法，用于返回值所占用的内存大小
type Value interface {
	Len() int
}

// New 方便实例化Cache，实现New函数
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Add 向缓存中增加一个entry
func (c *Cache) Add(key string, value Value) {
	//如果键存在，则更新对应节点的值，并将节点移到队尾
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		//不存在，此时是新增场景，首先队尾添加新节点entry，并向字典中添加key和结点的映射关系
		ele := c.ll.PushFront(&entry{key, value})
		//更新nbytes
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	//如果超过了设定了的最大值，则移除最少访问的节点
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

// Get 查找功能，第一步是从字典中找到对应的双向链表的节点，第二步，将该节点移动到队尾
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok { //如果存在的话
		c.ll.MoveToFront(ele) //将链表中的节点ele移动到队尾（双向链表作为队列，队首队尾是相对的，这里约定front为队尾）
		kv := ele.Value.(*entry)
		return kv.value, true //返回value字段
	}
	return
}

// RemoveOldest 缓存淘汰，移除最近最少访问的节点（队首）
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// Len 用来获取添加了多少数据
func (c *Cache) Len() int {
	return c.ll.Len()
}
