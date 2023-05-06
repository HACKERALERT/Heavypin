Heavypin is a lightweight HTTPS-based proxy for bypassing firewalls.

<strong>This documentation is a work in progress.</strong>

# Server
Install Heavypin on a server that will act as the proxy:
```
go install github.com/HACKERALERT/Heavypin/heavypin-server@latest
```
To start the server:
```
heavypin-server
```
This will start an HTTP server on `:8080`. You can use this port directly, however, keep in mind that for optimal circumvention and obfuscation of the connection, you should using a reverse proxy to forward `:443` to `:8080` to make it look like a normal HTTP server.

# Client
On your local machine, you will need the client to connect to the server:
```
go install github.com/HACKERALERT/Heavypin/heavypin-client@latest
```
To connect to the server:
```
heavypin-client -s <server_hostname:server_port>
```

# How It Works
Heavypin is an HTTPS-based proxy, meaning that instead of working with raw sockets, it uses HTTPS as a transport layer for data. Heavypin starts a local HTTP proxy server to catch your browser traffic, and then uses many HTTPS requests to the server to forward your traffic to the destination server. Then, your device will periodically check the server and fetch the response. Because everything is done over HTTPS, to an unsuspecting observer, your connection to the server looks like normal web traffic. This makes it possible to bypass firewalls that block certain ports and protocols.
