package heap

import (
	stdheap "container/heap"
	"math/rand"
	"sort"
	"testing"
)

type testNode struct {
	val int
	idx int32
}

func (n *testNode) Compare(o *testNode) int32 {
	return int32(n.val - o.val)
}

func (n *testNode) GetHeapIndex() int32  { return n.idx }
func (n *testNode) SetHeapIndex(i int32) { n.idx = i }

func TestPushPopSorted(t *testing.T) {
	var h Heap[*testNode]
	const N = 256
	vals := make([]int, 0, N)
	for i := 0; i < N; i++ {
		v := rand.Intn(10000)
		vals = append(vals, v)
		h.Push(&testNode{val: v, idx: -1})
	}
	sort.Ints(vals)

	for i := 0; i < N; i++ {
		n := h.Pop()
		if n == nil {
			t.Fatalf("pop returned nil at %d", i)
		}
		if int32(-1) != n.GetHeapIndex() {
			t.Fatalf("popped node index not -1: %d", n.GetHeapIndex())
		}
		if n.val != vals[i] {
			t.Fatalf("expected %d, got %d at %d", vals[i], n.val, i)
		}
	}
	if !h.Empty() {
		t.Fatalf("heap should be empty after popping all elements")
	}
}

func TestTop(t *testing.T) {
	var h Heap[*testNode]
	h.Push(&testNode{val: 5, idx: -1})
	h.Push(&testNode{val: 3, idx: -1})
	h.Push(&testNode{val: 8, idx: -1})

	top := h.Top()
	if top == nil || top.val != 3 {
		t.Fatalf("expected top=3, got %+v", top)
	}
	if h.Len() != 3 {
		t.Fatalf("len should remain 3, got %d", h.Len())
	}
	p := h.Pop()
	if p.val != 3 {
		t.Fatalf("expected pop=3, got %d", p.val)
	}
}

func TestFixDecreaseKey(t *testing.T) {
	var h Heap[*testNode]
	a := &testNode{val: 5, idx: -1}
	b := &testNode{val: 3, idx: -1}
	c := &testNode{val: 8, idx: -1}
	h.Push(a)
	h.Push(b)
	h.Push(c)

	if h.Top().val != 3 {
		t.Fatalf("expected top=3 before fix")
	}

	c.val = 1
	h.Fix(c)

	if h.Top().val != 1 {
		t.Fatalf("expected top=1 after decrease-key fix, got %d", h.Top().val)
	}
}

func TestFixIncreaseKey(t *testing.T) {
	var h Heap[*testNode]
	a := &testNode{val: 1, idx: -1}
	b := &testNode{val: 2, idx: -1}
	c := &testNode{val: 3, idx: -1}
	h.Push(a)
	h.Push(b)
	h.Push(c)

	if h.Top().val != 1 {
		t.Fatalf("expected top=1 before fix")
	}

	a.val = 9
	h.Fix(a)

	if h.Top().val != 2 {
		t.Fatalf("expected top=2 after increase-key fix, got %d", h.Top().val)
	}
}

func TestClear(t *testing.T) {
	var h Heap[*testNode]
	for i := 0; i < 10; i++ {
		h.Push(&testNode{val: i, idx: -1})
	}
	h.Clear()
	if !h.Empty() || h.Len() != 0 {
		t.Fatalf("heap should be empty after Clear, len=%d", h.Len())
	}
	if top := h.Top(); top != nil {
		t.Fatalf("top should be nil after Clear, got %+v", top)
	}
}

func TestIndexesConsistent(t *testing.T) {
	var h Heap[*testNode]
	ns := make([]*testNode, 0, 32)
	for i := 0; i < 32; i++ {
		n := &testNode{val: rand.Intn(1000), idx: -1}
		ns = append(ns, n)
		h.Push(n)
	}
	// all indexes should be consistent with internal slice positions
	for i := range h.nodes {
		if int32(i) != h.nodes[i].GetHeapIndex() {
			t.Fatalf("node index mismatch at %d: got %d", i, h.nodes[i].GetHeapIndex())
		}
	}
	// mutate some keys and fix
	for i := 0; i < 8; i++ {
		k := rand.Intn(len(ns))
		delta := rand.Intn(200) - 100
		ns[k].val += delta
		h.Fix(ns[k])
	}
	// recheck consistency
	for i := range h.nodes {
		if int32(i) != h.nodes[i].GetHeapIndex() {
			t.Fatalf("node index mismatch after fix at %d: got %d", i, h.nodes[i].GetHeapIndex())
		}
	}
}

