package main

import (
	"fmt"
	"log"
	"net"
)

var (
	serverAddr *net.UDPAddr
	serverConn *net.UDPConn
)

func init() {
	var err error
	serverAddr, err = net.ResolveUDPAddr("udp", ":10001")
	if err != nil {
		log.Fatalln(err)
	}
	serverConn, err = net.ListenUDP("udp", serverAddr)
	if err != nil {
		log.Fatalln(err)
	}
}

func receiver(r chan string) {
	buf := make([]byte, 65536)
	for {
		n, addr, err := serverConn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println(addr, err)
			continue
		}
		r <- string(buf[0:n])
	}
}

func main() {
	r := make(chan string, 512)
	go receiver(r)

	for {
		fmt.Println(<-r)
	}
}
