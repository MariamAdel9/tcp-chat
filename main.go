package main

import (
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	logFile := logFile()
	defer logFile.Close()
	if len(os.Args) > 2 {
		fmt.Println("[USAGE]: ./TCPChat.sh $port")
		return
	}

	port := "8989" // Default port
	if len(os.Args) == 2 {
		port = os.Args[1]
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer listener.Close()
	fmt.Printf("Server is listening on port %s", port)
	log.Printf("Server is listening on port %s", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go handleConnection(conn)
	}
}
