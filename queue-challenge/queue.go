package main

// Queue implements a custom generic queue data structure using a singly-linked list
type Queue[T any] struct {
	head *Node[T]
	tail *Node[T]
	size int
}

// Node represents a generic node in the queue's linked list
type Node[T any] struct {
	data T
	next *Node[T]
}

// NewQueue creates and returns a new empty queue
func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{
		head: nil,
		tail: nil,
		size: 0,
	}
}

// Enqueue adds an element to the back of the queue
func (q *Queue[T]) Enqueue(data T) {
	newNode := &Node[T]{
		data: data,
		next: nil,
	}

	if q.tail != nil {
		q.tail.next = newNode
	}

	q.tail = newNode

	if q.head == nil {
		q.head = newNode
	}

	q.size++
}

// Dequeue removes and returns the element from the front of the queue
func (q *Queue[T]) Dequeue() (T, bool) {
	var zero T
	if q.IsEmpty() {
		return zero, false
	}

	data := q.head.data
	q.head = q.head.next

	if q.head == nil {
		// Queue is now empty
		q.tail = nil
	}

	q.size--
	return data, true
}

// Peek returns the element at the front without removing it
func (q *Queue[T]) Peek() (T, bool) {
	var zero T
	if q.IsEmpty() {
		return zero, false
	}
	return q.head.data, true
}

// IsEmpty returns true if the queue has no elements
func (q *Queue[T]) IsEmpty() bool {
	return q.size == 0
}

// ForEach iterates through all elements in the queue and applies a function
func (q *Queue[T]) ForEach(fn func(T)) {
	current := q.head
	for current != nil {
		fn(current.data)
		current = current.next
	}
}
