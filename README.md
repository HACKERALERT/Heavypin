# Heavypin
A lightweight proxy designed to bypass firewalls by transferring traffic through HTTPS requests.

# Installation
You will need both a server and a client application for Heavypin to work. Make sure you have <a href="https://go.dev/dl/">Go</a> installed.

## Server
Install Heavypin on the server that will act as the proxy:
```
go install github.com/HACKERALERT/Heavypin/heavypin-server@latest
```
To start the proxy server:
```
heavypin-server -p password
```
This will start an HTTPS server on `:443` using a self-signed RSA 4096 certificate generated automatically on runtime.

## Client
On your local machine, you will need the client application to connect to the server:
```
go install github.com/HACKERALERT/Heavypin/heavypin-client@latest
```
Connect to the server by passing in the server's IP address and password:
```
heavypin-client -s https://<server_ip> -p password
```
Once the client application connects to the server, it will start a local HTTP proxy on `:8000`. You can then use `http://localhost:8000` as a proxy in your browser or application to securely access the free and open Internet.

# Goals
Heavypin is a mostly experimental and proof-of-concept project to demonstrate and implement some simple circumvention techniques. It's simple and limited in functionality, but usable enough that I felt it should be released into the public. As such, I don't plan to make Heavypin the next Shadowsocks or V2Ray by constantly adding new obfuscation and circumvention techniques. What Heavypin is now is what it will remain. Use this software for what it is, not what it may become.

# Inspiration
The name "Heavypin" comes from "<strong>H</strong>TTPS <strong>VPN</strong>", since it is essentially a "VPN" running over HTTPS.

# How It Works
Heavypin is an HTTPS-based proxy, meaning that instead of working with raw sockets, it uses HTTPS as a transport layer for tunneling data. Heavypin starts a local HTTP proxy server to catch your browser's traffic, and then uses many HTTPS requests to the proxy server to forward your traffic to the destination server. Then, through HTTP long polling, the client will fetch responses to previous requests from the proxy server and stream them back to the browser through the local HTTP proxy. Because everything is done over HTTPS, your connection to the proxy server looks like normal web traffic to an unsuspecting observer. This makes it possible to bypass firewalls that block certain ports and protocols. For further resistance against censorship, all traffic between the client and proxy server is randomly padded to protect against basic forms of traffic analysis, and the proxy server is protected against active probing by requiring a password to function. Requests to the proxy server that don't supply the correct password in the header or form data will receive an inconspicuous 404 Not Found, effectively concealing the actual proxy server that lies beneath it.
