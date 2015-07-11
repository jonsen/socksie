package main

// socksie is a SOCKS4/5 compatible proxy that forwards connections via
// SSH to a remote host

import (
	"errors"
	"flag"
	"fmt"
	"github.com/golang/crypto/ssh"
	"log"
	"net"
	"os"
	"time"

	"github.com/davecheney/profile"
)

var (
	USER  = flag.String("user", os.Getenv("USER"), "ssh username")
	HOST  = flag.String("host", "127.0.0.1", "ssh server hostname")
	PORT  = flag.Int("port", 7070, "socksie listening port")
	RPORT = flag.Int("rport", 22, "ssh server port")
	PASS  = flag.String("pass", os.Getenv("SOCKSIE_SSH_PASSWORD"), "ssh password")

	// ssh client hand
	conn *ssh.Client
)

func init() { flag.Parse() }

type Dialer interface {
	DialTCP(net string, laddr, raddr *net.TCPAddr) (net.Conn, error)
}

func sshDial() (err error) {
	var auths []ssh.AuthMethod
	if *PASS != "" {
		auths = append(auths, ssh.Password(*PASS))
	}

	config := &ssh.ClientConfig{
		User: *USER,
		Auth: auths,
	}
	addr := fmt.Sprintf("%s:%d", *HOST, *RPORT)
	conn, err = ssh.Dial("tcp", addr, config)

	if err != nil {
		errstr := fmt.Sprintf("unable to connect to [%s]: %v", addr, err)
		return errors.New(errstr)
	}

	return
}

func sshCheck() (ok bool) {

	session, err := conn.NewSession()
	if err != nil {
		log.Println("create session of ssh error", err)
		return false
	} else {
		//log.Println("create session ok")
	}

	defer session.Close()

	return true
}

func sshReconnect() {

	for {

		if !sshCheck() {
			if err := sshDial(); err != nil {
				log.Println(err)
			} else {
				log.Println("reconnect ssh ok")
			}
		}
		time.Sleep(5 * time.Second)
	}

}

func main() {
	defer profile.Start(profile.CPUProfile).Stop()

	err := sshDial()
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	go sshReconnect()

	addr := fmt.Sprintf("%s:%d", "0.0.0.0", *PORT)
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
