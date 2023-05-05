package main

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

var hostname = "http://magnificentunderstatedfossil.esu5.repl.co"//"http://localhost:9999"

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
	http.PostForm(hostname+"/create", form)

	go func() {
		for {
			form := url.Values{}
			form.Add("token", token)
			var err error
			var res *http.Response
			var body []byte
			if res, err = http.PostForm(hostname+"/retrieve", form); err != nil {
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
				http.PostForm(hostname+"/done", form)
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
		http.PostForm(hostname+"/proxy", form)
		time.Sleep(100 * time.Millisecond)
	}
}

func main() {
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
