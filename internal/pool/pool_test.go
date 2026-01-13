package pool

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockResettable struct {
	Value       int
	Name        string
	ResetCalled int
}

func (m *mockResettable) Reset() {
	m.Value = 0
	m.Name = ""
	m.ResetCalled++
}

func TestNewPool(t *testing.T) {
	pool := New[*mockResettable](5)
	require.NotNil(t, pool)
}

func TestPoolGet_EmptyPool(t *testing.T) {
	pool := New[*mockResettable](5)
	item := pool.Get()

	assert.Nil(t, item)
}

func TestPoolPutAndGet(t *testing.T) {
	pool := New[*mockResettable](5)

	obj := &mockResettable{Value: 42, Name: "test"}
	pool.Put(obj)

	retrieved := pool.Get()
	assert.NotNil(t, retrieved)
	assert.Equal(t, 0, retrieved.Value)
	assert.Equal(t, "", retrieved.Name)
	assert.Equal(t, 1, retrieved.ResetCalled)
}

func TestPoolResetOnPut(t *testing.T) {
	pool := New[*mockResettable](5)

	obj := &mockResettable{Value: 100, Name: "original"}
	assert.Equal(t, 0, obj.ResetCalled)

	pool.Put(obj)

	assert.Equal(t, 1, obj.ResetCalled)
	assert.Equal(t, 0, obj.Value)
	assert.Equal(t, "", obj.Name)
}

func TestPoolMultipleItems(t *testing.T) {
	pool := New[*mockResettable](3)

	obj1 := &mockResettable{Value: 1, Name: "first"}
	obj2 := &mockResettable{Value: 2, Name: "second"}
	obj3 := &mockResettable{Value: 3, Name: "third"}

	pool.Put(obj1)
	pool.Put(obj2)
	pool.Put(obj3)

	retrieved1 := pool.Get()
	retrieved2 := pool.Get()
	retrieved3 := pool.Get()
	empty := pool.Get()

	assert.NotNil(t, retrieved1)
	assert.NotNil(t, retrieved2)
	assert.NotNil(t, retrieved3)
	assert.Nil(t, empty)

	assert.Equal(t, 1, retrieved1.ResetCalled)
	assert.Equal(t, 1, retrieved2.ResetCalled)
	assert.Equal(t, 1, retrieved3.ResetCalled)
}

func TestPoolCapacityOverflow(t *testing.T) {
	pool := New[*mockResettable](2)

	obj1 := &mockResettable{Value: 1}
	obj2 := &mockResettable{Value: 2}
	obj3 := &mockResettable{Value: 3}

	pool.Put(obj1)
	pool.Put(obj2)
	pool.Put(obj3)

	retrieved1 := pool.Get()
	retrieved2 := pool.Get()
	retrieved3 := pool.Get()

	assert.NotNil(t, retrieved1)
	assert.NotNil(t, retrieved2)
	assert.Nil(t, retrieved3)
}

func TestPoolNilHandling(t *testing.T) {
	pool := New[*mockResettable](5)

	var nilObj *mockResettable
	pool.Put(nilObj)

	retrieved := pool.Get()
	assert.Nil(t, retrieved)
}

func TestPoolReuse(t *testing.T) {
	pool := New[*mockResettable](5)

	obj := &mockResettable{Value: 42, Name: "test"}
	pool.Put(obj)

	retrieved := pool.Get()
	assert.NotNil(t, retrieved)
	assert.Equal(t, 0, retrieved.Value)
	assert.Equal(t, "", retrieved.Name)

	retrieved.Value = 99
	retrieved.Name = "modified"
	pool.Put(retrieved)

	reused := pool.Get()
	assert.Equal(t, 0, reused.Value)
	assert.Equal(t, "", reused.Name)
	assert.Equal(t, 2, reused.ResetCalled)
}
