package main

import (
	"bytes"
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
	"strings"
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
	res, err := http.PostForm(*hostname+"/create", form)
	if err != nil || res.StatusCode != http.StatusNoContent {
		conn.Close()
		return
	}

	go func() {
		for {
			form := url.Values{}
			form.Add("token", token)
			res, err := http.PostForm(*hostname+"/retrieve", form)
			if err != nil {
				conn.Close()
				break
			}
			body, err := io.ReadAll(res.Body)
			if err != nil {
				conn.Close()
				break
			}
			res.Body.Close()

			if bytes.Equal(body, make([]byte, 1<<16+1)) {
				conn.Close()
				form := url.Values{}
				form.Add("token", token)
				http.PostForm(*hostname+"/done", form)
				break
			}

			conn.Write(body)
			time.Sleep(50 * time.Millisecond)
		}
	}()

	for {
		data := make([]byte, 1<<16)
		size, err := conn.Read(data)
		if err != nil {
			break
		}
		data = data[:size]

		form := url.Values{}
		form.Add("token", token)
		form.Add("content", hex.EncodeToString(data))
		http.PostForm(*hostname+"/proxy", form)
		time.Sleep(50 * time.Millisecond)
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

	fmt.Println("Attempting to connect to the server...")
	res, err := http.Get(*hostname)
	if err != nil || res.StatusCode != http.StatusNoContent {
		fmt.Println("Unable to establish a connection! Check your server address and try again.")
		os.Exit(1)
	}
	fmt.Println("Connected to server. Use http://localhost:8888 as an HTTP proxy to access the Internet.")

	server := &http.Server{
		Addr: ":8888",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodConnect {
				handle(w, r)
			} else {
				http.Redirect(w, r, strings.ReplaceAll(r.URL.String(), "http://", "https://"), http.StatusPermanentRedirect)
			}
		}),
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}
	server.ListenAndServe()
}
