package main

import (
	"flag"
	"log"
	"net/http"
	
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	glogging "github.com/op/go-logging"
	
	rserver "github.com/bencase/revis-service/server"
)

const pathPrefix = "/api"
const redisPathPrefix = "/redis"

const portFlag = "port"

var logger = glogging.MustGetLogger("main")

var port = "63799"

func init() {
	flag.StringVar(&port, portFlag, "63799", "the port on which to start the server")
	flag.Parse()
}

func main() {
	
	server, err := rserver.NewRedisServer()
	if err != nil {
		log.Fatalln("Error getting server instance:", err)
	}
	defer server.Close()
	
	r := mux.NewRouter()
	r.HandleFunc(pathPrefix + redisPathPrefix + "/hello", server.SayHello)
	
	r.HandleFunc(pathPrefix + redisPathPrefix + "/connections",
			server.GetConnections).
		Methods("GET")
	r.HandleFunc(pathPrefix + redisPathPrefix + "/connections",
			server.UpsertConnections).
		Methods("POST")
	r.HandleFunc(pathPrefix + redisPathPrefix + "/connections",
			server.DeleteConnections).
		Methods("DELETE")

	r.HandleFunc(pathPrefix + redisPathPrefix + "/connections/test",
			server.TestConnection).
		Methods("POST")
	
	r.HandleFunc(pathPrefix + redisPathPrefix + "/kvs",
			server.GetKeysWithValues).
		Methods("GET")
	
	corsOpts := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"HEAD", "GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{rserver.ConnNameHeader,
			rserver.PatternHeader,
			rserver.ScanIdHeader},
	})
	handler := corsOpts.Handler(r)
	http.Handle("/", handler)
	
	logger.Info("Serving on", ":" + port)
	if err := http.ListenAndServe(":" + port, nil); err != nil {
		panic(err)
	}
}
