package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}
	db := SQL(os.Getenv("PG_DSN"))

	redis_db, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		log.Fatalf("redis db parsing error: %v", err)
	}
	rdb := Redis(redis_db)

	handler := NewRecordHandler(db, rdb, log.Default())
	router := http.NewServeMux()
	RegisterRoutes(router, handler)

	srv := http.Server{
		Addr:         "localhost:4000",
		Handler:      router,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	go func ()  {
		
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	interruptCh := make(chan os.Signal, 1)
	signal.Notify(interruptCh, os.Interrupt, syscall.SIGTERM)

	<- interruptCh

}
