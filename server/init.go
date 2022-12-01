package main

import (
	"log"
	"net/http"
	"path/filepath"
)

// httpConn web server connection parameters
type httpConn struct {
	port       string
	root       string
	cascadeDir string
}

func main() {
	httpConn := &httpConn{
		port:       "6060",
		root:       "./",
		cascadeDir: "./cascade/",
	}
	initServer(httpConn)
}

// initServer initializes the webserver
func initServer(c *httpConn) {
	var err error
	c.root, err = filepath.Abs(c.root)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("serving %s on localhost:%s", c.root, c.port)
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir(c.root))))
	http.Handle("/cascade/", http.StripPrefix("/cascade/", http.FileServer(http.Dir(c.cascadeDir))))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Print(r.RemoteAddr + " " + r.Method + " " + r.URL.String())
		http.DefaultServeMux.ServeHTTP(w, r)
	})
	log.Fatalln(http.ListenAndServe(":"+c.port, handler))
}
