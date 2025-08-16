package main

import (
	"log"
	"os"
	"path/filepath"
	"rinha-storage/handler"
	"rinha-storage/store"

	"github.com/valyala/fasthttp"
)

func NewSocket(socketPath string) {
	socketDir := filepath.Dir(socketPath)
	if err := os.MkdirAll(socketDir, 0777); err != nil {
		log.Fatalf("Failed to create socket directory: %v", err)
	}
	if err := os.RemoveAll(socketPath); err != nil {
		log.Fatalf("Failed to remove existing socket: %v", err)
	}
	fp, err := os.Create(socketPath)
	if err != nil {
		log.Fatalf("Failed to create socket file: %v", err)
	}
	fp.Close()
}

func main() {
	storageSocket := os.Getenv("STORAGE_SOCKET")
	if storageSocket == "" {
		log.Fatalln("STORAGE_SOCKET environment variable not set")
	}

	NewSocket(storageSocket)

	// fileDB := store.NewFileDB("./payments.json1")
	// defer fileDB.Close()

	database := store.NewMemoryDB()

	// if records, err := fileDB.LoadRecords(); err == nil {
	// 	for _, record := range records {
	// 		database.AddRecord(record)
	// 	}
	// } else {
	// 	log.Printf("Failed to load initial records: %v", err)
	// }

	paymentHandler := handler.PaymentHandler(database)
	summaryHandler := handler.SummaryHandler(database)
	purgePaymentsHandler := handler.PurgePaymentsHandler(database)

	requestHandler := func(ctx *fasthttp.RequestCtx) {
		path := string(ctx.Path())
		switch path {
		case "/payments":
			paymentHandler(ctx)
		case "/payments-summary":
			summaryHandler(ctx)
		case "/purge-payments":
			purgePaymentsHandler(ctx)
		default:
			ctx.Error(path, fasthttp.StatusNotFound)
		}
	}

	log.Printf("Listening on %s", storageSocket)
	log.Fatal(fasthttp.ListenAndServeUNIX(storageSocket, 0666, requestHandler))
}
