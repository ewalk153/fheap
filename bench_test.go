package fheap

import (
	"container/heap"
	"math/rand"
	"testing"
)

const (
	HEAP_ADD_SIZE int = 200000
	HEAP_LOOP     int = 200000
)

// Reference implementation of PriorityQueue, with float64 priority numbers
type Item struct {
	value    string  // The value of the item; arbitrary.
	priority float64 // The priority of the item in the queue.
	// The index is needed by update and is maintained by the heap.Interface methods.
	index int // The index of the item in the heap.
}

// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue []*Item

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq[i].priority > pq[j].priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// update modifies the priority and value of an Item in the queue.
func (pq *PriorityQueue) update(item *Item, value string, priority float64) {
	heap.Remove(pq, item.index)
	item.value = value
	item.priority = priority
	heap.Push(pq, item)
}

func randSetup() *rand.Rand {
	return rand.New(rand.NewSource(99))
}

func BenchmarkFibPush(b *testing.B) {
	r := randSetup()
	h := FibHeap{}

	for j := 0; j < b.N; j++ {
		for i := 0; i < HEAP_ADD_SIZE; i++ {
			s := Item{priority: r.Float64()}
			h.Enqueue(s, s.priority)
		}
	}
}

func BenchmarkPriorityPush(b *testing.B) {
	r := randSetup()
	pq := &PriorityQueue{}
	heap.Init(pq)
	for j := 0; j < b.N; j++ {
		for i := 0; i < HEAP_ADD_SIZE; i++ {
			s := Item{priority: r.Float64()}
			heap.Push(pq, &s)
		}
	}
}

// func BenchmarkFibPop(b *testing.B) {
// 	r := randSetup()
// 	h := FibHeap{}

// 	s := Item{}
// 	for i := 0; i < HEAP_SIZE; i++ {
// 		h.Enqueue(s, r.Float64())
// 	}
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		h.DequeueMin()
// 	}
// }

// func BenchmarkPriorityPop(b *testing.B) {
// 	r := randSetup()
// 	pq := &PriorityQueue{}
// 	heap.Init(pq)
// 	for i := 0; i < HEAP_SIZE; i++ {
// 		s := Item{priority: r.Float64()}
// 		heap.Push(pq, &s)
// 	}
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		heap.Pop(pq)
// 	}
// }

func BenchmarkFibPushPop(b *testing.B) {
	r := randSetup()
	h := FibHeap{}

	for j := 0; j < b.N; j++ {
		for i := 0; i < HEAP_LOOP; i++ {
			s := Item{}
			h.Enqueue(s, r.Float64())
		}
		for i := 0; i < HEAP_LOOP; i++ {
			h.DequeueMin()
		}
	}
}

func BenchmarkPriorityPushPop(b *testing.B) {
	r := randSetup()
	pq := &PriorityQueue{}
	heap.Init(pq)
	for j := 0; j < b.N; j++ {
		for i := 0; i < HEAP_LOOP; i++ {
			s := Item{priority: r.Float64()}
			heap.Push(pq, &s)
		}
		for i := 0; i < HEAP_LOOP; i++ {
			heap.Pop(pq)
		}
	}
}
