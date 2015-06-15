package main

// socksie is a SOCKS4/5 compatible proxy that forwards connections via
// SSH to a remote host

import (
	"github.com/golang/crypto/ssh"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/davecheney/profile"
)

var (
	USER = flag.String("user", os.Getenv("USER"), "ssh username")
	HOST = flag.String("host", "127.0.0.1", "ssh server hostname")
	PORT = flag.Int("port", 7070, "socksie listening port")
	PASS = flag.String("pass", os.Getenv("SOCKSIE_SSH_PASSWORD"), "ssh password")
)

func init() { flag.Parse() }

type Dialer interface {
	DialTCP(net string, laddr, raddr *net.TCPAddr) (net.Conn, error)
}

func main() {
	defer profile.Start(profile.CPUProfile).Stop()

	var auths []ssh.AuthMethod
	if *PASS != "" {
		auths = append(auths, ssh.Password(*PASS))
	}

	config := &ssh.ClientConfig{
		User: *USER,
		Auth: auths,
	}
	addr := fmt.Sprintf("%s:%d", *HOST, 22)
	conn, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		log.Fatalf("unable to connect to [%s]: %v", addr, err)
	}
	defer conn.Close()

	addr = fmt.Sprintf("%s:%d", "0", *PORT)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("unable to listen on SOCKS port [%s]: %v", addr, err)
	}
	defer l.Close()
	log.Printf("listening for incoming SOCKS connections on [%s]\n", addr)

	for {
		c, err := l.Accept()
		if err != nil {
			log.Fatalf("failed to accept incoming SOCKS connection: %v", err)
		}
		accepted.Inc()
		go handleConn(c.(*net.TCPConn), conn)
	}
	log.Println("waiting for all existing connections to finish")
	connections.Wait()
	log.Println("shutting down")
}
