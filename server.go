package main

import (
	sql "database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"

	handlers "github.com/bankex/go-plasma/handlers"
	redis "github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	mux "github.com/gorilla/mux"
)

func main() {
	db, err := sql.Open("mysql", "root:example@/plasma")
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	defer redisClient.Close()
	sendRawRLPTXhandler := handlers.NewSendRawRLPTXHandler(db, redisClient)
	r := mux.NewRouter()
	r.HandleFunc("/sendRawRLPTX", sendRawRLPTXhandler.Handle).Methods("POST")
	err = r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err == nil {
			fmt.Println("ROUTE:", pathTemplate)
		}
		pathRegexp, err := route.GetPathRegexp()
		if err == nil {
			fmt.Println("Path regexp:", pathRegexp)
		}
		queriesTemplates, err := route.GetQueriesTemplates()
		if err == nil {
			fmt.Println("Queries templates:", strings.Join(queriesTemplates, ","))
		}
		queriesRegexps, err := route.GetQueriesRegexp()
		if err == nil {
			fmt.Println("Queries regexps:", strings.Join(queriesRegexps, ","))
		}
		methods, err := route.GetMethods()
		if err == nil {
			fmt.Println("Methods:", strings.Join(methods, ","))
		}
		fmt.Println()
		return nil
	})

	if err != nil {
		fmt.Println(err)
	}
	log.Fatal(http.ListenAndServe("0.0.0.0:3001", r))
}
