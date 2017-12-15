package main

import (
	"net/http"

	"github.com/csouls/counter_test"
	"github.com/gorilla/mux"
)

func init() {
	r := mux.NewRouter()

	r.HandleFunc("/memcached", counter_test.MemcachedHandler).Methods("POST")

	r.HandleFunc("/datastore", counter_test.DatastoreInfoHandler).Methods("GET")
	r.HandleFunc("/datastore", counter_test.DatastoreHandler).Methods("POST")

	r.HandleFunc("/sharding_datastore", counter_test.ShardingDatastoreInfoHandler).Methods("GET")
	r.HandleFunc("/sharding_datastore", counter_test.ShardingDatastoreHandler).Methods("POST")

	http.Handle("/", r)
}
