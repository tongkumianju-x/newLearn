package main

type LRUcache struct {
	size    int
	maxSize int
	cache   map[int]*Dlinkednode
	head    *Dlinkednode
	laster  *Dlinkednode
}
type Dlinkednode struct {
	key   int
	value int
	perv  *Dlinkednode
	next  *Dlinkednode
}

//func initA(k, v int) *Dlinkednode {
//    return &Dlinkednode{
//        key:   k,
//        value: v,
//    }
//}

func Construct(maxsize int) LRUcache {
	c := LRUcache{
		cache:   make(map[int]*Dlinkednode),
		head:    &Dlinkednode{key: 0, value: 0},
		laster:  &Dlinkednode{key: 0, value: 0},
		maxSize: maxsize,
	}
	c.head.next = c.laster
	c.laster.perv = c.head
	return c
}

func (c *LRUcache) Get(key int) int {
	if _, ok := c.cache[key]; !ok {
		return -1
	}
	node := c.cache[key]
	c.moveTohead(node)
	return node.value
}

func (c *LRUcache) moveTohead()
