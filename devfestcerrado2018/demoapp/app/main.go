package main

import (
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/mediocregopher/radix.v2/redis"
	"github.com/urfave/negroni"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

var (
	redisConn *redis.Client
	log       *logrus.Logger

	// Prometheus metrics
	openConnections = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "myapp_open_connections",
		Help: "The current number of open connections.",
	})

	handlerDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "myapp_handlers_duration_seconds",
		Help: "Handlers request duration in seconds",
	}, []string{"path"})

	redisRequestDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "myapp_redis_request_duration_seconds",
		Help: "The duration of the requests to the redis service",
	})

	redisCached = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "myapp_redis_cached",
		Help: "The current size of redis cache.",
	})

	redisReturnedCache = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "myapp_redis_returned_cache",
		Help: "The current number of access to cache.",
	})
)

func init() {
	prometheus.MustRegister(openConnections)
	prometheus.MustRegister(handlerDuration)
	prometheus.MustRegister(redisRequestDuration)
	prometheus.MustRegister(redisCached)
	prometheus.MustRegister(redisReturnedCache)
}

func main() {
	log = logrus.StandardLogger()
	log.Infoln("Starting API...")

	// getting redis envs
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")

	// connecting to redis
	conn, err := redis.Dial("tcp", redisHost+":"+redisPort)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	redisConn = conn

	// getting port env or setting default
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// creating and starting server
	s := &http.Server{
		Addr:         ":" + port,
		Handler:      router(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		ConnState: func(c net.Conn, s http.ConnState) {
			switch s {
			case http.StateNew:
				openConnections.Inc()
			case http.StateHijacked | http.StateClosed:
				openConnections.Dec()
			}
		},
	}
	log.Infof("HTTP server listening at :%s...", port)
	log.Fatal(s.ListenAndServe())
}

func router() http.Handler {
	r := mux.NewRouter().StrictSlash(true)
	r.HandleFunc("/", instrumentHandler("/", index)).Methods("GET")
	r.HandleFunc("/resize", instrumentHandler("/resize", resize)).Methods("POST")
	r.HandleFunc("/view/{hash}", instrumentHandler("/view/{hash}", view)).Methods("GET")
	r.HandleFunc("/healthcheck", instrumentHandler("/healthchceck", healthcheck)).Methods("GET")

	r.Handle("/metrics", promhttp.Handler())

	fs := http.StripPrefix("/static/", http.FileServer(http.Dir("/tmp")))
	r.PathPrefix("/static/").Handler(fs)

	n := negroni.Classic()
	n.UseHandler(r)
	return n
}
