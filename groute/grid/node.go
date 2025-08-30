package grid

import (
	"container/heap"
	"unsafe"
)

type NodeStatus int32

const (
	NodeNew = NodeStatus(iota)
	NodeOpen
	NodeClose
)

type Gpos struct {
	X, Y int32
}

func (p Gpos) Hash() int32 {
	return hash64(*(*int64)(unsafe.Pointer(&p)))
}

type Gnode struct {
	Pos, FPos Gpos
	Dir       int32
	Cost      int32 // g
	Total     int32 // g+h
	Status    NodeStatus
}

type NodePool struct {
	mNode []Gnode // size = maxNodes
	mHash map[Gpos]int32

	maxNodes, nodeCnt int32
}

func NewNodePool(maxNodes int32) *NodePool {
	pool := &NodePool{
		mNode:    make([]Gnode, maxNodes),
		mHash:    make(map[Gpos]int32, maxNodes),
		maxNodes: maxNodes,
	}
	return pool
}

func (p *NodePool) Clear() {
	clear(p.mHash)
	p.nodeCnt = 0
}

func (p *NodePool) GetNode(x, y int32) *Gnode {
	pos := Gpos{X: x, Y: y}
	i, ok := p.mHash[pos]
	if ok {
		return &p.mNode[i]
	}

	if p.nodeCnt >= p.maxNodes {
		return nil
	}
	i = p.nodeCnt
	p.mNode[i] = Gnode{
		Pos: pos,
	}
	p.mHash[pos] = i
	p.nodeCnt++
	return &p.mNode[i]
}

func (p *NodePool) FindNode(x, y int32) *Gnode {
	if i, ok := p.mHash[Gpos{X: x, Y: y}]; ok {
		return &p.mNode[i]
	}
	return nil
}

type nodes []*Gnode

func (n *nodes) Len() int {
	return len(*n)
}

func (n *nodes) Less(i, j int) bool {
	return (*n)[i].Total < (*n)[j].Total
}

func (n *nodes) Swap(i, j int) {
	(*n)[i], (*n)[j] = (*n)[j], (*n)[i]
}

func (n *nodes) Push(x interface{}) {
	*n = append(*n, x.(*Gnode))
}

func (n *nodes) Pop() interface{} {
	var l = len(*n) - 1
	var x = (*n)[l]
	(*n)[l] = nil
	*n = (*n)[:l]
	return x
}

type NodeQueue struct {
	mHeap nodes
}

func NewNodeQueue(size int) *NodeQueue {
	return &NodeQueue{
		mHeap: make([]*Gnode, 0, size+1),
	}
}

func (q *NodeQueue) Push(x *Gnode) {
	heap.Push(&q.mHeap, x)
}

func (q *NodeQueue) Pop() *Gnode {
	return heap.Pop(&q.mHeap).(*Gnode)
}

func (q *NodeQueue) Top() *Gnode {
	return q.mHeap[0]
}

func (q *NodeQueue) Fix(y *Gnode) {
	for i, x := range q.mHeap {
		if x == y {
			heap.Fix(&q.mHeap, i)
			return
		}
	}
}

func (q *NodeQueue) Empty() bool {
	return len(q.mHeap) == 0
}

func (q *NodeQueue) Clear() {
	q.mHeap = q.mHeap[:0]
}
