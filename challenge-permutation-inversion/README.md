# Generate Original Permutation from Array of Inversions

## Problem

Given an array `arr[]` of size N, where `arr[i]` represents the number of elements on the left that are greater than the ith element in the original permutation, find the original permutation of [1, N].

## Approach

Segment trees are used for efficiently answering range queries and performing updates on arrays.

The solution uses a Segment Tree data structure to efficiently track available numbers and reconstruct the permutation:

1. Build a segment tree where each leaf represents whether a number (1 to N) is still available
2. Traverse the inversion array from right to left
3. For each position, find the kth available number where k = inversions[i] + 1
4. Mark that number as used and continue
```
# Level 0:                       [1]
#                          /             \
# Level 1:              [2]               [3]
#                     /     \           /      \
# Level 2:         [4]      [5]       [6]       [7]
#                 /  \     /   \     /   \     /   \
# Level 3:      [8] [9]  [10] [11] [12] [13] [14] [15]
```

## Input Format

Array of integers where each element represents the inversion count

## Output Format

Array of integers representing the original permutation

## Example 1
```
Input: [0, 1, 1, 0, 3]
Output: [4, 1, 3, 5, 2]

Explanation:
The original permutation is ans[] = {4, 1, 3, 5, 2}
ans[0] = 4.
ans[1] = 1. Since {4} exists on its left, which exceeds 1, arr[1] = 1 holds valid.
ans[2] = 3. Since {4} exists on its left, which exceeds 3, arr[2] = 1 holds valid.
ans[3] = 5. Since no elements on its left exceeds 5, arr[3] = 0 holds valid.
ans[4] = 2. Since {4, 3, 5} exists on its left, which exceeds 2, arr[4] = 3 holds valid.
```

## Example 2
```
Input: [0, 1, 2]
Output: [3, 2, 1]
```

## Example 3
```
Input: [0, 0, 0, 0]
Output: [1, 2, 3, 4]
```

## Running

```bash
go run main.go
```
