package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

func instrumentHandler(pattern string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		next(w, r)
		handlerDuration.WithLabelValues(pattern).Observe(time.Since(now).Seconds())
	})
}

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>Thumbor Resizer</h1>"+
		"<form action=\"/resize\" method=\"POST\">"+
		"<p>Image URL</p><input name=\"url\" type=\"url\" placeholder=\"Insert Image URL\"><br>"+
		"<p>Width:</p><input name=\"width\" type=\"number\"><br>"+
		"<p>Height:</p><input name=\"height\" type=\"number\"><br>"+
		"<input name=\"smart\" type=\"checkbox\" value=\"smart\">Smart filter<br>"+
		"<input type=\"submit\" value=\"Resize!\">"+
		"</form>")

}

func resize(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue("url")
	widthStr := r.FormValue("width")
	width, _ := strconv.ParseInt(widthStr, 10, 64)
	heightStr := r.FormValue("height")
	height, _ := strconv.ParseInt(heightStr, 10, 64)
	smartStr := r.FormValue("smart")
	smart := smartStr == "smart"

	hasher := md5.New()
	hasher.Write([]byte(url + widthStr + heightStr + smartStr))
	hash := hex.EncodeToString(hasher.Sum(nil))

	now := time.Now()
	resizeUrl, _ := redisConn.Cmd("HGET", "hash:"+hash, "url").Str()
	redisRequestDuration.Observe(time.Since(now).Seconds())

	if resizeUrl != "" {
		redisReturnedCache.Inc()
		http.Redirect(w, r, "/view/"+hash, http.StatusFound)
		return
	}

	// thumbor
	client := GetClient()
	resp, err := client.do(width, height, smart, url)
	if err != nil {
		log.WithField("err", err).Error("error getting image")
		fmt.Fprint(w, "sorry error")
		return
	}
	defer resp.Body.Close()

	// create file
	file, err := os.Create("/tmp/" + hash + ".jpg")
	if err != nil {
		log.WithField("err", err).Error("error creating file")
		fmt.Fprint(w, "sorry error")
		return
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		log.WithField("err", err).Error("error copying body to file")
		fmt.Fprint(w, "sorry error")
		return
	}
	file.Close()

	// add to cache
	err = redisConn.Cmd("HMSET", "hash:"+hash, "url", url).Err
	if err != nil {
		log.WithField("err", err).Error("error saving on redis")
		fmt.Fprint(w, "sorry error")
		return
	}
	redisCached.Inc()

	// redirect to view
	http.Redirect(w, r, "/view/"+hash, http.StatusFound)
}

func view(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]

	// get file name in cache
	url, _ := redisConn.Cmd("HGET", "hash:"+hash, "url").Str()
	if url == "" {
		fmt.Fprintf(w, "<h1>url for hash %s not found</h1>", hash)
		return
	}

	// return the file
	fmt.Fprintf(w, "<p>hash: %s -> url: %s</p>"+
		"<img src=\"/static/%s.jpg\"></img>", hash, url, hash)
}

func healthcheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
