package fixtures

import (
	"fmt"

	"go-template/internal/model"
)

type QueueTestData struct{}

func NewQueueTestData() *QueueTestData {
	return &QueueTestData{}
}

// ValidPppoeQueue returns a valid PPPoE queue for testing
func (q *QueueTestData) ValidPppoeQueue() model.PppoeQueue {
	return model.PppoeQueue{
		ID:             "1",
		Name:           "test-queue",
		Target:         "192.168.1.100/32",
		MaxLimit:       "10M/10M",
		BurstLimit:     "20M/20M",
		BurstThreshold: "8M/8M",
		BurstTime:      "30s",
		Priority:       "8",
		Parent:         "global-out",
		Comment:        "Test Queue",
	}
}

// ValidQueueStats returns valid queue statistics for testing
func (q *QueueTestData) ValidQueueStats() model.QueueStats {
	return model.QueueStats{
		Name:          "test-queue",
		BytesIn:       1024000000,
		BytesOut:      512000000,
		PacketsIn:     10000000,
		PacketsOut:    5000000,
		RateIn:        1000000,
		RateOut:       500000,
		PacketRateIn:  10000,
		PacketRateOut: 5000,
	}
}

// MultiplePppoeQueues returns multiple PPPoE queues for testing
func (q *QueueTestData) MultiplePppoeQueues(count int) []model.PppoeQueue {
	queues := make([]model.PppoeQueue, count)

	for i := 0; i < count; i++ {
		priority := fmt.Sprintf("%d", 8-i%8)
		queues[i] = model.PppoeQueue{
			ID:             fmt.Sprintf("%d", i+1),
			Name:           fmt.Sprintf("queue-%d", i+1),
			Target:         fmt.Sprintf("192.168.1.%d/32", i+100),
			MaxLimit:       fmt.Sprintf("%dM/%dM", 10-i, 10-i),
			BurstLimit:     fmt.Sprintf("%dM/%dM", 20-i, 20-i),
			BurstThreshold: fmt.Sprintf("%dM/%dM", 8-i, 8-i),
			BurstTime:      "30s",
			Priority:       priority,
			Parent:         "global-out",
			Comment:        fmt.Sprintf("Queue %d", i+1),
		}
	}

	return queues
}

// MultipleQueueStats returns multiple queue statistics for testing
func (q *QueueTestData) MultipleQueueStats(count int) []model.QueueStats {
	stats := make([]model.QueueStats, count)

	for i := 0; i < count; i++ {
		stats[i] = model.QueueStats{
			Name:          fmt.Sprintf("queue-%d", i+1),
			BytesIn:       int64(1024000000 + i*100000000),
			BytesOut:      int64(512000000 + i*50000000),
			PacketsIn:     int64(10000000 + i*1000000),
			PacketsOut:    int64(5000000 + i*500000),
			RateIn:        int64(1000000 + i*100000),
			RateOut:       int64(500000 + i*50000),
			PacketRateIn:  int64(10000 + i*1000),
			PacketRateOut: int64(5000 + i*500),
		}
	}

	return stats
}

// PppoeQueueWithTarget returns a queue with specific target
func (q *QueueTestData) PppoeQueueWithTarget(target string) model.PppoeQueue {
	queue := q.ValidPppoeQueue()
	queue.Target = target
	return queue
}

// PppoeQueueWithName returns a queue with specific name
func (q *QueueTestData) PppoeQueueWithName(name string) model.PppoeQueue {
	queue := q.ValidPppoeQueue()
	queue.Name = name
	return queue
}

// PppoeQueueWithMaxLimit returns a queue with specific max limit
func (q *QueueTestData) PppoeQueueWithMaxLimit(maxLimit string) model.PppoeQueue {
	queue := q.ValidPppoeQueue()
	queue.MaxLimit = maxLimit
	return queue
}

// PppoeQueueWithPriority returns a queue with specific priority
func (q *QueueTestData) PppoeQueueWithPriority(priority string) model.PppoeQueue {
	queue := q.ValidPppoeQueue()
	queue.Priority = priority
	return queue
}

// HighPriorityQueue returns a high priority queue (priority 1)
func (q *QueueTestData) HighPriorityQueue() model.PppoeQueue {
	return q.PppoeQueueWithPriority("1")
}

// LowPriorityQueue returns a low priority queue (priority 8)
func (q *QueueTestData) LowPriorityQueue() model.PppoeQueue {
	return q.PppoeQueueWithPriority("8")
}

// QueueStatsWithName returns queue stats for specific queue name
func (q *QueueTestData) QueueStatsWithName(name string) model.QueueStats {
	stats := q.ValidQueueStats()
	stats.Name = name
	return stats
}
