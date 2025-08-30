package heap

type Node[T any] interface {
	Compare(T) int32
	GetHeapIndex() int32
	SetHeapIndex(int32)
}

type Heap[T Node[T]] struct {
	nodes []T
}

func NewHeap[T Node[T]](size int) *Heap[T] {
	return &Heap[T]{
		nodes: make([]T, 0, size+1),
	}
}

func (q *Heap[T]) Len() int {
	return len(q.nodes)
}

func (q *Heap[T]) Push(x T) {
	q.nodes = append(q.nodes, x)
	i := len(q.nodes) - 1
	for i > 0 {
		p := (i - 1) / 2
		y := q.nodes[p]
		if x.Compare(y) >= 0 {
			break
		}
		q.nodes[i] = y
		y.SetHeapIndex(int32(i))
		i = p
	}
	q.nodes[i] = x
	x.SetHeapIndex(int32(i))
}

func (q *Heap[T]) Pop() T {
	var zero T
	n := len(q.nodes)
	ret := q.nodes[0]
	x := q.nodes[n-1]
	q.nodes[n-1] = zero
	q.nodes = q.nodes[:n-1]
	if n-1 > 0 {
		i := 0
		for {
			l := 2*i + 1
			if l >= len(q.nodes) {
				break
			}
			r := l + 1
			m := l
			y := q.nodes[l]
			if r < len(q.nodes) {
				z := q.nodes[r]
				if z.Compare(y) < 0 {
					m = r
					y = z
				}
			}
			if x.Compare(y) <= 0 {
				break
			}
			q.nodes[i] = y
			y.SetHeapIndex(int32(i))
			i = m
		}
		q.nodes[i] = x
		x.SetHeapIndex(int32(i))
	}
	ret.SetHeapIndex(-1)
	return ret
}

func (q *Heap[T]) Top() T {
	var zero T
	if len(q.nodes) == 0 {
		return zero
	}
	return q.nodes[0]
}

func (q *Heap[T]) Fix(x T) {
	var (
		i = int(x.GetHeapIndex())
		n = len(q.nodes)
		p int
	)
	if i > 0 && x.Compare(q.nodes[(i-1)/2]) < 0 {
		for i > 0 {
			p = (i - 1) / 2
			y := q.nodes[p]
			if x.Compare(y) >= 0 {
				break
			}
			q.nodes[i] = y
			y.SetHeapIndex(int32(i))
			i = p
		}
		q.nodes[i] = x
		x.SetHeapIndex(int32(i))
		return
	}
	for {
		l := 2*i + 1
		if l >= n {
			break
		}
		r := l + 1
		m := l
		y := q.nodes[l]
		if r < n {
			z := q.nodes[r]
			if z.Compare(y) < 0 {
				m = r
				y = z
			}
		}
		if x.Compare(y) <= 0 {
			break
		}
		q.nodes[i] = y
		y.SetHeapIndex(int32(i))
		i = m
	}
	q.nodes[i] = x
	x.SetHeapIndex(int32(i))
}

func (q *Heap[T]) Empty() bool {
	return len(q.nodes) == 0
}

func (q *Heap[T]) Clear() {
	var zero T
	for i := range q.nodes {
		q.nodes[i] = zero
	}
	q.nodes = q.nodes[:0]
}
