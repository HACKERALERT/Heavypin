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
This will start an HTTP server on `:80`. You can use this port directly as is, however, you should setup a reverse HTTPS proxy (using Nginx or Apache) from `:443` to `:80` if possible to better obfuscate the connection and hide the fact that you are proxying your traffic through that port. This makes the connection much harder to detect and block, especially for censors that actively monitor your network traffic. For the password, choose anything that is reasonably long and unguessable. It's only used to protect against active probing, though, so the password you choose won't affect the security or privacy of your connection.

## Client
On your local machine, you will need the client application to connect to the server:
```
go install github.com/HACKERALERT/Heavypin/heavypin-client@latest
```
Connect to the server by passing in the server's address and password:
```
heavypin-client -s http(s)://<server_hostname_or_ip> -p password
```
For example, if you are hosting directly on `:80` and have the server IP:
```
heavypin-client -s http://1.2.3.4 -p password
```
Or if you are reverse proxying `:443` to `:80` and have a hostname:
```
heavypin-client -s https://myhostname.com -p password
```
Once the client application connects to the server, it will start a local HTTP proxy on `:8000`. You can then use `http://localhost:8000` as a proxy in your browser or application to securely access the free and open Internet.

# Goals
Heavypin is a mostly experimental and proof-of-concept project to demonstrate and implement some simple circumvention techniques. It's simple and limited in functionality, but usable enough that I felt it should be released into the public. As such, I don't plan to make Heavypin the next Shadowsocks or V2Ray by constantly adding new obfuscation and circumvention techniques. What Heavypin is now is what it will remain. Use this software for what it is, not what it may become.

# Inspiration
The name "Heavypin" comes from "<strong>H</strong>TTPS <strong>VPN</strong>", which makes sense since it is essentially a "VPN" running over HTTPS.

# How It Works
Heavypin is an HTTPS-based proxy, meaning that instead of working with raw sockets, it uses HTTPS as a transport layer for tunneling data. Heavypin starts a local HTTP proxy server to catch your browser's traffic, and then uses many HTTPS requests to the proxy server to forward your traffic to the destination server. Then, through HTTP long polling, the client will fetch responses to previous requests from the proxy server and stream them back to the browser through the local HTTP proxy. Because everything is done over HTTPS, or at least should be, your connection to the proxy server looks like normal web traffic to an unsuspecting observer. This makes it possible to bypass firewalls that block certain ports and protocols. For further resistance against censorship, all traffic between the client and proxy server is randomly padded to protect against basic forms of traffic analysis, and the proxy server is protected against active probing by requiring a password to function. Requests to the proxy server that don't supply the correct password in the header or form data will receive an inconspicuous 404 Not Found, effectively concealing the actual proxy server that lies beneath it.
