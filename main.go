package main

import (
	_ "embed"
	"encoding/json"
	"io"
	"log"
	"math/rand/v2"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	nwep "github.com/usenwep/nwep-go"
	"github.com/usenwep/velocity"
)

//go:embed reasons.json
var reasonsJSON []byte

var reasons []string

func main() {
	if err := json.Unmarshal(reasonsJSON, &reasons); err != nil {
		log.Fatalf("failed to load reasons: %v", err)
	}
	log.Printf("loaded %d reasons", len(reasons))

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
			publicPort := os.Getenv("PUBLIC_PORT")
			if publicPort == "" {
				publicPort = port
			}
			if publicIP != "" {
				portNum, _ := strconv.Atoi(publicPort)
				log.Printf("public ip: %s:%s", publicIP, publicPort)
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
		return c.OK([]byte(reasons[rand.IntN(len(reasons))]))
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
