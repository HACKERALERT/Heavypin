Heavypin is a lightweight HTTPS-based proxy for bypassing firewalls.

# Installation
You will need to install Heavypin on a server that will act as the proxy:
```
go install github.com/HACKERALERT/Heavypin/heavypin-server@latest
```
To start the server:
```
heavypin-server
```

Then, on your local machine, you will need the client to connect to the server:
```
go install github.com/HACKERALERT/Heavypin/heavypin-client@latest
```
To connect to the server:
```
heavypin-client
```

# How It Works
Heavypin is an HTTPS-based proxy, meaning that instead of working with raw sockets, it uses HTTPS as a transport layer for data. Heavypin starts a local HTTP proxy server to catch your browser traffic, and then uses many HTTPS requests to the server to forward your traffic to the destination server. Then, your device will periodically check the server and fetch the response. Because everything is done over HTTPS, to an unsuspecting observer, your connection to the server looks like normal web traffic. This makes it possible to bypass firewalls that block certain ports and protocols.
