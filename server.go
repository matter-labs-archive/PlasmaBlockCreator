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

	"github.com/bankex/go-plasma/foundationdb"
	"github.com/ethereum/go-ethereum/common"

	"github.com/bankex/go-plasma/transaction"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	handlers "github.com/bankex/go-plasma/handlers"
	env "github.com/caarlos0/env"
	redis "github.com/go-redis/redis"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/reuseport"
)

type config struct {
	Port                 int    `env:"PORT" envDefault:"3001"`
	RedisHost            string `env:"REDIS_HOST" envDefault:"127.0.0.1"`
	RedisPort            int    `env:"REDIS_PORT" envDefault:"6379"`
	RedisPassword        string `env:"REDIS_PASSWORD" envDefault:""`
	FundingTXSigningKey  string `env:"FUNDINGTX_ETH_KEY" envDefault:"0x34d8598d99b70cd57cb55ebfbbfd0c68847ce0faad5320b79665a281c39bc0d9"`
	BlockSigningKey      string `env:"BLOCK_ETH_KEY" envDefault:"0x34d8598d99b70cd57cb55ebfbbfd0c68847ce0faad5320b79665a281c39bc0d9"`
	DatabaseConcurrency  int    `env:"FDB_CONCURRENCY" envDefault:"-1"`
	ECRecoverConcurrency int    `env:"EC_CONCURRENCY" envDefault:"-1"`
	MaxProc              int    `env:"GOMAXPROCS" envDefault:"-1"`
}

const defaultDatabaseConcurrency = 100000
const defaultECRecoverConcurrency = 30000

func main() {
	fdb.MustAPIVersion(520)
	foundDB := fdb.MustOpenDefault()
	cfg := config{}
	err := env.Parse(&cfg)
	if err != nil {
		log.Printf("%+v\n", err)
	}
	fmt.Printf("%+v\n", cfg)
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisHost + ":" + strconv.Itoa(cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       0,
	})
	defer redisClient.Close()

	redisCounter, err := redisClient.Get("ctr").Uint64()
	if err != nil {
		log.Println(err)
		redisClient.Close()
		os.Exit(1)
	}

	maxCounterInDatabase, err := foundationdb.GetMaxTransactionCounter(&foundDB)
	if err != nil {
		log.Println(err)
		redisClient.Close()
		os.Exit(1)
	}
	fmt.Println("Redis counter = ", strconv.FormatUint(redisCounter, 10))
	fmt.Println("Database counter = ", strconv.FormatUint(maxCounterInDatabase, 10))

	if maxCounterInDatabase > redisCounter {
		log.Println("Counters mismatch")
		redisClient.Close()
		os.Exit(1)
	}

	ECRecoverConcurrency := cfg.ECRecoverConcurrency
	if ECRecoverConcurrency == -1 {
		ECRecoverConcurrency = cfg.MaxProc
		if ECRecoverConcurrency == -1 {
			ECRecoverConcurrency = runtime.NumCPU()
		}

	}
	DatabaseConcurrency := cfg.DatabaseConcurrency
	if DatabaseConcurrency == -1 {
		DatabaseConcurrency = defaultDatabaseConcurrency
	}

	fmt.Println("ECRecover concurrency = " + strconv.Itoa(ECRecoverConcurrency))
	fmt.Println("FDB concurrency = " + strconv.Itoa(DatabaseConcurrency))

	transactionParser := transaction.NewTransactionParser(ECRecoverConcurrency)
	sendRawTXHandler := handlers.NewSendRawTXHandler(&foundDB, redisClient, transactionParser, DatabaseConcurrency)
	createUTXOHandler := handlers.NewCreateUTXOHandler(&foundDB)
	listUTXOsHandler := handlers.NewListUTXOsHandler(&foundDB)
	assembleBlockHandler := handlers.NewAssembleBlockHandler(&foundDB, redisClient, common.FromHex(cfg.BlockSigningKey))
	createFundingTXhandler := handlers.NewCreateFundingTXHandler(&foundDB, redisClient, common.FromHex(cfg.FundingTXSigningKey))
	writeBlockHandler := handlers.NewWriteBlockHandler(&foundDB)
	lastBlockHandler := handlers.NewLastBlockHandler(&foundDB)

	m := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/sendRawTX":
			sendRawTXHandler.HandlerFunc(ctx)
		case "/createUTXO":
			createUTXOHandler.HandlerFunc(ctx)
		case "/listUTXOs":
			listUTXOsHandler.HandlerFunc(ctx)
		case "/assembleBlock":
			assembleBlockHandler.HandlerFunc(ctx)
		case "/createFundingTX":
			createFundingTXhandler.HandlerFunc(ctx)
		case "/lastWrittenBlock":
			lastBlockHandler.HandlerFunc(ctx)
		case "/writeBlock":
			writeBlockHandler.HandlerFunc(ctx)
		default:
			ctx.Error("Not found", fasthttp.StatusNotFound)
		}
	}

	server := fasthttp.Server{
		Name:               "Plasma",
		Concurrency:        100000,
		MaxConnsPerIP:      100000,
		WriteTimeout:       time.Second * 15,
		ReadTimeout:        time.Second * 15,
		Handler:            m,
		MaxRequestBodySize: 500000000,
	}

	var listener net.Listener
	go func() {
		listener, err = reuseport.Listen("tcp4", "0.0.0.0"+":"+strconv.Itoa(cfg.Port))
		if err != nil {
			panic("Can not bind")
		}
		if err = server.Serve(listener); err != nil {
			log.Println(err)
		}
	}()
	fmt.Println("Started to listen on " + "0.0.0.0" + ":" + strconv.Itoa(cfg.Port))
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