// ---------------- Benchmarks: our heap vs container/heap ----------------

type cNode struct {
	val int
	idx int
}

type cNodes []*cNode

func (h cNodes) Len() int           { return len(h) }
func (h cNodes) Less(i, j int) bool { return h[i].val < h[j].val }
func (h cNodes) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].idx = i
	h[j].idx = j
}
func (h *cNodes) Push(x any) {
	n := x.(*cNode)
	n.idx = len(*h)
	*h = append(*h, n)
}
func (h *cNodes) Pop() any {
	old := *h
	n := len(old) - 1
	x := old[n]
	x.idx = -1
	*h = old[:n]
	return x
}

func BenchmarkOurHeap_PushPop_4096(b *testing.B) {
	b.ReportAllocs()
	const N = 4096
	rng := rand.New(rand.NewSource(1))
	vals := make([]int, N)
	for i := 0; i < N; i++ {
		vals[i] = rng.Int()
	}
	b.ResetTimer()
	for it := 0; it < b.N; it++ {
		h := NewHeap[*testNode](N)
		for i := 0; i < N; i++ {
			h.Push(&testNode{val: vals[i], idx: -1})
		}
		for i := 0; i < N; i++ {
			_ = h.Pop()
		}
	}
}

func BenchmarkStdHeap_PushPop_4096(b *testing.B) {
	b.ReportAllocs()
	const N = 4096
	rng := rand.New(rand.NewSource(1))
	vals := make([]int, N)
	for i := 0; i < N; i++ {
		vals[i] = rng.Int()
	}
	b.ResetTimer()
	for it := 0; it < b.N; it++ {
		h := make(cNodes, 0, N)
		for i := 0; i < N; i++ {
			stdheap.Push(&h, &cNode{val: vals[i], idx: -1})
		}
		for i := 0; i < N; i++ {
			_ = stdheap.Pop(&h).(*cNode)
		}
	}
}

func BenchmarkOurHeap_Fix(b *testing.B) {
	b.ReportAllocs()
	const (
		N = 4096
		K = 4096
	)
	rng := rand.New(rand.NewSource(2))
	idxs := make([]int, K)
	deltas := make([]int, K)
	for i := 0; i < K; i++ {
		idxs[i] = rng.Intn(N)
		deltas[i] = rng.Intn(201) - 100
	}
	for it := 0; it < b.N; it++ {
		b.StopTimer()
		h := NewHeap[*testNode](N)
		nodes := make([]*testNode, N)
		for i := 0; i < N; i++ {
			nodes[i] = &testNode{val: rng.Intn(10000), idx: -1}
			h.Push(nodes[i])
		}
		b.StartTimer()
		for i := 0; i < K; i++ {
			n := nodes[idxs[i]]
			n.val += deltas[i]
			h.Fix(n)
		}
	}
}

func BenchmarkStdHeap_Fix(b *testing.B) {
	b.ReportAllocs()
	const (
		N = 4096
		K = 4096
	)
	rng := rand.New(rand.NewSource(2))
	idxs := make([]int, K)
	deltas := make([]int, K)
	for i := 0; i < K; i++ {
		idxs[i] = rng.Intn(N)
		deltas[i] = rng.Intn(201) - 100
	}
	for it := 0; it < b.N; it++ {
		b.StopTimer()
		h := make(cNodes, 0, N)
		nodes := make([]*cNode, N)
		for i := 0; i < N; i++ {
			nodes[i] = &cNode{val: rng.Intn(10000), idx: -1}
			stdheap.Push(&h, nodes[i])
		}
		b.StartTimer()
		for i := 0; i < K; i++ {
			n := nodes[idxs[i]]
			n.val += deltas[i]
			stdheap.Fix(&h, n.idx)
		}
	}
}
