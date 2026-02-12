package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	nwep "github.com/usenwep/nwep-go"
	"github.com/usenwep/velocity"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "6937"
	}

	srv, err := velocity.New(":"+port,
		velocity.WithKeyFile("server.key"),
		velocity.OnStart(func(s *velocity.Server) {
			addr := s.URL("/")
			publicIP := os.Getenv("PUBLIC_IP")
			if publicIP == "" {
				publicIP = discoverPublicIP()
			}
			if publicIP != "" {
				log.Printf("public ip: %s", publicIP)
				portNum, _ := strconv.Atoi(port)
				if u, err := nwep.FormatURL(net.ParseIP(publicIP), uint16(portNum), s.NodeID(), "/"); err == nil {
					addr = u
				}
			} else {
				log.Printf("public ip: unknown (using local address)")
			}
			log.Printf("node address: %s", addr)
		}),
		velocity.WithOnConnect(func(conn *nwep.Conn) {
			log.Printf("peer connected: %s", conn.NodeID())
		}),
		velocity.WithOnDisconnect(func(conn *nwep.Conn, errCode int) {
			log.Printf("peer disconnected: %s (code %d)", conn.NodeID(), errCode)
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	srv.Handle("/", func(c *velocity.Context) error {
		c.SetHeader("content-type", "text/html")
		return c.OK([]byte(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Meme Server</title>
  <meta property="og:title" content="Meme Server">
  <meta property="og:description" content="A meme server on the new web.">
  <meta property="og:type" content="website">
</head>
<body>
  <h1>Meme Server</h1>
  <p>A meme server on the new web.</p>
</body>
</html>`))
	})

	srv.Handle("/hello", func(c *velocity.Context) error {
		return c.OK([]byte("hello from velocity"))
	})

	log.Fatal(srv.Run())
}

func discoverPublicIP() string {
	resp, err := http.Get("https://api.ipify.org")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	ip := strings.TrimSpace(string(body))
	if net.ParseIP(ip) == nil {
		return ""
	}
	return ip
}
