package main

import (
	"fmt"
	"sync"
	"time"
)

// 1.主Cache需要记录头节点，尾节点，和最长度以及 存储结构是map+双向链表 同时需要配置读写锁和stop管道
// 2.新建主Cache时需要新建stop管道+缓存；而传入参数是最大值和过期时间，最大值赋值即可，过期时间则需要ticker
// 3.三个工具函数 1）node移动至头节点：（场景：set元素/get已有元素）(需要注意第一次插入时，没有设置头节点和尾节点的情况)
//                 1。node就是头节点，就不用移动
//                 2。如果是已有元素，node需要先移除节点，再放为头节点/且需要更新数据
//                 3。如果是set未有节点，node插入后，需要删除尾节点
//              2）删除尾节点 （先存再移动再删，需要注意如果只有一个节点的情况）
//              3）移除节点
// 4.过期清除+删除节点+写入节点都需要加写锁，获取节点时加读锁

// LRU节点结构
type lruNode[K comparable, V any] struct {
	key   K
	value Itemi[V]
	prev  *lruNode[K, V]
	next  *lruNode[K, V]
}

type MyCache[K comparable, V any] struct {
	Items    map[K]*lruNode[K, V] // 改为指向LRU节点的指针
	mu       sync.RWMutex
	stop     chan struct{}
	capacity int            // 容量限制
	head     *lruNode[K, V] // LRU链表头
	tail     *lruNode[K, V] // LRU链表尾
}

type Itemi[V any] struct {
	value   V
	OutTime int64
}

func NewCaches[K comparable, V any](clearupIneterVal time.Duration, capacity int) *MyCache[K, V] {
	c := &MyCache[K, V]{
		Items:    make(map[K]*lruNode[K, V]),
		stop:     make(chan struct{}),
		capacity: capacity,
	}
	if clearupIneterVal > 0 {
		go c.startClearnup(clearupIneterVal)
	}
	return c
}

// 将节点移动到链表头部（表示最近使用）
func (c *MyCache[K, V]) moveToHead(node *lruNode[K, V]) {
	if node == c.head {
		return
	}
	// 从原位置移除
	if node.prev != nil {
		node.prev.next = node.next
	}
	if node.next != nil {
		node.next.prev = node.prev
	}
	if node == c.tail {
		c.tail = node.prev
	}

	// 插入到头部
	if c.head != nil {
		c.head.prev = node
	}
	node.next = c.head
	node.prev = nil
	c.head = node

	if c.tail == nil {
		c.tail = node
	}
}

// 移除尾部节点（最久未使用）
func (c *MyCache[K, V]) removeTail() {
	if c.tail == nil {
		return
	}

	delete(c.Items, c.tail.key)

	if c.tail.prev != nil {
		c.tail.prev.next = nil
	} else {
		c.head = nil
	}
	c.tail = c.tail.prev
}

func (c *MyCache[K, V]) startClearnup(clearupIneterVal time.Duration) {
	ticker := time.NewTicker(clearupIneterVal)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.clearup()
		case <-c.stop:
			return
		}
	}
}

func (c *MyCache[K, V]) clearup() {
	now := time.Now().UnixNano()
	c.mu.Lock()
	defer c.mu.Unlock()

	for k, node := range c.Items {
		if node.value.OutTime > 0 && now > node.value.OutTime {
			// 从链表中移除过期节点
			if node.prev != nil {
				node.prev.next = node.next
			}
			if node.next != nil {
				node.next.prev = node.prev
			}
			if node == c.head {
				c.head = node.next
			}
			if node == c.tail {
				c.tail = node.prev
			}
			delete(c.Items, k)
		}
	}
}

func (c *MyCache[K, V]) Set(key K, value V, ttl time.Duration) {
	var outTime int64
	if ttl > 0 {
		outTime = time.Now().Add(ttl).UnixNano()
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// 如果key已存在，更新值并移动到头部
	if node, exists := c.Items[key]; exists {
		node.value = Itemi[V]{
			value:   value,
			OutTime: outTime,
		}
		c.moveToHead(node)
		return
	}

	// 如果达到容量限制，移除最久未使用的
	if len(c.Items) >= c.capacity && c.capacity > 0 {
		c.removeTail()
	}

	// 创建新节点
	newNode := &lruNode[K, V]{
		key: key,
		value: Itemi[V]{
			value:   value,
			OutTime: outTime,
		},
	}

	// 添加到链表头部
	if c.head != nil {
		c.head.prev = newNode
	}
	newNode.next = c.head
	c.head = newNode
	if c.tail == nil {
		c.tail = newNode
	}

	c.Items[key] = newNode
}

func (c *MyCache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	node, ok := c.Items[key]
	if !ok {
		var zero V
		return zero, false
	}

	// 检查是否过期
	if node.value.OutTime > 0 && node.value.OutTime < time.Now().UnixNano() {
		// 从链表中移除过期节点
		if node.prev != nil {
			node.prev.next = node.next
		}
		if node.next != nil {
			node.next.prev = node.prev
		}
		if node == c.head {
			c.head = node.next
		}
		if node == c.tail {
			c.tail = node.prev
		}
		delete(c.Items, key)
		var zero V
		return zero, false
	}

	// 移动到头部表示最近使用
	c.moveToHead(node)

	return node.value.value, true
}

func (c *MyCache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if node, exists := c.Items[key]; exists {
		// 从链表中移除
		if node.prev != nil {
			node.prev.next = node.next
		}
		if node.next != nil {
			node.next.prev = node.prev
		}
		if node == c.head {
			c.head = node.next
		}
		if node == c.tail {
			c.tail = node.prev
		}
		delete(c.Items, key)
	}
}

func (c *MyCache[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.Items)
}

func (c *MyCache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Items = make(map[K]*lruNode[K, V])
	c.head = nil
	c.tail = nil
}

func (c *MyCache[K, V]) close() {
	close(c.stop)
}

func main() {
	// 创建一个缓存，容量为3，每隔5秒清理一次过期键
	cache := NewCaches[string, string](5*time.Second, 3)

	// 设置键值，有效期2秒
	cache.Set("name", "Alice", 2*time.Second)
	cache.Set("age", "25", 0) // 永不过期
	cache.Set("city", "Beijing", 0)

	// 测试LRU：访问name使其成为最近使用的
	if val, ok := cache.Get("name"); ok {
		fmt.Println("Get:", val) // 输出: Get: Alice
	}

	// 添加第四个元素，应该淘汰最久未使用的(age)
	cache.Set("country", "China", 0)

	// 检查age是否被淘汰
	if _, ok := cache.Get("age"); !ok {
		fmt.Println("age was evicted by LRU")
	}

	// 检查其他键是否存在
	fmt.Println("Current items:")
	if val, ok := cache.Get("name"); ok {
		fmt.Println("  name:", val)
	}
	if val, ok := cache.Get("city"); ok {
		fmt.Println("  city:", val)
	}
	if val, ok := cache.Get("country"); ok {
		fmt.Println("  country:", val)
	}

	// 关闭缓存
	defer cache.close()
}
