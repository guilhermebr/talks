package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/globocom/gothumbor"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	client *thumborClient
)

var (
	thumborRequestsCurrent = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "myapp_thumbor_requests_current",
		Help: "The current number of requests to the weather service.",
	})

	thumborRequestsDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "myapp_thumbor_request_duration_seconds",
		Help: "The duration of the requests to the thumbor service",
	})

	thumborRequestsStatus = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "myapp_thumbor_requests_total",
		Help: "The total number of requests to the thumbor service by status.",
	}, []string{"status"})

	thumborClientErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "myapp_thumbor_errors",
		Help: "The total number of thumbor client errors",
	})
)

func init() {
	prometheus.MustRegister(thumborRequestsDuration)
	prometheus.MustRegister(thumborRequestsCurrent)
	prometheus.MustRegister(thumborRequestsStatus)
	prometheus.MustRegister(thumborClientErrors)
}

type thumborClient struct {
	httpClient *http.Client
	secret     string
	url        string
}

func GetClient() *thumborClient {
	secret := os.Getenv("THUMBOR_SECRET")
	host := os.Getenv("THUMBOR_HOST")
	port := os.Getenv("THUMBOR_PORT")
	url := fmt.Sprintf("http://%s:%s", host, port)

	if client == nil {
		client = &thumborClient{
			httpClient: &http.Client{Timeout: time.Second * 15},
			secret:     secret,
			url:        url,
		}
	}
	return client
}

func (c *thumborClient) do(width, height int64, smart bool, imageURL string) (resp *http.Response, err error) {
	now := time.Now()
	thumborRequestsCurrent.Inc()

	defer func() {
		thumborRequestsDuration.Observe(time.Since(now).Seconds())
		thumborRequestsCurrent.Dec()
		if resp != nil {
			thumborRequestsStatus.WithLabelValues(strconv.Itoa(resp.StatusCode)).Inc()
		}
		if err != nil {
			thumborClientErrors.Inc()
		}
	}()

	options := gothumbor.ThumborOptions{Width: int(width), Height: int(height), Smart: smart}
	encryptedURL, _ := gothumbor.GetCryptedThumborPath(c.secret, imageURL, options)
	url := fmt.Sprintf("%s/%s", c.url, encryptedURL)
	fmt.Printf("DEBUG: %s\n", url)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return c.httpClient.Do(request)
}
