package main

import (
	"log"
	"net/http"
	"path/filepath"
)

// httpParams stores the http connection parameters
type httpParams struct {
	address string
	prefix  string
	root    string
}

func main() {
	httpConn := &httpParams{
		address: "localhost:5000",
		prefix:  "/",
		root:    ".",
	}
	initServer(httpConn)
}

// initServer initializes the webserver
func initServer(p *httpParams) {
	var err error
	p.root, err = filepath.Abs(p.root)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("serving %s as %s on %s", p.root, p.prefix, p.address)
	http.Handle(p.prefix, http.StripPrefix(p.prefix, http.FileServer(http.Dir(p.root))))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Print(r.RemoteAddr + " " + r.Method + " " + r.URL.String())
		http.DefaultServeMux.ServeHTTP(w, r)
	})

	httpServer := http.Server{
		Addr:    p.address,
		Handler: handler,
	}
	err = httpServer.ListenAndServe()
	if err != nil {
		log.Fatalln(err)
	}
}
