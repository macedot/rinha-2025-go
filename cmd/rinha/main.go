package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"

	"github.com/macedot/rinha-2025-go/internal/api"
	"github.com/macedot/rinha-2025-go/internal/worker"
)

func main() {
	runtime.GOMAXPROCS(1)
	go func() {
		log.Println(http.ListenAndServe(":8888", nil))
	}()
	go func() {
		log.Fatalln(worker.Run())
	}()
	log.Fatalln(api.Run())
}
