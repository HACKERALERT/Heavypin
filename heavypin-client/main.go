package main

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
)

var hostname *string

func random() string {
	buff := make([]byte, 32)
	rand.Read(buff)
	return hex.EncodeToString(buff)
}

func handle(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	hijacker, _ := w.(http.Hijacker)
	conn, _, _ := hijacker.Hijack()
	go transfer(conn, r.Host, random())
}

func transfer(conn net.Conn, host, token string) {
	form := url.Values{}
	form.Add("host", host)
	form.Add("token", token)
	http.PostForm(*hostname+"/create", form)

	go func() {
		for {
			form := url.Values{}
			form.Add("token", token)
			var err error
			var res *http.Response
			var body []byte
			if res, err = http.PostForm(*hostname+"/retrieve", form); err != nil {
				conn.Close()
				break
			}
			if body, err = io.ReadAll(res.Body); err != nil {
				conn.Close()
				break
			}
			res.Body.Close()

			if len(body) == 1<<15+1 {
				conn.Close()
				fmt.Println("Closed", token)
				form := url.Values{}
				form.Add("token", token)
				http.PostForm(*hostname+"/done", form)
				break
			}

			fmt.Println(token, "received", len(body), "bytes from server")
			conn.Write(body)
			time.Sleep(100 * time.Millisecond)
		}
	}()

	for {
		data := make([]byte, 1<<15)
		size, err := conn.Read(data)
		if err != nil {
			break
		}
		data = data[:size]

		form := url.Values{}
		form.Add("token", token)
		form.Add("content", hex.EncodeToString(data))
		http.PostForm(*hostname+"/proxy", form)
		time.Sleep(100 * time.Millisecond)
	}
}

func main() {
	flag.Usage = func() { fmt.Println("Usage: heavypin-client -s \"http(s)://<server_hostname_or_ip>:<server_port>\"") }
	hostname = flag.String("s", "", "")
	flag.Parse()

	if *hostname == "" {
		flag.Usage()
		os.Exit(1)
	}

	fmt.Println("Connecting to", *hostname)
	res, err := http.Get(*hostname)
	if err != nil || res.StatusCode != http.StatusNoContent {
		fmt.Println("Couldn't connect to server!")
		os.Exit(1)
	}
	fmt.Println("Connected to server, starting HTTP proxy on :8888")

	server := &http.Server{
		Addr: ":8888",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("Connecting to", r.URL.String())
			if r.Method == http.MethodConnect {
				handle(w, r)
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		}),
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}
	server.ListenAndServe()
}
