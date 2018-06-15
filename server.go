package main

import (
	"context"
	"fmt"
	"log"
	"net"
	// _ "net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	handlers "github.com/bankex/go-plasma/handlers"
	env "github.com/caarlos0/env"
	redis "github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/reuseport"
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
	FundingTXSigningKey      string `env:"FUNDINGTX_ETH_KEY" envDefault:"0x34d8598d99b70cd57cb55ebfbbfd0c68847ce0faad5320b79665a281c39bc0d9"`
	BlockSigningKey          string `env:"BLOCK_ETH_KEY" envDefault:"0x34d8598d99b70cd57cb55ebfbbfd0c68847ce0faad5320b79665a281c39bc0d9"`
}

func main() {
	// runtime.SetCPUProfileRate(1000)
	// go http.ListenAndServe("0.0.0.0:8080", nil)
	// defer profile.Start().Stop()

	// defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
	fdb.MustAPIVersion(510)
	foundDB := fdb.MustOpenDefault()
	cfg := config{}
	err := env.Parse(&cfg)
	if err != nil {
		log.Printf("%+v\n", err)
	}
	fmt.Printf("%+v\n", cfg)
	// db, err := sql.Open("mysql", cfg.DatabaseUser+":"+
	// 	cfg.DatabasePassword+"@"+"tcp("+
	// 	cfg.DatabaseHost+":"+strconv.Itoa(cfg.DatabasePort)+")"+"/"+cfg.DatabaseName)
	// if err != nil {
	// 	panic("Failed to connect database")
	// }
	// db.SetMaxOpenConns(cfg.DatabaseConnectionsLimit)
	// defer db.Close()
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

	// maxCounterReader := sqlfunctions.NewMaxCounterReader(db)
	// dbCounter, err := maxCounterReader.GetMaxCounter()
	// if err != nil {
	// log.Println(err)
	// }
	fmt.Println("Redis counter = ", redisCounter)
	// fmt.Println("Database counter = ", dbCounter)
	// if redisCounter < dbCounter {
	// log.Fatal("Counters are out of order")
	// panic("Failed to connect database")
	// db.Close()
	// redisClient.Close()
	// os.Exit(1)
	// }

	// sendRawRLPTXhandler := handlers.NewSendRawRLPTXHandler(db, redisClient)
	sendRawTXHandler := handlers.NewSendRawTXHandler(&foundDB, redisClient)
	createUTXOHandler := handlers.NewCreateUTXOHandler(&foundDB)
	// listUTXOsHandler := handlers.NewListUTXOsHandler(&foundDB)
	// assembleBlockHandler := handlers.NewAssembleBlockHandler(&foundDB, redisClient, common.FromHex(cfg.BlockSigningKey))
	// createFundingTXhandler := handlers.NewCreateFundingTXHandler(&foundDB, redisClient, common.FromHex(cfg.FundingTXSigningKey))
	// lastBlockHandler := handlers.NewLastBlockHandler(&foundDB)

	// router := routing.New()

	// router.Post("/sendRawTX", func(c *routing.Context) error {
	// 	sendRawTXHandler.HandlerFunc(c)
	// 	return nil
	// })

	m := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/sendRawTX":
			sendRawTXHandler.HandlerFunc(ctx)
		case "/createUTXO":
			createUTXOHandler.HandlerFunc(ctx)
		default:
			ctx.Error("Not found", fasthttp.StatusNotFound)
		}
	}

	server := fasthttp.Server{
		Name:          "Plasma",
		Concurrency:   1000000,
		MaxConnsPerIP: 1000000,
		WriteTimeout:  time.Second * 15,
		ReadTimeout:   time.Second * 15,
		Handler:       m,
	}

	// r := mux.NewRouter()
	// // r.HandleFunc("/sendRawRLPTX", sendRawRLPTXhandler.Handle).Methods("POST")
	// r.HandleFunc("/sendRawTX", sendRawTXHandler.Handle).Methods("POST")
	// r.HandleFunc("/createUTXO", createUTXOHandler.Handle).Methods("POST")
	// r.HandleFunc("/listUTXOs", listUTXOsHandler.Handle).Methods("POST")
	// r.HandleFunc("/assembleBlock", assembleBlockHandler.Handle).Methods("POST")
	// r.HandleFunc("/createFundingTX", createFundingTXhandler.Handle).Methods("POST")
	// r.HandleFunc("/lastWrittenBlock", lastBlockHandler.Handle).Methods("GET")

	// srv := &http.Server{
	// 	Addr:         "0.0.0.0" + ":" + strconv.Itoa(cfg.Port),
	// 	WriteTimeout: time.Second * 15,
	// 	ReadTimeout:  time.Second * 15,
	// 	IdleTimeout:  time.Second * 60,
	// 	Handler:      r,
	// }

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

	var listener net.Listener
	go func() {
		// log.Println("Listening at " + srv.Addr)
		// if err := srv.ListenAndServe(); err != nil {
		// 	log.Println(err)
		// }
		listener, err = reuseport.Listen("tcp4", "0.0.0.0"+":"+strconv.Itoa(cfg.Port))
		if err != nil {
			panic("Can not bind")
		}
		if err = server.Serve(listener); err != nil {
			log.Println(err)
		}
		// if err := server.ListenAndServe("0.0.0.0" + ":" + strconv.Itoa(cfg.Port)); err != nil {
		// 	log.Println(err)
		// }
	}()
	wait := time.Second * 15
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	// ctx, cancel := context.WithTimeout(context.Background(), wait)
	_, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	listener.Close()
	// srv.Shutdown(ctx)
	log.Println("Shutting down")
	os.Exit(0)
}
