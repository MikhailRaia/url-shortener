package pool

import "sync"

// Resettable is a constraint for types that have a Reset() method.
type Resettable interface {
	Reset()
}

// Poolable is a constraint for types that can be pooled (must be resettable and comparable).
type Poolable interface {
	Resettable
	comparable
}

// Pool is a generic object pool for storing and reusing objects of type T.
// T must satisfy the Poolable constraint (have a Reset() method and be comparable).
type Pool[T Poolable] struct {
	mu    sync.Mutex
	items chan T
}

// New creates and returns a new Pool for objects of type T.
// The capacity parameter specifies the maximum number of objects the pool can hold.
func New[T Poolable](capacity int) *Pool[T] {
	return &Pool[T]{
		items: make(chan T, capacity),
	}
}

// Get retrieves an object from the pool.
// If the pool is empty, it returns the zero value of type T.
// If an object is retrieved from the pool, it is returned as-is.
func (p *Pool[T]) Get() T {
	select {
	case item := <-p.items:
		return item
	default:
		var zero T
		return zero
	}
}

// Put returns an object to the pool after calling its Reset() method.
// If the pool is full, the object is discarded.
// This ensures that objects are reset before being reused.
func (p *Pool[T]) Put(item T) {
	var zero T
	if item != zero {
		item.Reset()
	}

	select {
	case p.items <- item:
	default:
	}
}
