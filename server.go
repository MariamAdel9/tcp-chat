package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	clients        = make(map[net.Conn]string)
	usedNames      = make(map[string]bool)
	removedClients = make(map[net.Conn]string) // Map to track removed clients
	clientsMutex   sync.Mutex
	messages       []string
	timestamp = time.Now().Format("2006-01-02 15:04:05")

)

func logFile() *os.File {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("logfile_%s.log", timestamp)
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Failed to create log file: %v", err)
	}
	// Redirect the global log output to the file
	log.SetOutput(file)
	return file
}

func handleConnection(conn net.Conn) {
	
	defer conn.Close()
// Check if the number of connected clients exceeds the limit (10)
clientsMutex.Lock()
if len(clients) >= 10 {
	conn.Write([]byte("Server is full. Please try again later.\n"))
	log.Printf("Connection from %v rejected: server is full", conn.RemoteAddr())
	clientsMutex.Unlock()
	return
}
clientsMutex.Unlock()

conn.Write([]byte("Currently there are "))
conn.Write([]byte(fmt.Sprintf("Number of connected clients: %d\n", len(clients))))
	conn.Write([]byte("Welcome to TCP-Chat!\n" +
		"		         _nnnn_\n" +
		"		        dGGGGMMb\n" +
		"		       @p~qp~~qMb\n" +
		"		       M|@||@) M|\n" +
		"		       @,----.JM|\n" +
		"		      JS^\\__/  qKL\n" +
		"		     dZP        qKRb\n" +
		"		    dZP          qKKb\n" +
		"		   fZP            SMMb\n" +
		"		   HZM            MMMM\n" +
		"		   FqM            MMMM\n" +
		"		  __| \".        |\\dS\"qML\n" +
		"		  |    `.       | `' \\Zq\n" +
		"		 _)      \\.___.,|     .'\n" +
		"		\\____   )MMMMMP|   .' \n" +
		"		     `-'       `--'\n" +
		"[ENTER YOUR NAME]:"))


	reader := bufio.NewReader(conn)

	var name string
	for {
		nameInput, _ := reader.ReadString('\n')
		name = strings.TrimSpace(nameInput)

		if name == "" {
			conn.Write([]byte("Name cannot be empty. Please try again:\n[ENTER YOUR NAME]: "))
			continue
		}

		clientsMutex.Lock()
		if usedNames[name] {
			conn.Write([]byte("Your username already exists, please try again:\n[ENTER YOUR NAME]: "))
			clientsMutex.Unlock()
			continue
		}
		// Name is unique
		usedNames[name] = true
		clients[conn] = name
		clientsMutex.Unlock()
		break
	}

		defer appendRemovedClient(conn)


	log.Printf("Client %s has joined the chat\n", name)
	sendPreviousMessages(conn)
	broadcast(fmt.Sprintf("%s has joined the chat...", name), "")
	conn.Write([]byte("["+timestamp+"]"))
                    
	conn.Write([]byte("["+name+"]:"))
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			broadcast(fmt.Sprintf("%s has left the chat...", name), "")
			log.Printf("%s has left the chat...", name)
			return
		}

		message = strings.TrimSpace(message)

		currentName := getClientName(conn) // Fetch the current name from the map
		handleMessage(conn, currentName, message)
	}
}

func handleMessage(conn net.Conn, name, message string) {
	conn.Write([]byte("["+timestamp+"]"))
                    
	conn.Write([]byte("["+name+"]:"))
	if strings.HasPrefix(message, "/name ") {
		newName := strings.TrimSpace(strings.TrimPrefix(message, "/name "))
		if newName == "" {
			conn.Write([]byte("Name cannot be empty.\n"))
			log.Printf("Name cannot be empty.\n")
		} else {
			clientsMutex.Lock()
			if usedNames[newName] {
				conn.Write([]byte("Name already in use.\n"))
				log.Printf("Name already in use.\n")
				clientsMutex.Unlock()
				return
			}
			delete(usedNames, name)
			usedNames[newName] = true
			clients[conn] = newName
			clientsMutex.Unlock()
			broadcast(fmt.Sprintf("%s has changed their name to %s.", name, newName), "")
			log.Printf("%s has changed their name to %s.", name, newName)
		}
	} else if message != "" {
		fullMessage := fmt.Sprintf("[%s][%s]: %s", timestamp, name, message)
		log.Printf("[%s][%s]: %s", timestamp, name, message)
		broadcast(fullMessage, conn.RemoteAddr().String())
	}
}

func appendRemovedClient(conn net.Conn) {
	clientsMutex.Lock()
	name := clients[conn]
	removedClients[conn] = name // Add to removed clients instead of deleting
	delete(usedNames, name)
	delete(clients, conn)
	log.Printf("Connection %v was left", conn.RemoteAddr())
	clientsMutex.Unlock()

	//broadcast(fmt.Sprintf("%s has left the chat...", name), "")
}

func broadcast(message, senderAddr string) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	messages = append(messages, message)
	fmt.Println(message)

	for client := range clients {
		if client.RemoteAddr().String() != senderAddr {
			_, err := client.Write([]byte(message + "\n"))
			if err != nil {
				log.Printf("Error writing to client %s: %v", clients[client], err)
				client.Close()
				removedClients[client] = clients[client] // Add to removed clients instead of deleting
				delete(usedNames, clients[client])
				delete(clients, client)
			}
		}
	}
}

func sendPreviousMessages(conn net.Conn) {
	for _, msg := range messages {
		conn.Write([]byte(msg + "\n"))
	}
}

func getClientName(conn net.Conn) string {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()
	return clients[conn]
}
