package main

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

var password *string

type Tunnel struct {
	conn net.Conn
	data []byte
}

func padding() string {
	size, _ := rand.Int(rand.Reader, big.NewInt(1<<10))
	buff := make([]byte, size.Int64()+1)
	rand.Read(buff)
	return hex.EncodeToString(buff)
}

func main() {
	flag.Usage = func() { fmt.Println("Usage: heavypin-server -p password") }
	password = flag.String("p", "", "")
	flag.Parse()

	if *password == "" {
		flag.Usage()
		os.Exit(1)
	}

	tunnels := struct {
		sync.Mutex
		m map[string]*Tunnel
	}{m: make(map[string]*Tunnel)}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("password") == *password {
			w.Header().Add("padding", padding())
			w.WriteHeader(http.StatusNoContent)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	})

	http.HandleFunc("/create", func(w http.ResponseWriter, r *http.Request) {
		if r.FormValue("password") != *password {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		conn, err := net.DialTimeout("tcp", r.FormValue("host"), 5*time.Second)
		if err != nil {
			w.Header().Add("padding", padding())
			w.WriteHeader(http.StatusGatewayTimeout)
			return
		}
		tunnels.Lock()
		tunnels.m[r.FormValue("token")] = &Tunnel{
			conn: conn,
			data: nil,
		}
		tunnels.Unlock()
		go func() {
			conn.SetDeadline(time.Now().Add(5 * time.Second))
			for {
				data := make([]byte, 1<<20)
				size, err := conn.Read(data)
				if err != nil {
					for {
						tunnels.Lock()
						data := tunnels.m[r.FormValue("token")].data
						tunnels.Unlock()
						if len(data) == 0 {
							break
						}
						time.Sleep(5 * time.Millisecond)
					}
					tunnels.Lock()
					tunnels.m[r.FormValue("token")].data = make([]byte, 1<<20+1)
					tunnels.Unlock()
					break
				}
				if size != 0 {
					conn.SetDeadline(time.Now().Add(5 * time.Second))
					data = data[:size]
					tunnels.Lock()
					tunnels.m[r.FormValue("token")].data = append(tunnels.m[r.FormValue("token")].data, data...)
					tunnels.Unlock()
				}
			}
			conn.Close()
		}()
		w.Header().Add("padding", padding())
		w.WriteHeader(http.StatusNoContent)
	})

	http.HandleFunc("/proxy", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("password") != *password {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		content, _ := io.ReadAll(r.Body)
		r.Body.Close()
		tunnels.Lock()
		if _, ok := tunnels.m[r.Header.Get("token")]; ok {
			tunnels.m[r.Header.Get("token")].conn.Write(content)
		}
		tunnels.Unlock()
		w.Header().Add("padding", padding())
		w.WriteHeader(http.StatusNoContent)
	})

	http.HandleFunc("/retrieve", func(w http.ResponseWriter, r *http.Request) {
		if r.FormValue("password") != *password {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		tunnels.Lock()
		if _, ok := tunnels.m[r.FormValue("token")]; ok {
			tunnels.Unlock()
			for {
				tunnels.Lock()
				if len(tunnels.m[r.FormValue("token")].data) > 0 {
					break
				}
				tunnels.Unlock()
				time.Sleep(5 * time.Millisecond)
			}

			var body bytes.Buffer
			zw := gzip.NewWriter(&body)
			zw.Write(tunnels.m[r.FormValue("token")].data)
			zw.Close()
			w.Header().Add("Content-Encoding", "gzip")
			w.Write(body.Bytes())
			tunnels.m[r.FormValue("token")].data = nil
		}
		tunnels.Unlock()
	})

	http.HandleFunc("/done", func(w http.ResponseWriter, r *http.Request) {
		if r.FormValue("password") != *password {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		tunnels.Lock()
		delete(tunnels.m, r.FormValue("token"))
		tunnels.Unlock()
		w.Header().Add("padding", padding())
		w.WriteHeader(http.StatusNoContent)
	})

	fmt.Println("Listening on port 80...")
	http.ListenAndServe(":80", nil)
}
