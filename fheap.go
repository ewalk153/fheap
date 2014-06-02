package fheap

import (
	"errors"
	"math"
)

type Entry struct {
	degree   int    // Number of children
	marked   bool   // Whether this node is marked
	next     *Entry // Next and previous elements in the list
	prev     *Entry
	parent   *Entry      // Parent in the tree, if any.
	child    *Entry      // Child node, if any.
	Element  interface{} // Element being stored here
	Priority float64     // Its priority
}

// NewEntry creates a new Entry element
func newEntry(element interface{}, priority float64) *Entry {
	e := &Entry{
		degree:   0,
		marked:   false,
		Element:  element,
		Priority: priority,
	}
	e.next = e
	e.prev = e
	return e
}

type FibHeap struct {
	min  *Entry // min element in the heap
	size int    // Cached size of the heap, so we don't have to recompute this explicitly
}

func (f *FibHeap) Enqueue(element interface{}, priority float64) *Entry {
	f.checkPriority(priority)
	result := newEntry(element, priority)
	f.min = mergeLists(f.min, result)
	f.size += 1
	return result
}

/**
 * Utility function which, given a user-specified priority, checks whether
 * it's a valid double and throws an IllegalArgumentException otherwise.
 *
 * @param priority The user's specified priority.
 * @throws IllegalArgumentException If it is not valid.
 */
func (f *FibHeap) checkPriority(priority float64) {
	// not sure if is useful.
	// Doublecheck because java code throws a runtime exception here
	// if (Double.isNaN(priority))
	//   throw new IllegalArgumentException(priority + " is invalid.");
}

func mergeLists(one, two *Entry) *Entry {
	if one == nil && two == nil {
		return nil
	} else if one != nil && two == nil {
		return one
	} else if one == nil && two != nil {
		return two
	}
	// Both non-null; actually do the splice.
	/* This is actually not as easy as it seems.  The idea is that we'll
	 * have two lists that look like this:
	 *
	 * +----+     +----+     +----+
	 * |    |--N->|one |--N->|    |
	 * |    |<-P--|    |<-P--|    |
	 * +----+     +----+     +----+
	 *
	 *
	 * +----+     +----+     +----+
	 * |    |--N->|two |--N->|    |
	 * |    |<-P--|    |<-P--|    |
	 * +----+     +----+     +----+
	 *
	 * And we want to relink everything to get
	 *
	 * +----+     +----+     +----+---+
	 * |    |--N->|one |     |    |   |
	 * |    |<-P--|    |     |    |<+ |
	 * +----+     +----+<-\  +----+ | |
	 *                  \  P        | |
	 *                   N  \       N |
	 * +----+     +----+  \->+----+ | |
	 * |    |--N->|two |     |    | | |
	 * |    |<-P--|    |     |    | | P
	 * +----+     +----+     +----+ | |
	 *              ^ |             | |
	 *              | +-------------+ |
	 *              +-----------------+
	 *
	 */

	oneNext := one.next
	one.next = two.next
	one.next.prev = one
	two.next = oneNext
	two.next.prev = two

	/* Return a pointer to whichever's smaller. */
	if one.Priority < two.Priority {
		return one
	}
	return two
}

func (f *FibHeap) Min() *Entry {
	return f.min
}

func (f *FibHeap) Len() int {
	return f.size
}

func MergeFibHeap(one, two *FibHeap) *FibHeap {
	result := &FibHeap{}
	result.min = mergeLists(one.min, two.min)
	result.size = one.size + two.size

	one.min = nil
	two.min = nil
	one.size = 0
	two.size = 0

	return result
}

func (f *FibHeap) DequeueMin() *Entry {
	if f.size == 0 {
		return nil
	}
	f.size -= 1
	minElem := f.min

	/* Now, we need to get rid of this element from the list of roots.  There
	 * are two cases to consider.  First, if this is the only element in the
	 * list of roots, we set the list of roots to be null by clearing mMin.
	 * Otherwise, if it's not null, then we write the elements next to the
	 * min element around the min element to remove it, then arbitrarily
	 * reassign the min.
	 */
	if f.min.next == f.min { // Case one
		f.min = nil
	} else { // Case two
		f.min.prev.next = f.min.next
		f.min.next.prev = f.min.prev
		f.min = f.min.next
	}

	/* Next, clear the parent fields of all of the min element's children,
	 * since they're about to become roots.  Because the elements are
	 * stored in a circular list, the traversal is a bit complex.
	 */
	if minElem.child != nil {
		curr := minElem.child
		for {
			curr.parent = nil
			curr = curr.next
			if curr == minElem.child {
				break
			}
		}
	}

	f.min = mergeLists(f.min, minElem.child)

	if f.min == nil {
		return minElem
	}

	/* Next, we need to coalsce all of the roots so that there is only one
	 * tree of each degree.  To track trees of each size, we allocate an
	 * ArrayList where the entry at position i is either null or the
	 * unique tree of degree i.
	 */
	treeTable := []*Entry{}

	/* We need to traverse the entire list, but since we're going to be
	 * messing around with it we have to be careful not to break our
	 * traversal order mid-stream.  One major challenge is how to detect
	 * whether we're visiting the same node twice.  To do this, we'll
	 * spent a bit of overhead adding all of the nodes to a list, and
	 * then will visit each element of this list in order.
	 */
	toVisit := []*Entry{}

	/* To add everything, we'll iterate across the elements until we
	 * find the first element twice.  We check this by looping while the
	 * list is empty or while the current element isn't the first element
	 * of that list.
	 */
	for curr := f.min; len(toVisit) == 0 || toVisit[0] != curr; curr = curr.next {
		toVisit = append(toVisit, curr)
	}

	/* Traverse this list and perform the appropriate unioning steps. */
	for _, curr := range toVisit {
		/* Keep merging until a match arises. */
		for {
			/* Ensure that the list is long enough to hold an element of this
			 * degree.
			 */

			for curr.degree >= len(treeTable) {
				treeTable = append(treeTable, nil)
			}

			/* If nothing's here, we're can record that this tree has this size
			 * and are done processing.
			 */
			if treeTable[curr.degree] == nil {
				treeTable[curr.degree] = curr
				break
			}

			/* Otherwise, merge with what's there. */
			other := treeTable[curr.degree]
			treeTable[curr.degree] = nil // clear the slot

			/* Determine which of the two trees has the smaller root, storing
			 * the two tree accordingly.
			 */
			var min, max *Entry
			if other.Priority < curr.Priority {
				min = other
				max = curr
			} else {
				min = curr
				max = other
			}

			/* Break max out of the root list, then merge it into min's child
			 * list.
			 */
			max.next.prev = max.prev
			max.prev.next = max.next

			/* Make it a singleton so that we can merge it. */
			max.next = max
			max.prev = max
			min.child = mergeLists(min.child, max)

			/* Reparent max appropriately. */
			max.parent = min

			/* Clear max's mark, since it can now lose another child. */
			max.marked = false

			/* Increase min's degree; it now has another child. */
			min.degree += 1

			/* Continue merging this tree. */
			curr = min
		}

		/* Update the global min based on this node.  Note that we compare
		 * for <= instead of < here.  That's because if we just did a
		 * reparent operation that merged two different trees of equal
		 * priority, we need to make sure that the min pointer points to
		 * the root-level one.
		 */
		if curr.Priority <= f.min.Priority {
			f.min = curr
		}
	}

	return minElem
}

