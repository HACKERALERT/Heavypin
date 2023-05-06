package main

import (
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

type Tunnel struct {
	conn net.Conn
	data []byte
}

func main() {
	tunnels := struct {
		sync.RWMutex
		m map[string]*Tunnel
	}{m: make(map[string]*Tunnel)}

	http.HandleFunc("/create", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Creating tunnel", r.FormValue("token"), "to", r.FormValue("host"))
		conn, _ := net.DialTimeout("tcp", r.FormValue("host"), 10*time.Second)
		tunnels.Lock()
		tunnels.m[r.FormValue("token")] = &Tunnel{
			conn: conn,
			data: nil,
		}
		tunnels.Unlock()
		go func() {
			for {
				data := make([]byte, 1<<15)
				tunnels.Lock()
				if _, ok := tunnels.m[r.FormValue("token")]; conn == nil || !ok {
					tunnels.Unlock()
					break
				}
				tunnels.Unlock()
				size, err := conn.Read(data)
				if err != nil {
					for {
						tunnels.Lock()
						data := tunnels.m[r.FormValue("token")].data
						tunnels.Unlock()
						if len(data) == 0 {
							break
						}
						time.Sleep(100 * time.Millisecond)
					}
					tunnels.Lock()
					tunnels.m[r.FormValue("token")].data = make([]byte, 1<<15+1)
					tunnels.Unlock()
					break
				}
				fmt.Println("Received", size, "bytes for", r.FormValue("token"))
				data = data[:size]
				tunnels.Lock()
				tunnels.m[r.FormValue("token")].data = append(tunnels.m[r.FormValue("token")].data, data...)
				tunnels.Unlock()
				time.Sleep(100 * time.Millisecond)
			}
			if conn != nil {
				conn.Close()
			}
		}()
		w.WriteHeader(http.StatusNoContent)
	})

	http.HandleFunc("/proxy", func(w http.ResponseWriter, r *http.Request) {
		tunnels.Lock()
		if _, ok := tunnels.m[r.FormValue("token")]; !ok || tunnels.m[r.FormValue("token")].conn == nil {
			w.WriteHeader(http.StatusNoContent)
			tunnels.Unlock()
			return
		}
		fmt.Println("Proxying request for", r.FormValue("token"))
		content, _ := hex.DecodeString(r.FormValue("content"))
		tunnels.m[r.FormValue("token")].conn.Write(content)
		tunnels.Unlock()
		w.WriteHeader(http.StatusNoContent)
	})

	http.HandleFunc("/retrieve", func(w http.ResponseWriter, r *http.Request) {
		tunnels.Lock()
		if _, ok := tunnels.m[r.FormValue("token")]; !ok {
			w.Write(make([]byte, 1<<15+1))
			tunnels.Unlock()
			return
		}
		fmt.Println("Retrieving", len(tunnels.m[r.FormValue("token")].data), "bytes for", r.FormValue("token"))
		w.Write(tunnels.m[r.FormValue("token")].data)
		tunnels.m[r.FormValue("token")].data = nil
		tunnels.Unlock()
	})

	http.HandleFunc("/done", func(w http.ResponseWriter, r *http.Request) {
		tunnels.Lock()
		delete(tunnels.m, r.FormValue("token"))
		tunnels.Unlock()
	})

	fmt.Println("Listening on :8080...")
	http.ListenAndServe(":8080", nil)
}
