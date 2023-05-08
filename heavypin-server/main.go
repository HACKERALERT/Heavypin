package main

import (
	"encoding/hex"
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
		sync.Mutex
		m map[string]*Tunnel
	}{m: make(map[string]*Tunnel)}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	http.HandleFunc("/create", func(w http.ResponseWriter, r *http.Request) {
		conn, err := net.DialTimeout("tcp", r.FormValue("host"), 10*time.Second)
		if err != nil {
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
			conn.SetDeadline(time.Now().Add(10 * time.Second))
			for {
				data := make([]byte, 1<<16)
				size, err := conn.Read(data)
				if err != nil {
					for {
						tunnels.Lock()
						data := tunnels.m[r.FormValue("token")].data
						tunnels.Unlock()
						if len(data) == 0 {
							break
						}
						time.Sleep(50 * time.Millisecond)
					}
					tunnels.Lock()
					tunnels.m[r.FormValue("token")].data = make([]byte, 1<<16+1)
					tunnels.Unlock()
					break
				}
				if size != 0 {
					conn.SetDeadline(time.Now().Add(10 * time.Second))
					data = data[:size]
					tunnels.Lock()
					tunnels.m[r.FormValue("token")].data = append(tunnels.m[r.FormValue("token")].data, data...)
					tunnels.Unlock()
				}
				time.Sleep(50 * time.Millisecond)
			}
			conn.Close()
		}()
		w.WriteHeader(http.StatusNoContent)
	})

	http.HandleFunc("/proxy", func(w http.ResponseWriter, r *http.Request) {
		content, _ := hex.DecodeString(r.FormValue("content"))
		tunnels.Lock()
		if _, ok := tunnels.m[r.FormValue("token")]; ok {
			tunnels.m[r.FormValue("token")].conn.Write(content)
		}
		tunnels.Unlock()
		w.WriteHeader(http.StatusNoContent)
	})

	http.HandleFunc("/retrieve", func(w http.ResponseWriter, r *http.Request) {
		tunnels.Lock()
		if _, ok := tunnels.m[r.FormValue("token")]; ok {
			w.Write(tunnels.m[r.FormValue("token")].data)
			tunnels.m[r.FormValue("token")].data = nil
		}
		tunnels.Unlock()
	})

	http.HandleFunc("/done", func(w http.ResponseWriter, r *http.Request) {
		tunnels.Lock()
		delete(tunnels.m, r.FormValue("token"))
		tunnels.Unlock()
		w.WriteHeader(http.StatusNoContent)
	})

	http.ListenAndServe(":8080", nil)
}
