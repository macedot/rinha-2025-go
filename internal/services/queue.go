package services

import (
	"rinha-2025-go/internal/models"
)

type PaymentQueue struct {
	items chan *models.Payment
}

func NewPaymentQueue(size int) *PaymentQueue {
	return &PaymentQueue{items: make(chan *models.Payment, size)}
}

func (q *PaymentQueue) Enqueue(item *models.Payment) {
	q.items <- item
}

func (q *PaymentQueue) Dequeue() *models.Payment {
	return <-q.items
}

func (q *PaymentQueue) Length() int {
	return len(q.items)
}

func (q *PaymentQueue) Close() {
	close(q.items)
}
