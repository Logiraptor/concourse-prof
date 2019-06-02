package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

func main() {
	http.HandleFunc("/api/", func(rw http.ResponseWriter, req *http.Request) {
		concourseUrl := req.Header.Get("X-Concourse-URL")
		url, err := url.Parse(concourseUrl)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
		proxy := httputil.NewSingleHostReverseProxy(url)
		proxy.FlushInterval = -1
		proxy.ServeHTTP(rw, req)
	})
	http.Handle("/", http.FileServer(http.Dir("./frontend/build")))
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("Listening on port", port)
	fmt.Println(http.ListenAndServe(":"+port, nil))
}
