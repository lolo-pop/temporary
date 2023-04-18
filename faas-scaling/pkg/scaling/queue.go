package scaling

import (
	"fmt"
)

type Queue struct {
	items []int
	size  int
	front int
	rear  int
}

func NewQueue(size int) *Queue {
	return &Queue{
		items: make([]int, size),
		size:  size,
		front: 0,
		rear:  -1,
	}
}

func (q *Queue) Enqueue(item int) error {
	if q.isFull() {
		return fmt.Errorf("Queue is full")
	}
	q.rear++
	q.items[q.rear%q.size] = item
	return nil
}

func (q *Queue) Dequeue() (int, error) {
	if q.isEmpty() {
		return 0, fmt.Errorf("Queue is empty")
	}
	item := q.items[q.front%q.size]
	q.front++
	return item, nil
}

func (q *Queue) isFull() bool {
	return q.rear-q.front+1 == q.size
}

func (q *Queue) isEmpty() bool {
	return q.rear < q.front
}
