Heavypin is a lightweight HTTPS-based proxy for bypassing firewalls.

<strong>This documentation is a work in progress.</strong>

# Server
Install Heavypin on the server that will act as the proxy:
```
go install github.com/HACKERALERT/Heavypin/heavypin-server@latest
```
To start the proxy server:
```
heavypin-server
```
This will start an HTTP server on `:8080`. You can use this port directly as is, however, you should setup a reverse HTTPS proxy from `:443` to `:8080` if possible to better obfuscate the connection and hide the fact that you are proxying your traffic through that port. This makes the connection harder to detect and therefore block.

# Client
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
Once the client application connects to the server, it will start a local HTTP proxy on `:8888`. You can then use `http://localhost:8888` as a proxy in your browser to securely access the free and open Internet.

# Goals
Heavypin is a mostly experimental and proof-of-concept project to demonstrate and implement some simple circumvention techniques. It's simple and limited in functionality, but usable enough that I felt it's good enough to be released into the public. As such, I don't plan to make Heavypin the next Shadowsocks or V2Ray by constantly adding new obfuscation and circumvention techniques. What Heavypin is now is what it will remain. Use this software for what it is.

# Inspiration
The name "Heavypin" comes from "<strong>H</strong>TTPS <strong>VPN</strong>", which makes sense since it is essentially a VPN running over HTTPS.

# How It Works
Heavypin is an HTTPS-based proxy, meaning that instead of working with raw sockets, it uses HTTPS as a transport layer for data. Heavypin starts a local HTTP proxy server to catch your browser traffic, and then uses many HTTPS requests to the server to forward your traffic to the destination server. Then, your device will periodically check the server and fetch the response. Because everything is done over HTTPS, to an unsuspecting observer, your connection to the server looks like normal web traffic. This makes it possible to bypass firewalls that block certain ports and protocols.
