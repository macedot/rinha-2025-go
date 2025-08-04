package services

import (
	"log"
	"net/http"
	"rinha-2025/config"
	"rinha-2025/database"
	"sync"
)

func PurgePayments() error {
	var wg sync.WaitGroup
	services := config.ConfigInstance().GetServices()
	for _, service := range services {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := purgePaymentProcessor(&service)
			if err != nil {
				log.Print("purgePaymentProcessor:", err)
			}
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		db := database.RedisInstance()
		err := db.FlushAll()
		if err != nil {
			log.Print("FlushAll:", err)
		}
	}()
	wg.Wait()
	return nil
}

func purgePaymentProcessor(instance *config.Service) error {
	paymentURL := instance.URL + "/admin/purge-payments"
	req, err := http.NewRequest("POST", paymentURL, nil)
	if err == nil {
		req.Header.Set("X-Rinha-Token", instance.Token)
		client := &http.Client{Timeout: 0}
		_, err = client.Do(req)
	}
	return err
}
