package fheap

import (
	"fmt"
	"math/rand"
)

// This example illustrates queing and dequeuing from
// a Fibonacci heap.
func Example_fHeap() {
	rand.Seed(22)
	h := FibHeap{}
	for i := 0; i < 10; i++ {
		h.Enqueue(i, rand.Float64())
	}
	for x := h.DequeueMin(); x != nil; x = h.DequeueMin() {
		fmt.Println("Min", x.Element)
	}
	// Output:
	// Min 4
	// Min 9
	// Min 1
	// Min 6
	// Min 0
	// Min 2
	// Min 3
	// Min 8
	// Min 5
	// Min 7
}
