package pkg

import (
	"sync"
	"sync/atomic"
)

// SequenceGenerator 是一个简单的递增序列号生成器接口
type SequenceGenerator interface {
	// Next 返回下一个序列号
	Next() int
	// Current 返回当前序列号，不增加计数
	Current() int
	// Reset 重置计数器到指定值
	Reset(value int)
}

// atomicSequenceGenerator 是一个基于原子操作的线程安全的序列号生成器
type atomicSequenceGenerator struct {
	counter int64
}

// NewAtomicSequenceGenerator 创建一个新的基于原子操作的序列号生成器
// startValue 是初始值，默认从此值开始递增
func NewAtomicSequenceGenerator(startValue int) SequenceGenerator {
	return &atomicSequenceGenerator{
		counter: int64(startValue),
	}
}

// Next 原子性地递增并返回一个新的序列号
func (g *atomicSequenceGenerator) Next() int {
	return int(atomic.AddInt64(&g.counter, 1))
}

// Current 返回当前序列号，不增加计数
func (g *atomicSequenceGenerator) Current() int {
	return int(atomic.LoadInt64(&g.counter))
}

// Reset 重置计数器到指定值
func (g *atomicSequenceGenerator) Reset(value int) {
	atomic.StoreInt64(&g.counter, int64(value))
}

// mutexSequenceGenerator 是一个基于互斥锁的线程安全的序列号生成器
type mutexSequenceGenerator struct {
	counter int
	mu      sync.Mutex
}

// NewMutexSequenceGenerator 创建一个新的基于互斥锁的序列号生成器
// startValue 是初始值，默认从此值开始递增
func NewMutexSequenceGenerator(startValue int) SequenceGenerator {
	return &mutexSequenceGenerator{
		counter: startValue,
	}
}

// Next 使用互斥锁递增并返回一个新的序列号
func (g *mutexSequenceGenerator) Next() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.counter++
	return g.counter
}

// Current 返回当前序列号，不增加计数
func (g *mutexSequenceGenerator) Current() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.counter
}

// Reset 重置计数器到指定值
func (g *mutexSequenceGenerator) Reset(value int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.counter = value
}

// 默认的全局序列号生成器
var defaultSequenceGenerator = NewAtomicSequenceGenerator(0)

// NextSequence 获取下一个全局序列号
func NextSequence() int {
	return defaultSequenceGenerator.Next()
}

// CurrentSequence 获取当前全局序列号
func CurrentSequence() int {
	return defaultSequenceGenerator.Current()
}

// ResetSequence 重置全局序列号生成器
func ResetSequence(value int) {
	defaultSequenceGenerator.Reset(value)
}
