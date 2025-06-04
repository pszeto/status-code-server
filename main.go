package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var startTime time.Time

func uptime() time.Duration {
	return time.Since(startTime)
}

func init() {
	startTime = time.Now()
}

func status(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Server", "status-code-server")
	resp := make(map[string]string)
	resp["uptime"] = uptime().String()
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}
	w.Write(jsonResp)
	return
}

func root(w http.ResponseWriter, req *http.Request) {
	statusCode := 200
	statusCodeFromEnv, ok := os.LookupEnv("STATUS_CODE")
	if !ok {
		log.Println("STATUS_CODE not defined.  Defaulting to 200")
	} else {
		// convert to int
		num, err := strconv.Atoi(statusCodeFromEnv)
		if err != nil {
			log.Println("STATUS_CODE not a number.  Defaulting to 200")
			log.Println("STATUS_CODE from env is: ", num)
		}
		statusCode = num
	}
	w.Header().Set("X-Server", "status-code-server")
	w.WriteHeader(statusCode)
	w.Write([]byte("DONE"))
}


func Run(addr string, sslAddr string, ssl map[string]string) chan error {

	errs := make(chan error)

	// Starting HTTP server
	go func() {
		log.Printf("Staring HTTP service on %s", addr)

		if err := http.ListenAndServe(addr, nil); err != nil {
			errs <- err
		}

	}()

	// Starting HTTPS server
	go func() {
		log.Printf("Staring HTTPS service on %s", sslAddr)
		if err := http.ListenAndServeTLS(sslAddr, ssl["cert"], ssl["key"], nil); err != nil {
			errs <- err
		}
	}()

	return errs
}

func main() {
	httpPort, ok := os.LookupEnv("HTTP_PORT")
	if !ok {
		log.Println("HTTP_PORT not defined.  Defaulting to 8080")
		httpPort = ":8080"
	} else {
		httpPort = ":" + httpPort
	}

	httpsPort, ok := os.LookupEnv("HTTPS_PORT")
	if !ok {
		log.Println("HTTPS_PORT not defined.  Defaulting to 8443")
		httpsPort = ":8443"
	} else {
		httpsPort = ":" + httpsPort
	}


	http.HandleFunc("/status", status)
	http.HandleFunc("/", root)

	log.Println("Version 0.1")

	errs := Run(httpPort, httpsPort, map[string]string{
		"cert": "server.crt",
		"key":  "server.key",
	})

	// This will run forever until channel receives error
	select {
	case err := <-errs:
		log.Printf("Could not start serving service due to (error: %s)", err)
	}
}
