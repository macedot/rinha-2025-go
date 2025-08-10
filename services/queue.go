package services

import (
	"rinha-2025/models"
	"sync"
)

type Queue struct {
	items []models.Payment
	lock  sync.Mutex
	cond  *sync.Cond
}

func NewQueue() *Queue {
	q := &Queue{items: make([]models.Payment, 0, 20*1024)}
	q.cond = sync.NewCond(&q.lock)
	return q
}

func (q *Queue) Enqueue(item *models.Payment) {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.items = append(q.items, *item)
	q.cond.Signal()
}

func (q *Queue) Dequeue() models.Payment {
	q.lock.Lock()
	defer q.lock.Unlock()
	for len(q.items) == 0 {
		q.cond.Wait()
	}
	item := q.items[0]
	q.items = q.items[1:]
	return item
}
