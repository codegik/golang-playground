package main

import (
	"fmt"
)

// Leaf nodes don't store the numbers, they store availability (0 or 1)
// The leaf's position in the range [1,N] represents the actual number
// When query returns left, it's returning the position, which equals the number
// The tree structure maps: position → availability status
type SegmentTree struct {
	tree []int
	size int
}

// A segment tree needs its size to be a power of 2 because.
// Without power-of-2 sizing, the tree becomes unbalanced and indexing formulas break down.
func NewSegmentTree(n int) *SegmentTree {
	size := 1
	for size < n {
		size *= 2
	}
	return &SegmentTree{
		tree: make([]int, 2*size),
		size: size,
	}
}

// GO has feature Method Call Syntax, which allows us to call methods on sequences like `tree.build(...)`
// This function initializes the segment tree where each node stores
// the count of available numbers in its range.
func (st *SegmentTree) build(length, node, left, right int) {
	// When left == right, we're at a leaf representing a single number
	if left == right {
		// If that number ≤ length, it's valid and available (set to 1)
		// If that number > length, it's padding for power-of-2 sizing (stays 0)
		if left <= length {
			st.tree[node] = 1
		}
		return
	}
	mid := (left + right) / 2
	// Why multiply by 2? Because 2 * node is the left child's index in the 1-indexed segment tree array.
	st.build(length, 2*node, left, mid)                 // build left subtree
	st.build(length, 2*node+1, mid+1, right)            // build right subtree
	st.tree[node] = st.tree[2*node] + st.tree[2*node+1] // sum children, In a segment tree, each node represents a range
}

// This function finds the k-th smallest available number in the segment tree.
// k: which available number to find (1st, 2nd, 3rd, etc.)
// node: current position in tree array
// left, right: the range this node represents

// Suppose available = [1,2,3,5] (4 was removed), and we want k=3:
// Query(k=3, node=1, range=[1,8])
// ├─ left child has 2 available (1 and 2)
// ├─ 2 < 3, so k=3 is NOT in left subtree
// └─ Go right with k = 3-2 = 1 (find 1st available in right)
//
//	Query(k=1, node=3, range=[5,8])
//	├─ left child has 1 available (5)
//	├─ 1 >= 1, so k=1 IS in left subtree
//	└─ Go left with k=1
//	   Query(k=1, node=6, range=[5,6])
//	   Eventually returns: 5
func (st *SegmentTree) query(k, node, left, right int) int {
	// When we reach a leaf, that's our answer - the k-th available number
	if left == right {
		return left
	}
	// If left subtree has ≥ k available: the k-th smallest must be in left subtree, recurse left with same
	mid := (left + right) / 2
	if st.tree[2*node] >= k {
		return st.query(k, 2*node, left, mid)
	}
	// Otherwise: skip the left subtree entirely and search right, but adjust k by subtracting left count
	return st.query(k-st.tree[2*node], 2*node+1, mid+1, right)
}

// This function marks a number as used (no longer available) by setting it to 0,
// then propagates the change up the tree.
func (st *SegmentTree) update(pos, node, left, right int) {
	// Found the leaf
	if left == right {
		st.tree[node] = 0
		return
	}
	mid := (left + right) / 2
	if pos <= mid {
		st.update(pos, 2*node, left, mid) // navigate left
	} else {
		st.update(pos, 2*node+1, mid+1, right) // navigate right
	}
	// After recursion returns, recalculate this node's count from children
	st.tree[node] = st.tree[2*node] + st.tree[2*node+1]
}

// how many larger numbers are to the LEFT of position i?
// The formula i - inversions[i] + 1:
// i + 1 = total positions processed so far (including current)
// Subtract inversions[i] = numbers that must be larger than result[i]
// Result = rank of the number we need among remaining available numbers

// Concrete walkthrough with inversions = [0, 1, 1, 0, 3]:
// Initially available: [1, 2, 3, 4, 5]
// Result: [_, _, _, _, _]

// i=4: inversions[4]=3
// k = 4 - 3 + 1 = 2 → find 2nd available number
// query returns: 2
// result = [_, _, _, _, 2]
// remove 2 → available: [1, 3, 4, 5]

// i=3: inversions[3]=0
// k = 3 - 0 + 1 = 4 → find 4th available number
// query returns: 5
// result = [_, _, _, 5, 2]
// remove 5 → available: [1, 3, 4]

// i=2: inversions[2]=1
// k = 2 - 1 + 1 = 2 → find 2nd available number
// query returns: 3
// result = [_, _, 3, 5, 2]
// remove 3 → available: [1, 4]

// i=1: inversions[1]=1
// k = 1 - 1 + 1 = 1 → find 1st available number
// query returns: 1
// result = [_, 1, 3, 5, 2]
// remove 1 → available: [4]

// i=0: inversions[0]=0
// k = 0 - 0 + 1 = 1 → find 1st available number
// query returns: 4
// result = [4, 1, 3, 5, 2]
// remove 4 → available: []

// Final result: [4, 1, 3, 5, 2]
func generatePermutation(inversions []int) []int {
	length := len(inversions)
	result := make([]int, length)

	st := NewSegmentTree(length)
	st.build(length, 1, 1, st.size)

	for i := length - 1; i >= 0; i-- {
		pos := st.query(i-inversions[i]+1, 1, 1, st.size)
		result[i] = pos
		st.update(pos, 1, 1, st.size)
	}

	return result
}

func main() {
	inversions := []int{0, 1, 1, 0, 3}
	result := generatePermutation(inversions)
	fmt.Printf("Input: %v\nOutput: %v\n", inversions, result)
}
