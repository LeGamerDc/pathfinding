package grid

import "container/heap"

type nodeStatus uint32

const (
	nodeNew = nodeStatus(iota)
	nodeOpen
	nodeClose
)

type gNode struct {
	pos, fPos int32
	dir       int32
	cost      int32 // g
	total     int32 // g+h
	status    nodeStatus
}

type gNodePool struct {
	mNode                       []gNode // size = maxNodes
	mNext                       []int32 // size = maxNodes
	mFirst                      []int32 // size = hashSize
	maxNodes, hashSize, nodeCnt int32
}

func newNodePool(maxNodes int32) *gNodePool {
	var p = new(gNodePool)
	p.maxNodes = maxNodes
	p.hashSize = int32(nextPow2(uint32(maxNodes / 4)))
	p.mNode = make([]gNode, maxNodes)
	p.mNext = make([]int32, maxNodes)
	p.mFirst = make([]int32, p.hashSize)
	memset(p.mFirst, -1)
	memset(p.mNext, -1)
	return p
}

func (p *gNodePool) clear() {
	memset(p.mFirst, -1)
	p.nodeCnt = 0
}

func (p *gNodePool) getNode(pos int32) *gNode {
	var bucket = hash32(pos) & (p.hashSize - 1)
	var i = p.mFirst[bucket]
	for i != -1 {
		if p.mNode[i].pos == pos {
			return &p.mNode[i]
		}
		i = p.mNext[i]
	}
	if p.nodeCnt >= p.maxNodes {
		return nil
	}
	i = p.nodeCnt
	p.nodeCnt++
	p.mNode[i] = gNode{
		pos: pos,
	}
	p.mNext[i] = p.mFirst[bucket]
	p.mFirst[bucket] = i
	return &p.mNode[i]
}

func (p *gNodePool) findNode(pos int32) *gNode {
	var bucket = hash32(pos) & (p.hashSize - 1)
	var i = p.mFirst[bucket]
	for i != -1 {
		if p.mNode[i].pos == pos {
			return &p.mNode[i]
		}
		i = p.mNext[i]
	}
	return nil
}

////////////////////////// node queue //////////////////////////////

type nodes []*gNode

func (n *nodes) Len() int {
	return len(*n)
}

func (n *nodes) Less(i, j int) bool {
	return (*n)[i].total < (*n)[j].total
}

func (n *nodes) Swap(i, j int) {
	(*n)[i], (*n)[j] = (*n)[j], (*n)[i]
}

func (n *nodes) Push(x interface{}) {
	*n = append(*n, x.(*gNode))
}

func (n *nodes) Pop() interface{} {
	var l = len(*n) - 1
	var x = (*n)[l]
	(*n)[l] = nil
	*n = (*n)[:l]
	return x
}

type gNodeQueue struct {
	mHeap nodes
}

func newNodeQueue(size int32) *gNodeQueue {
	var q = new(gNodeQueue)
	q.mHeap = make([]*gNode, 0, size+1)
	return q
}

func (q *gNodeQueue) push(n *gNode) {
	heap.Push(&q.mHeap, n)
}

func (q *gNodeQueue) top() *gNode {
	return q.mHeap[0]
}

func (q *gNodeQueue) pop() *gNode {
	return heap.Pop(&q.mHeap).(*gNode)
}

func (q *gNodeQueue) empty() bool {
	return len(q.mHeap) == 0
}

func (q *gNodeQueue) fix(n *gNode) {
	for i, x := range q.mHeap {
		if x == n {
			heap.Fix(&q.mHeap, i)
			return
		}
	}
}

func (q *gNodeQueue) clear() {
	q.mHeap = q.mHeap[0:0]
}
