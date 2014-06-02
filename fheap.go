/***********************************************************************
 * File: fheap.go (Original was FibonacciHeap.java)
 * Author: Keith Schwarz (htiek@cs.stanford.edu)
 * Adapted for Go by Eric Walker
 *
 * An implementation of a priority queue backed by a Fibonacci heap,
 * as described by Fredman and Tarjan.  Fibonacci heaps are interesting
 * theoretically because they have asymptotically good runtime guarantees
 * for many operations.  In particular, insert, peek, and decrease-key all
 * run in amortized O(1) time.  dequeueMin and delete each run in amortized
 * O(lg n) time.  This allows algorithms that rely heavily on decrease-key
 * to gain significant performance boosts.  For example, Dijkstra's algorithm
 * for single-source shortest paths can be shown to run in O(m + n lg n) using
 * a Fibonacci heap, compared to O(m lg n) using a standard binary or binomial
 * heap.
 *
 * Internally, a Fibonacci heap is represented as a circular, doubly-linked
 * list of trees obeying the min-heap property.  Each node stores pointers
 * to its parent (if any) and some arbitrary child.  Additionally, every
 * node stores its degree (the number of children it has) and whether it
 * is a "marked" node.  Finally, each Fibonacci heap stores a pointer to
 * the tree with the minimum value.
 *
 * To insert a node into a Fibonacci heap, a singleton tree is created and
 * merged into the rest of the trees.  The merge operation works by simply
 * splicing together the doubly-linked lists of the two trees, then updating
 * the min pointer to be the smaller of the minima of the two heaps.  Peeking
 * at the smallest element can therefore be accomplished by just looking at
 * the min element.  All of these operations complete in O(1) time.
 *
 * The tricky operations are dequeueMin and decreaseKey.  dequeueMin works
 * by removing the root of the tree containing the smallest element, then
 * merging its children with the topmost roots.  Then, the roots are scanned
 * and merged so that there is only one tree of each degree in the root list.
 * This works by maintaining a dynamic array of trees, each initially null,
 * pointing to the roots of trees of each dimension.  The list is then scanned
 * and this array is populated.  Whenever a conflict is discovered, the
 * appropriate trees are merged together until no more conflicts exist.  The
 * resulting trees are then put into the root list.  A clever analysis using
 * the potential method can be used to show that the amortized cost of this
 * operation is O(lg n), see "Introduction to Algorithms, Second Edition" by
 * Cormen, Rivest, Leiserson, and Stein for more details.
 *
 * The other hard operation is decreaseKey, which works as follows.  First, we
 * update the key of the node to be the new value.  If this leaves the node
 * smaller than its parent, we're done.  Otherwise, we cut the node from its
 * parent, add it as a root, and then mark its parent.  If the parent was
 * already marked, we cut that node as well, recursively mark its parent,
 * and continue this process.  This can be shown to run in O(1) amortized time
 * using yet another clever potential function.  Finally, given this function,
 * we can implement delete by decreasing a key to -\infty, then calling
 * dequeueMin to extract it.
 */

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
	f.size++
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

/**
 * Utility function which, given two pointers into disjoint circularly-
 * linked lists, merges the two lists together into one circularly-linked
 * list in O(1) time.  Because the lists may be empty, the return value
 * is the only pointer that's guaranteed to be to an element of the
 * resulting list.
 *
 * This function assumes that one and two are the minimum elements of the
 * lists they are in, and returns a pointer to whichever is smaller.  If
 * this condition does not hold, the return value is some arbitrary pointer
 * into the doubly-linked list.
 *
 * @param one A pointer into one of the two linked lists.
 * @param two A pointer into the other of the two linked lists.
 * @return A pointer to the smallest element of the resulting list.
 */
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
	f.size--

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

/**
 * Deletes this Entry from the Fibonacci heap that contains it.
 *
 * It is assumed that the entry belongs in this heap.  For efficiency
 * reasons, this is not checked at runtime.
 *
 * @param entry The entry to delete.
 */
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
	e.parent.degree--

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
