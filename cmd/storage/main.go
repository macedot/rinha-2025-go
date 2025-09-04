package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"

	"github.com/macedot/rinha-2025-go/internal/storage"
)

func main() {
	runtime.GOMAXPROCS(1)
	go func() {
		log.Println(http.ListenAndServe(":8888", nil))
	}()
	log.Fatalln(storage.Run())
}