/**
 * Decreases the key of the specified element to the new priority.  If the
 * new priority is greater than the old priority, this function throws an
 * IllegalArgumentException.  The new priority must be a finite double,
 * so you cannot set the priority to be NaN, or +/- infinity.  Doing
 * so also throws an IllegalArgumentException.
 *
 * It is assumed that the entry belongs in this heap.  For efficiency
 * reasons, this is not checked at runtime.
 *
 * @param entry The element whose priority should be decreased.
 * @param newPriority The new priority to associate with this entry.
 * @throws IllegalArgumentException If the new priority exceeds the old
 *         priority, or if the argument is not a finite double.
 */
func (f *FibHeap) DecreaseKey(e *Entry, newPriority float64) error {
	f.checkPriority(newPriority)
	if newPriority > e.Priority {
		return ErrorExceedsPriority
	}
	f.decreaseKeyUnchecked(e, newPriority)
	return nil
}

var ErrorExceedsPriority = errors.New("Error new priority exceeds old")

func (f *FibHeap) Delete(e *Entry) {
	/* Use decreaseKey to drop the entry's key to -infinity.  This will
	 * guarantee that the node is cut and set to the global minimum.
	 */
	f.decreaseKeyUnchecked(e, math.Inf(-1))

	/* Call dequeueMin to remove it. */
	f.DequeueMin()
}

/**
 * Decreases the key of a node in the tree without doing any checking to ensure
 * that the new priority is valid.
 *
 * @param entry The node whose key should be decreased.
 * @param priority The node's new priority.
 */
func (f *FibHeap) decreaseKeyUnchecked(e *Entry, priority float64) {
	/* First, change the node's priority. */
	e.Priority = priority

	/* If the node no longer has a higher priority than its parent, cut it.
	 * Note that this also means that if we try to run a delete operation
	 * that decreases the key to -infinity, it's guaranteed to cut the node
	 * from its parent.
	 */
	if e.parent != nil && e.Priority <= e.parent.Priority {
		f.cutNode(e)
	}

	/* If our new value is the new min, mark it as such.  Note that if we
	 * ended up decreasing the key in a way that ties the current minimum
	 * priority, this will change the min accordingly.
	 */
	if e.Priority <= f.min.Priority {
		f.min = e
	}
}

/**
 * Cuts a node from its parent.  If the parent was already marked, recursively
 * cuts that node from its parent as well.
 *
 * @param entry The node to cut from its parent.
 */
func (f *FibHeap) cutNode(e *Entry) {
	/* Begin by clearing the node's mark, since we just cut it. */
	e.marked = false

	/* Base case: If the node has no parent, we're done. */
	if e.parent == nil {
		return
	}

	/* Rewire the node's siblings around it, if it has any siblings. */
	if e.next != e { // Has siblings
		e.next.prev = e.prev
		e.prev.next = e.next
	}

	/* If the node is the one identified by its parent as its child,
	 * we need to rewrite that pointer to point to some arbitrary other
	 * child.
	 */
	if e.parent.child == e {
		/* If there are any other children, pick one of them arbitrarily. */
		if e.next != e {
			e.parent.child = e.next
		} else {
			/* Otherwise, there aren't any children left and we should clear the
			 * pointer and drop the node's degree.
			 */
			e.parent.child = nil
		}
	}

	/* Decrease the degree of the parent, since it just lost a child. */
	e.parent.degree -= 1

	/* Splice this tree into the root list by converting it to a singleton
	 * and invoking the merge subroutine.
	 */
	e.prev = e
	e.next = e
	f.min = mergeLists(f.min, e)

	/* Mark the parent and recursively cut it if it's already been
	 * marked.
	 */
	if e.parent.marked {
		f.cutNode(e.parent)
	} else {
		e.parent.marked = true
	}

	/* Clear the relocated node's parent; it's now a root. */
	e.parent = nil
}
