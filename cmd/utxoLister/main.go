package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"time"

	"github.com/matterinc/PlasmaBlockCreator/foundationdb"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	redis "github.com/go-redis/redis"
	configs "github.com/matterinc/PlasmaBlockCreator/configs"
	handlers "github.com/matterinc/PlasmaBlockCreator/handlers"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/reuseport"
)

func main() {
	fdb.MustAPIVersion(520)

	httpConfig, redisConfig, concurrencyConfig, databaseConfig, _, err := configs.ParseConfigs()
	if err != nil {
		log.Printf("%+v\n", err)
		os.Exit(1)
	}

	// Init foundationDB

	foundDB, err := configs.InitDB(databaseConfig)
	if err != nil {
		log.Printf("%+v\n", err)
		os.Exit(1)
	}

	// Init redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisConfig.RedisHost + ":" + strconv.Itoa(redisConfig.RedisPort),
		Password: redisConfig.RedisPassword,
		DB:       0,
	})
	defer redisClient.Close()

	// Test for linearizability
	fmt.Println("Testing for the counter in Redis")
	redisCounter, err := redisClient.Get("ctr").Uint64()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	fmt.Println("Redis counter = ", strconv.FormatUint(redisCounter, 10))

	fmt.Println("Testing for the database")
	maxCounterInDatabase, err := foundationdb.GetMaxTransactionCounter(foundDB)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	fmt.Println("Database counter = ", strconv.FormatUint(maxCounterInDatabase, 10))

	if maxCounterInDatabase > redisCounter {
		log.Println("Counters mismatch")
		os.Exit(1)
	}

	ECRecoverConcurrency := concurrencyConfig.ECRecoverConcurrency
	if ECRecoverConcurrency == -1 {
		ECRecoverConcurrency = concurrencyConfig.MaxProc
		if ECRecoverConcurrency == -1 {
			ECRecoverConcurrency = runtime.NumCPU()
		}

	}
	DatabaseConcurrency := concurrencyConfig.DatabaseConcurrency
	if DatabaseConcurrency < 0 {
		log.Println("FoundationDB concurrency should be > 0")
		os.Exit(1)
	}

	fmt.Println("ECRecover concurrency = " + strconv.Itoa(ECRecoverConcurrency))
	fmt.Println("FDB concurrency = " + strconv.Itoa(DatabaseConcurrency))

	listUTXOsHandler := handlers.NewListUTXOsHandler(foundDB)
	m := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/listUTXOs":
			listUTXOsHandler.HandlerFunc(ctx)
		default:
			ctx.Error("Not found", fasthttp.StatusNotFound)
		}
	}

	server := fasthttp.Server{
		Name:               "PlasmaUTXOlister",
		Concurrency:        httpConfig.HTTPConcurrency,
		MaxConnsPerIP:      httpConfig.MaxConnectionsPerIP,
		WriteTimeout:       time.Second * 15,
		ReadTimeout:        time.Second * 15,
		Handler:            m,
		MaxRequestBodySize: httpConfig.MaxBodySize,
	}

	var listener net.Listener
	go func() {
		listener, err = reuseport.Listen("tcp4", "0.0.0.0"+":"+strconv.Itoa(httpConfig.Port))
		if err != nil {
			panic("Can not bind")
		}
		if err = server.Serve(listener); err != nil {
			log.Println(err)
		}
	}()

	fmt.Println("Started to listen on " + "0.0.0.0" + ":" + strconv.Itoa(httpConfig.Port))
	wait := time.Second * 15
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	_, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	listener.Close()
	log.Println("Shutting down")
	os.Exit(0)
}
