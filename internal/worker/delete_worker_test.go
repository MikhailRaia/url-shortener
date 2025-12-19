package worker

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockDeleteService struct {
	mu          sync.Mutex
	deleteCalls []DeleteCall
	deleteDelay time.Duration
	shouldFail  bool
	callCount   atomic.Int32
}

type DeleteCall struct {
	UserID string
	URLIDs []string
}

func (m *MockDeleteService) DeleteUserURLs(userID string, urlIDs []string) error {
	m.callCount.Add(1)

	if m.deleteDelay > 0 {
		time.Sleep(m.deleteDelay)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.deleteCalls = append(m.deleteCalls, DeleteCall{
		UserID: userID,
		URLIDs: urlIDs,
	})

	if m.shouldFail {
		return assert.AnError
	}

	return nil
}

func (m *MockDeleteService) GetCalls() []DeleteCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]DeleteCall{}, m.deleteCalls...)
}

func (m *MockDeleteService) GetCallCount() int {
	return int(m.callCount.Load())
}

func (m *MockDeleteService) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deleteCalls = nil
	m.callCount.Store(0)
}

func TestNewDeleteWorkerPool(t *testing.T) {
	service := &MockDeleteService{}
	config := DefaultConfig()

	pool := NewDeleteWorkerPool(service, config)

	assert.NotNil(t, pool)
	assert.Equal(t, config.WorkerCount, pool.workerCount)
	assert.Equal(t, config.BatchSize, pool.batchSize)
	assert.Equal(t, config.BatchTimeout, pool.batchTimeout)
	assert.NotNil(t, pool.requestChan)
	assert.Equal(t, config.BufferSize, cap(pool.requestChan))
}

func TestDeleteWorkerPool_SingleRequest(t *testing.T) {
	service := &MockDeleteService{}
	config := Config{
		WorkerCount:  2,
		BufferSize:   10,
		BatchSize:    5,
		BatchTimeout: 100 * time.Millisecond,
	}

	pool := NewDeleteWorkerPool(service, config)
	pool.Start()
	defer pool.Shutdown(time.Second)

	err := pool.Submit("user1", []string{"url1", "url2"})
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	calls := service.GetCalls()
	require.Len(t, calls, 1)
	assert.Equal(t, "user1", calls[0].UserID)
	assert.ElementsMatch(t, []string{"url1", "url2"}, calls[0].URLIDs)
}

func TestDeleteWorkerPool_BatchProcessing(t *testing.T) {
	service := &MockDeleteService{}
	config := Config{
		WorkerCount:  2,
		BufferSize:   50,
		BatchSize:    10,
		BatchTimeout: 5 * time.Second,
	}

	pool := NewDeleteWorkerPool(service, config)
	pool.Start()
	defer pool.Shutdown(time.Second)

	err := pool.Submit("user1", []string{"url1", "url2", "url3"})
	require.NoError(t, err)

	err = pool.Submit("user1", []string{"url4", "url5"})
	require.NoError(t, err)

	err = pool.Submit("user2", []string{"url6", "url7", "url8", "url9", "url10"})
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	calls := service.GetCalls()
	require.NotEmpty(t, calls)

	totalURLs := 0
	for _, call := range calls {
		totalURLs += len(call.URLIDs)
	}
	assert.Equal(t, 10, totalURLs)
}

func TestDeleteWorkerPool_MultipleUsers(t *testing.T) {
	service := &MockDeleteService{}
	config := Config{
		WorkerCount:  3,
		BufferSize:   50,
		BatchSize:    20,
		BatchTimeout: 100 * time.Millisecond,
	}

	pool := NewDeleteWorkerPool(service, config)
	pool.Start()
	defer pool.Shutdown(time.Second)

	users := []string{"user1", "user2", "user3"}
	for _, userID := range users {
		err := pool.Submit(userID, []string{"url1", "url2", "url3"})
		require.NoError(t, err)
	}

	time.Sleep(200 * time.Millisecond)

	calls := service.GetCalls()
	require.NotEmpty(t, calls)

	processedUsers := make(map[string]bool)
	for _, call := range calls {
		processedUsers[call.UserID] = true
	}

	for _, userID := range users {
		assert.True(t, processedUsers[userID], "User %s should be processed", userID)
	}
}

func TestDeleteWorkerPool_ConcurrentSubmits(t *testing.T) {
	service := &MockDeleteService{}
	config := Config{
		WorkerCount:  5,
		BufferSize:   100,
		BatchSize:    50,
		BatchTimeout: 200 * time.Millisecond,
	}

	pool := NewDeleteWorkerPool(service, config)
	pool.Start()
	defer pool.Shutdown(2 * time.Second)

	const goroutines = 10
	const requestsPerGoroutine = 5

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < requestsPerGoroutine; j++ {
				userID := "user" + string(rune('0'+id))
				err := pool.Submit(userID, []string{"url1", "url2"})
				assert.NoError(t, err)
			}
		}(i)
	}

	wg.Wait()

	time.Sleep(500 * time.Millisecond)

	calls := service.GetCalls()
	assert.NotEmpty(t, calls)

	totalURLs := 0
	for _, call := range calls {
		totalURLs += len(call.URLIDs)
	}

	expectedURLs := goroutines * requestsPerGoroutine * 2
	assert.Equal(t, expectedURLs, totalURLs)
}

func TestDeleteWorkerPool_GracefulShutdown(t *testing.T) {
	service := &MockDeleteService{
		deleteDelay: 50 * time.Millisecond,
	}
	config := Config{
		WorkerCount:  2,
		BufferSize:   10,
		BatchSize:    5,
		BatchTimeout: 100 * time.Millisecond,
	}

	pool := NewDeleteWorkerPool(service, config)
	pool.Start()

	for i := 0; i < 3; i++ {
		err := pool.Submit("user1", []string{"url1", "url2"})
		require.NoError(t, err)
	}

	err := pool.Shutdown(2 * time.Second)
	assert.NoError(t, err)

	calls := service.GetCalls()
	assert.NotEmpty(t, calls)
}

func TestDeleteWorkerPool_Stats(t *testing.T) {
	service := &MockDeleteService{}
	config := Config{
		WorkerCount:  3,
		BufferSize:   50,
		BatchSize:    10,
		BatchTimeout: 1 * time.Second,
	}

	pool := NewDeleteWorkerPool(service, config)
	pool.Start()
	defer pool.Shutdown(time.Second)

	for i := 0; i < 5; i++ {
		err := pool.Submit("user1", []string{"url1"})
		require.NoError(t, err)
	}

	stats := pool.Stats()
	assert.Equal(t, 3, stats.WorkerCount)
	assert.Equal(t, 50, stats.QueueCap)
	assert.LessOrEqual(t, stats.QueueSize, 5)
}

func TestDeleteWorkerPool_ErrorHandling(t *testing.T) {
	service := &MockDeleteService{
		shouldFail: true, // Имитируем ошибки
	}
	config := Config{
		WorkerCount:  2,
		BufferSize:   10,
		BatchSize:    5,
		BatchTimeout: 100 * time.Millisecond,
	}

	pool := NewDeleteWorkerPool(service, config)
	pool.Start()
	defer pool.Shutdown(time.Second)

	err := pool.Submit("user1", []string{"url1", "url2"})
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	calls := service.GetCalls()
	require.Len(t, calls, 1)
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, 5, config.WorkerCount)
	assert.Equal(t, 100, config.BufferSize)
	assert.Equal(t, 10, config.BatchSize)
	assert.Equal(t, 5*time.Second, config.BatchTimeout)
}
