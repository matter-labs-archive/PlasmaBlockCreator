package main

import (
	"context"
	sql "database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	handlers "github.com/bankex/go-plasma/handlers"
	"github.com/bankex/go-plasma/sqlfunctions"
	env "github.com/caarlos0/env"
	redis "github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	mux "github.com/gorilla/mux"
)

type config struct {
	Port                     int    `env:"PORT" envDefault:"3001"`
	DatabaseHost             string `env:"DB_HOST" envDefault:"127.0.0.1"`
	DatabasePort             int    `env:"DB_PORT" envDefault:"3306"`
	DatabaseName             string `env:"DB_SCHEMA" envDefault:"plasma"`
	DatabaseUser             string `env:"DB_LOGIN" envDefault:"root"`
	DatabasePassword         string `env:"DB_PASSWORD" envDefault:"example"`
	DatabaseConnectionsLimit int    `env:"DB_CONNECTIONS" envDefault:"16"`
	RedisHost                string `env:"REDIS_HOST" envDefault:"127.0.0.1"`
	RedisPort                int    `env:"REDIS_PORT" envDefault:"6379"`
	RedisPassword            string `env:"REDIS_PASSWORD" envDefault:""`
}

func main() {
	fdb.MustAPIVersion(510)
	foundDB := fdb.MustOpenDefault()
	cfg := config{}
	err := env.Parse(&cfg)
	if err != nil {
		log.Printf("%+v\n", err)
	}
	fmt.Printf("%+v\n", cfg)
	db, err := sql.Open("mysql", cfg.DatabaseUser+":"+
		cfg.DatabasePassword+"@"+"tcp("+
		cfg.DatabaseHost+":"+strconv.Itoa(cfg.DatabasePort)+")"+"/"+cfg.DatabaseName)
	if err != nil {
		panic("Failed to connect database")
	}
	db.SetMaxOpenConns(cfg.DatabaseConnectionsLimit)
	defer db.Close()
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisHost + ":" + strconv.Itoa(cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       0,
	})
	defer redisClient.Close()

	redisCounter, err := redisClient.Get("ctr").Uint64()
	if err != nil {
		log.Println(err)
	}

	maxCounterReader := sqlfunctions.NewMaxCounterReader(db)
	dbCounter, err := maxCounterReader.GetMaxCounter()
	if err != nil {
		log.Println(err)
	}
	fmt.Println("Redis counter = ", redisCounter)
	fmt.Println("Database counter = ", dbCounter)
	if redisCounter < dbCounter {
		log.Fatal("Counters are out of order")
		panic("Failed to connect database")
		db.Close()
		redisClient.Close()
		os.Exit(1)
	}

	sendRawRLPTXhandler := handlers.NewSendRawRLPTXHandler(db, redisClient)
	sendRawTXHandler := handlers.NewSendRawTXHandler(&foundDB, redisClient)
	createUTXOHandler := handlers.NewCreateUTXOHandler(&foundDB)
	listUTXOsHandler := handlers.NewListUTXOsHandler(&foundDB)
	r := mux.NewRouter()
	r.HandleFunc("/sendRawRLPTX", sendRawRLPTXhandler.Handle).Methods("POST")
	r.HandleFunc("/sendRawTX", sendRawTXHandler.Handle).Methods("POST")
	r.HandleFunc("/createUTXO", createUTXOHandler.Handle).Methods("POST")
	r.HandleFunc("/listUTXOs", listUTXOsHandler.Handle).Methods("POST")

	srv := &http.Server{
		Addr:         "0.0.0.0" + ":" + strconv.Itoa(cfg.Port),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}

	// err = r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
	// 	pathTemplate, err := route.GetPathTemplate()
	// 	if err == nil {
	// 		fmt.Println("ROUTE:", pathTemplate)
	// 	}
	// 	pathRegexp, err := route.GetPathRegexp()
	// 	if err == nil {
	// 		fmt.Println("Path regexp:", pathRegexp)
	// 	}
	// 	queriesTemplates, err := route.GetQueriesTemplates()
	// 	if err == nil {
	// 		fmt.Println("Queries templates:", strings.Join(queriesTemplates, ","))
	// 	}
	// 	queriesRegexps, err := route.GetQueriesRegexp()
	// 	if err == nil {
	// 		fmt.Println("Queries regexps:", strings.Join(queriesRegexps, ","))
	// 	}
	// 	methods, err := route.GetMethods()
	// 	if err == nil {
	// 		fmt.Println("Methods:", strings.Join(methods, ","))
	// 	}
	// 	fmt.Println()
	// 	return nil
	// })

	// if err != nil {
	// 	fmt.Println(err)
	// }

	go func() {
		log.Println("Listening at " + srv.Addr)
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()
	wait := time.Second * 15
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	srv.Shutdown(ctx)
	log.Println("Shutting down")
	os.Exit(0)
}
