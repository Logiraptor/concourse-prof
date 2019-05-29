package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
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
	http.Handle("/", http.FileServer(http.Dir("../concourse-prof/build")))
	http.ListenAndServe(":8080", nil)
}
