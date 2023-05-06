# Heavypin
Heavypin is a lightweight proxy designed to bypass firewalls by transferring traffic through HTTPS requests.

# Installation
You will need both a server and a client application for Heavypin to work. Make sure you have <a href="https://go.dev/dl/">Go</a> installed.

## Server
Install Heavypin on the server that will act as the proxy:
```
go install github.com/HACKERALERT/Heavypin/heavypin-server@latest
```
To start the proxy server:
```
heavypin-server
```
This will start an HTTP server on `:8080`. You can use this port directly as is, however, you should setup a reverse HTTPS proxy from `:443` to `:8080` if possible to better obfuscate the connection and hide the fact that you are proxying your traffic through that port. This makes the connection much harder to detect and block.

## Client
On your local machine, you will need the client application to connect to the server:
```
go install github.com/HACKERALERT/Heavypin/heavypin-client@latest
```
Connect to the server by passing in the server's address:
```
heavypin-client -s "http(s)://<server_hostname_or_ip>:<server_port>"
```
For example, if you are hosting directly on `:8080` and have the server IP:
```
heavypin-client -s "http://1.1.1.1:8080"
```
Or if you are reverse proxying `:443` to `:8080` and have a hostname:
```
heavypin-client -s "https://myhostname.com"
```
Once the client application connects to the server, it will start a local HTTP proxy on `:8888`. You can then use `http://localhost:8888` as a proxy in your browser to securely access the free and open Internet. Note that accessing an insecure HTTP website in your browser will yield a `405 Method Not Allowed` error -- this is intentionally done to ensure that there is always at least one layer of encryption between your browser and the target website. To avoid this, just remember to always use the secure HTTPS version of the website.

# Goals
Heavypin is a mostly experimental and proof-of-concept project to demonstrate and implement some simple circumvention techniques. It's simple and limited in functionality, but usable enough that I felt it should be released into the public. As such, I don't plan to make Heavypin the next Shadowsocks or V2Ray by constantly adding new obfuscation and circumvention techniques. What Heavypin is now is what it will remain. Use this software for what it is, not what it may become.

# Inspiration
The name "Heavypin" comes from "<strong>H</strong>TTPS <strong>VPN</strong>", which makes sense since it is essentially a VPN running over HTTPS.

# Internals
Heavypin is an HTTPS-based proxy, meaning that instead of working with raw sockets, it uses HTTPS as a transport layer for tunneling data. Heavypin starts a local HTTP proxy server to catch your browser's traffic, and then uses many HTTPS requests to the proxy server to forward your traffic to the destination server. Then, the client will periodically poll the server for responses to previous requests and fetch them if available. After the responses are fetched, they are streamed back to the browser through the local HTTP proxy. Because everything is done over HTTPS, your connection to the server looks like normal web traffic to an unsuspecting observer. This makes it possible to bypass firewalls that block certain ports and protocols.
