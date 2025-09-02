package services

import (
	"rinha-2025/config"
	"rinha-2025/database"
	"rinha-2025/models"
	"time"

	"github.com/ohler55/ojg/oj"
)

func EnqueuePayment(payment *models.Payment, queue *Queue) {
	queue.Enqueue(payment)
}

func ProcessPayment(payment *models.Payment) error {
	var instances []config.Service
	for {
		instances = GetInstancesCache()
		if len(instances) > 0 {
			break
		}
		time.Sleep(time.Millisecond)
	}
	payload, err := oj.Marshal(payment)
	if err != nil {
		return err
	}
	return forwardPayment(&instances[0], payment, payload)
}

func forwardPayment(instance *config.Service, payment *models.Payment, payload []byte) error {
	client := HttpClientInstance()
	db := database.RedisInstance()
	payment.Timestamp = time.Now().UTC()
	if err := client.Post(instance.URL+"/payments", payload); err != nil {
		return err
	}
	db.SavePayment(instance, payment)
	return nil
}
