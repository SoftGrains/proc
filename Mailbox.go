package proc

import (
	"sync/atomic"
	"unsafe"
)

// Mailbox is a lock-free unbounded queue.
type Mailbox struct {
	head unsafe.Pointer
	tail unsafe.Pointer
}

type node struct {
	value interface{}
	next  unsafe.Pointer
}

// NewMailbox returns a pointer to an empty queue.
func NewMailbox() *Mailbox {
	n := unsafe.Pointer(&node{})

	return &Mailbox{
		head: n,
		tail: n,
	}
}

// Enqueue puts the given value v at the tail of the queue.
func (q *Mailbox) Enqueue(v interface{}) {
	n := &node{value: v}

	for {
		last := load(&q.tail)
		next := load(&last.next)
		if last == load(&q.tail) {
			if next == nil {
				if cas(&last.next, next, n) {
					cas(&q.tail, last, n)
					return
				}
			} else {
				cas(&q.tail, last, next)
			}
		}
	}
}

// Dequeue removes and returns the value at the head of the queue.
// It returns nil if the queue is empty.
func (q *Mailbox) Dequeue() (interface{}, bool) {
	for {
		first := load(&q.head)
		last := load(&q.tail)
		next := load(&first.next)
		if first == load(&q.head) {
			if first == last {
				if next == nil {
					return nil, false
				}
				cas(&q.tail, last, next)
			} else {
				v := next.value
				if cas(&q.head, first, next) {
					return v, true
				}
			}
		}
	}
}

func load(p *unsafe.Pointer) (n *node) {
	return (*node)(atomic.LoadPointer(p))
}

func store(p *unsafe.Pointer, n *node) {
	atomic.StorePointer(p, unsafe.Pointer(n))
}

func cas(p *unsafe.Pointer, old, new *node) bool {
	return atomic.CompareAndSwapPointer(p, unsafe.Pointer(old), unsafe.Pointer(new))
}
