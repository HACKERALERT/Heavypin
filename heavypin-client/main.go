package main

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var hostname *string
var password *string

func random() string {
	buff := make([]byte, 32)
	rand.Read(buff)
	return hex.EncodeToString(buff)
}

func padding() string {
	size, _ := rand.Int(rand.Reader, big.NewInt(1<<10))
	buff := make([]byte, size.Int64()+1)
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
	form.Add("padding", padding())
	form.Add("password", *password)
	res, err := http.PostForm(*hostname+"/create", form)
	if err != nil || res.StatusCode != http.StatusNoContent {
		conn.Close()
		return
	}

	go func() {
		for {
			form := url.Values{}
			form.Add("token", token)
			form.Add("padding", padding())
			form.Add("password", *password)
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

			if bytes.Equal(body, make([]byte, 1<<20+1)) {
				conn.Close()
				form := url.Values{}
				form.Add("token", token)
				form.Add("padding", padding())
				form.Add("password", *password)
				http.PostForm(*hostname+"/done", form)
				break
			}

			conn.Write(body)
		}
	}()

	for {
		data := make([]byte, 1<<20)
		size, err := conn.Read(data)
		if err != nil {
			break
		}
		if size == 0 {
			continue
		}
		data = data[:size]
		body := bytes.NewReader(data)

		req, _ := http.NewRequest("POST", *hostname+"/proxy", body)
		req.Header.Add("token", token)
		req.Header.Add("padding", padding())
		req.Header.Add("password", *password)
		client := &http.Client{}
		client.Do(req)
	}
}

func main() {
	flag.Usage = func() {
		fmt.Println("Usage: heavypin-client -s http(s)://<server_hostname_or_ip> -p password")
	}
	hostname = flag.String("s", "", "")
	password = flag.String("p", "", "")
	flag.Parse()

	if *hostname == "" || *password == "" {
		flag.Usage()
		os.Exit(1)
	}

	fmt.Print("Connecting to the server...")
	req, _ := http.NewRequest("GET", *hostname, nil)
	req.Header.Add("padding", padding())
	req.Header.Add("password", *password)
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(" Unable to establish a connection!")
		os.Exit(1)
	}
	fmt.Print(" Success.\nLogging in with password...")
	if res.StatusCode != http.StatusNoContent {
		fmt.Println(" Incorrect password!")
		os.Exit(1)
	}
	fmt.Print(" Success.\nCurrent connection suite... ", res.Proto)
	if res.TLS == nil {
		fmt.Println(" (unencrypted)")
	} else {
		versions := map[uint16]string{
			tls.VersionSSL30: " (SSL 3.0)",
			tls.VersionTLS10: " (TLS 1.0)",
			tls.VersionTLS11: " (TLS 1.1)",
			tls.VersionTLS12: " (TLS 1.2)",
			tls.VersionTLS13: " (TLS 1.3)",
		}
		fmt.Println(versions[res.TLS.Version])
	}
	fmt.Println("Use http://localhost:8000 as an HTTP proxy to access the Internet.")

	server := &http.Server{
		Addr: ":8000",
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
