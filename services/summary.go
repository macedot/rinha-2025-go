package services

import (
	"fmt"
	"log"
	"rinha-2025/config"
	"rinha-2025/database"
	"rinha-2025/models"
	"strconv"
	"sync"
	"time"
)

func GetSummary(summaryReg *models.SummaryRequest) (models.SummaryResponse, error) {
	res := models.SummaryResponse{
		Default:  models.ProcessorSummary{},
		Fallback: models.ProcessorSummary{},
	}
	param, err := processSummary(summaryReg)
	if err != nil {
		return res, err
	}
	services := config.ConfigInstance().GetServices()
	db := database.RedisInstance()
	var wg sync.WaitGroup
	wg.Add(1)
	start := time.Now()
	go func() {
		defer wg.Done()
		res.Default, err = db.GetSummary(&services[0], &param)
		if err != nil {
			log.Println("GetSummary:0:", err.Error())
		}
	}()
	if len(services) > 1 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res.Fallback, err = db.GetSummary(&services[1], &param)
			if err != nil {
				log.Println("GetSummary:1:", err.Error())
			}
		}()
	}
	wg.Wait()
	log.Print("GetSummary:", time.Since(start))
	return res, nil
}

func processSummary(summaryReg *models.SummaryRequest) (models.SummaryParam, error) {
	var res models.SummaryParam
	var err error
	if res.StartTime, err = processTime(summaryReg.StartTime, "-inf"); err != nil {
		return res, fmt.Errorf("invalid start time format")
	}
	if res.EndTime, err = processTime(summaryReg.EndTime, "+inf"); err != nil {
		return res, fmt.Errorf("invalid end time format")
	}
	return res, nil
}

func processTime(param string, value string) (string, error) {
	if param == "" {
		return value, nil
	}
	if timeValue, err := time.Parse(time.RFC3339, param); err == nil {
		ts := float64(timeValue.UTC().UnixNano()) / 1e9
		return strconv.FormatFloat(ts, 'f', -1, 64), nil
	}
	return param, fmt.Errorf("invalid end time format")
}
