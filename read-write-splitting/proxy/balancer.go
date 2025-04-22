package proxy

import (
	"sync"
	"sync/atomic"
)

// Balancer 负载均衡器接口
type Balancer interface {
	// Next 获取下一个可用的Redis从库索引
	Next(slaveCount int) int

	// MarkDown 标记某个从库为不可用
	MarkDown(index int)

	// MarkUp 标记某个从库为可用
	MarkUp(index int)
}

// RoundRobinBalancer 实现简单的轮询负载均衡
type RoundRobinBalancer struct {
	counter  uint64       // 请求计数器，用于轮询
	status   []bool       // 各从库状态，true表示可用
	statusMu sync.RWMutex // 保护status的互斥锁
}

// NewRoundRobinBalancer 创建一个新的轮询负载均衡器
func NewRoundRobinBalancer(slaveCount int) *RoundRobinBalancer {
	status := make([]bool, slaveCount)
	// 初始化所有从库状态为可用
	for i := range status {
		status[i] = true
	}

	return &RoundRobinBalancer{
		counter: 0,
		status:  status,
	}
}

// Next 获取下一个可用的Redis从库索引
// 如果所有从库都不可用，返回-1
func (b *RoundRobinBalancer) Next(slaveCount int) int {
	// 增加计数器并获取当前值
	current := atomic.AddUint64(&b.counter, 1) - 1

	b.statusMu.RLock()
	defer b.statusMu.RUnlock()

	// 尝试slaveCount次，确保我们考虑了所有可能的从库
	for i := 0; i < slaveCount; i++ {
		// 计算当前应该使用的从库索引（轮询）
		index := int((current + uint64(i)) % uint64(slaveCount))

		// 检查该从库是否可用
		if b.status[index] {
			return index
		}
	}

	// 所有从库都不可用
	return -1
}

// MarkDown 标记某个从库为不可用
func (b *RoundRobinBalancer) MarkDown(index int) {
	b.statusMu.Lock()
	defer b.statusMu.Unlock()

	if index >= 0 && index < len(b.status) {
		b.status[index] = false
	}
}

// MarkUp 标记某个从库为可用
func (b *RoundRobinBalancer) MarkUp(index int) {
	b.statusMu.Lock()
	defer b.statusMu.Unlock()

	if index >= 0 && index < len(b.status) {
		b.status[index] = true
	}
}
