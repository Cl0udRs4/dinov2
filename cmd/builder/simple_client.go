package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

// SimpleClient is a basic client implementation for testing
const SimpleClientTemplate = `package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	// Parse command line flags
	serverAddr := flag.String("server", "", "C2 server address")
	protocolList := flag.String("protocol", "tcp", "Comma-separated list of protocols to use (tcp,http,websocket)")
	flag.Parse()

	if *serverAddr == "" {
		fmt.Println("Error: Server address is required")
		flag.Usage()
		os.Exit(1)
	}

	// Parse protocol list
	protocols := strings.Split(*protocolList, ",")
	if len(protocols) == 0 {
		fmt.Println("Error: At least one valid protocol must be specified")
		flag.Usage()
		os.Exit(1)
	}

	fmt.Println("Client started with configuration:")
	fmt.Println("- Server:", *serverAddr)
	fmt.Println("- Protocols:", *protocolList)

	// Connect to server
	conn, err := net.Dial("tcp", *serverAddr)
	if err != nil {
		fmt.Printf("Error connecting to server: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Println("C2 Client started. Connected to server:", *serverAddr)
	fmt.Println("Using protocols:", *protocolList)

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start heartbeat
	go func() {
		for {
			time.Sleep(30 * time.Second)
			_, err := conn.Write([]byte("heartbeat"))
			if err != nil {
				fmt.Printf("Error sending heartbeat: %v\n", err)
				return
			}
		}
	}()

	// Wait for termination signal
	<-sigChan
	fmt.Println("\nShutting down client...")
	
	fmt.Println("Client shutdown complete.")
}
`
