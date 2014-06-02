package main

import (
	"fmt"
	"github.com/ewalk153/fheap"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(22)
	time.Now()
	// rand.Seed(time.Now().UnixNano())
}

func main() {
	h := fheap.FibHeap{}
	for i := 0; i < 20; i++ {
		h.Enqueue(i, rand.Float64())
	}
	for x := h.DequeueMin(); x != nil; x = h.DequeueMin() {
		fmt.Println("Min", x.Element)
	}
}
