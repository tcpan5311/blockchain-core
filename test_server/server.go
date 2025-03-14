package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

// StartServer initializes a TCP server to listen for incoming messages
func StartServer(port string, handleMessage func(string)) {
	listener, err := net.Listen("tcp", "localhost:"+port)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Listening on port", port, "...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Connection error:", err)
			continue
		}
		go handleConnection(conn, handleMessage)
	}
}

// handleConnection handles incoming messages from a connection
func handleConnection(conn net.Conn, handleMessage func(string)) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Connection closed.")
			return
		}
		handleMessage(strings.TrimSpace(message))
	}
}

// StartClient reads input from the user and sends messages to the specified peer
func StartClient(peerPort string) {
	fmt.Println("Type messages and press Enter to send. Type 'exit' to quit.")

	for {
		reader := bufio.NewReader(os.Stdin)
		message, _ := reader.ReadString('\n')
		message = strings.TrimSpace(message)

		if message == "exit" {
			fmt.Println("Exiting...")
			return
		}

		SendMessage(peerPort, message)
	}
}

// SendMessage connects to a peer and sends a message
func SendMessage(port, message string) {
	conn, err := net.Dial("tcp", "localhost:"+port)
	if err != nil {
		fmt.Println("Error connecting to peer:", err)
		return
	}
	defer conn.Close()

	_, err = conn.Write([]byte(message + "\n"))
	if err != nil {
		fmt.Println("Error sending message:", err)
	}
}

func main() {
	var wg sync.WaitGroup

	// Define the node's port
	var myPort, peerPort string
	fmt.Print("Enter your port (e.g., 9000): ")
	fmt.Scanln(&myPort)
	fmt.Print("Enter peer's port (e.g., 9001): ")
	fmt.Scanln(&peerPort)

	// Start listening for incoming messages
	wg.Add(1)
	go func() {
		defer wg.Done()
		StartServer(myPort, func(message string) {
			fmt.Println("Received:", message)
		})
	}()

	// Start sending messages to peer
	wg.Add(1)
	go func() {
		defer wg.Done()
		StartClient(peerPort)
	}()

	wg.Wait() // Wait for both goroutines
}
