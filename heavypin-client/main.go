package main

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var hostname *string
var password *string
var config = &tls.Config{
	InsecureSkipVerify: true,
}
var transport = &http.Transport{
	TLSClientConfig: config,
}
var client = &http.Client{
	Timeout:   time.Minute,
	Transport: transport,
}

func random() string {
	buff := make([]byte, 32)
	rand.Read(buff)
	return hex.EncodeToString(buff)
}

func padding() string {
	size, err := rand.Int(rand.Reader, big.NewInt(1<<10))
	if err != nil {
		panic(err)
	}
	buff := make([]byte, size.Int64()+1)
	rand.Read(buff)
	return hex.EncodeToString(buff)
}

func handle(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		panic(errors.New("hijacker not ok"))
	}
	conn, _, err := hijacker.Hijack()
	if err != nil {
		panic(err)
	}
	go transfer(conn, r.Host, random())
}

func transfer(conn net.Conn, host, token string) {
	form := url.Values{}
	form.Add("host", host)
	form.Add("token", token)
	form.Add("padding", padding())
	form.Add("password", *password)
	res, err := client.PostForm(*hostname+"/create", form)
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
			res, err := client.PostForm(*hostname+"/retrieve", form)
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
				client.PostForm(*hostname+"/done", form)
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

		req, err := http.NewRequest("POST", *hostname+"/proxy", body)
		if err != nil {
			panic(err)
		}
		req.Header.Add("token", token)
		req.Header.Add("padding", padding())
		req.Header.Add("password", *password)
		client.Do(req)
	}
}

func main() {
	flag.Usage = func() {
		fmt.Println("Usage: heavypin-client -s https://<server_ip> -p password")
	}
	hostname = flag.String("s", "", "")
	password = flag.String("p", "", "")
	flag.Parse()

	if *hostname == "" || *password == "" {
		flag.Usage()
		os.Exit(1)
	}

	if !strings.HasPrefix(*hostname, "https") {
		panic(errors.New("server URL must be https"))
	}

	fmt.Print("Connecting to the server...")
	req, err := http.NewRequest("GET", *hostname, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("padding", padding())
	req.Header.Add("password", *password)
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
		panic(errors.New("failed to use TLS"))
	} else {
		versions := map[uint16]string{
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
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		}),
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
