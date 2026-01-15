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
// It works similarly to sync.Pool but with automatic Reset() call on Put.
// T must satisfy the Poolable constraint (have a Reset() method and be comparable).
type Pool[T Poolable] struct {
	pool sync.Pool
}

// New creates and returns a new Pool for objects of type T.
// Unlike the previous implementation, this pool has no capacity limit and can grow dynamically,
// similar to the standard library's sync.Pool.
func New[T Poolable](capacity int) *Pool[T] {
	return &Pool[T]{}
}

// Get retrieves an object from the pool.
// If the pool is empty, it returns the zero value of type T.
func (p *Pool[T]) Get() T {
	val := p.pool.Get()
	if val == nil {
		var zero T
		return zero
	}
	return val.(T)
}

// Put returns an object to the pool after calling its Reset() method.
// Objects in the pool may be automatically deleted during garbage collection,
// similar to sync.Pool behavior.
func (p *Pool[T]) Put(item T) {
	var zero T
	if item != zero {
		item.Reset()
		p.pool.Put(item)
	}
}
