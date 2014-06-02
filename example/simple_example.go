package main

import (
	"fmt"
	"github.com/ewalk153/fheap"
)

func main() {
	h := fheap.FibHeap{}
	h.Enqueue(5, 3)
	h.Enqueue(2, 5)
	h.Enqueue(1, 9)
	for h.Len() > 0 {
		x := h.DequeueMin()
		fmt.Println("Poppping", x)
	}
}
