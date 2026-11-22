package main

import (
	"fmt"
	"flag"
	"os"
	"strings"
	"net"
	"time"
	"errors"
)

func printUsage(execStr string) {
	fmt.Fprintf(os.Stderr, "Usage: %s start -u <username>\n", execStr)
}

func logServerError(serverLn net.Listener, message string, a ...any) {
	fmt.Printf("[ERROR] [%s] [%s] %s", time.Now().Format(time.RFC3339), serverLn.Addr().String(), fmt.Sprintf(message, a...))
}

func logServerWarn(serverLn net.Listener, message string, a ...any) {
	fmt.Printf("[WARN] [%s] [%s] %s", time.Now().Format(time.RFC3339), serverLn.Addr().String(), fmt.Sprintf(message, a...))
}

func logServerInfo(serverLn net.Listener, message string, a ...any) {
	fmt.Printf("[INFO] [%s] [%s] %s", time.Now().Format(time.RFC3339), serverLn.Addr().String(), fmt.Sprintf(message, a...))
}

func handleConnections(serverLn net.Listener, connList []net.Conn) []net.Conn {
	newConnList := make([]net.Conn, 0, len(connList))

	for _, conn := range connList {
		buf := make([]byte, 1024)

		err := conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	        if err != nil {
	        	logServerWarn(serverLn, "Failed to setup read timeout: %s, conn %s\n", err, conn.RemoteAddr().String())
	        }

		bytes_read, err := conn.Read(buf)

		if err != nil {
			switch {
			case errors.Is(err, os.ErrDeadlineExceeded):
				// Timed out due to SetDeadline / SetReadDeadline.
				// Append back to the list for future operations
				newConnList = append(newConnList, conn)
			default:
				logServerWarn(serverLn, "Closing connection %s due to error: %s\n", conn.RemoteAddr().String(), err)
				conn.Close()
			}

			continue
		}

		logServerInfo(serverLn, "Read %d bytes from connection %s\n", bytes_read, conn.RemoteAddr().String())
		// Append back to the list for future operations
		newConnList = append(newConnList, conn)
	}

	return newConnList
}

func startServer(username string) {
	// Listen at localhost. Port 0 allows the port to be automatically chosen
	serverAddressStr := "127.0.0.1:0"
	ln, err := net.Listen("tcp4", serverAddressStr)

	if err != nil {
		panic(err)
	}

	logServerInfo(ln, "Listening at %s\n", ln.Addr().String())
	logServerInfo(ln, "Hello %s!\n", username)

	tcpLn, _ := ln.(*net.TCPListener)

	// Reserve capacity for the list of connection
	connList := make([]net.Conn, 0, 10)

	for {
		// handle existing connections
		connList = handleConnections(ln, connList)

	        // Configure listener to dont block the thread while waiting for a new connection
                if err := tcpLn.SetDeadline(time.Now().Add(2 * time.Second)); err != nil {
			logServerWarn(ln, "Failed to configure listener Accept timeout: %s\n", err)
	        }

		conn, err := ln.Accept()

		if err != nil {
			switch {
			case errors.Is(err, os.ErrDeadlineExceeded):
			default:
			        logServerError(ln, "Failed to establish connection: %s\n", err)
			}
			continue
		}
		
		logServerInfo(ln, "Established connection with %s\n", conn.RemoteAddr().String())
		   
		connList = append(connList, conn)
        }

	ln.Close()

}

func main() {

	serverStartCmd := flag.NewFlagSet("start", flag.ExitOnError)
	serverHostUserFlag := serverStartCmd.String("u", "", "username to use in chat")

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Error: expected at least 2 arguments, but got %d\n", len(os.Args))
		printUsage(os.Args[0])
		os.Exit(1)
	}

	switch os.Args[1] {
	
	case "start":
		serverStartCmd.Parse(os.Args[2:])

		username := strings.TrimSpace(*serverHostUserFlag)

		if len(username) == 0 {
			fmt.Fprintf(os.Stderr, "Error: username should not be empty\n")
		        printUsage(os.Args[0])
			os.Exit(1)
		}

		startServer(username)

	default:
		fmt.Fprintf(os.Stderr, "Error: unexpected command `%s`\n", os.Args[1])
		printUsage(os.Args[0])
		os.Exit(1)
		
	}
}
