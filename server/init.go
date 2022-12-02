package main

import (
	"log"
	"net/http"
	"path/filepath"
)

var defaultConn = &httpConn{
	host:        "",
	port:        "6060",
	path:        "./",
	cascadePath: "./cascade/",
}

// httpConn web server connection parameters
type httpConn struct {
	host        string
	port        string
	path        string
	cascadePath string
}

func (c *httpConn) addr() string {
	return c.host + ":" + c.port
}

func init() {
	var err error
	defaultConn.path, err = filepath.Abs(defaultConn.path)
	if err != nil {
		log.Fatalln(err)
	}
	defaultConn.cascadePath, err = filepath.Abs(defaultConn.cascadePath)
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	log.Printf("serving %s on %s", defaultConn.path, defaultConn.addr())

	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir(defaultConn.path))))
	http.Handle("/cascade/", http.StripPrefix("/cascade/", http.FileServer(http.Dir(defaultConn.cascadePath))))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Print(r.RemoteAddr + " " + r.Method + " " + r.URL.String())
		http.DefaultServeMux.ServeHTTP(w, r)
	})

	log.Fatalln(http.ListenAndServe(defaultConn.addr(), handler))
}
