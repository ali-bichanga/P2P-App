package main


import (
"fmt"
"io"
"log"
"net"
"os"
"strings"
"sync"
"time"
)


var FILES_LIST = []string{} // Initialize empty file list; files will be read from the directory
var IP_ADDRESS = "192.168.1.159"
var NEIGHBORS []string
var PORTS_OPEN []string


// Track processed request IDs to avoid duplicate handling
var processedRequests = make(map[string]bool)
var processedMutex sync.Mutex


// Helper method to check if an item is in a list
func contains(slice []string, item string) bool {
for _, v := range slice {
if v == item {
return true
}
}
return false
}


// Helper method to check if a request ID has already been processed
func isProcessed(requestID string) bool {
processedMutex.Lock()
defer processedMutex.Unlock()
if processedRequests[requestID] {
return true
}
processedRequests[requestID] = true
return false
}


// //////////////////TRACKER SERVER CONNECTIONS/////////////////////////////////////////////////////////
// This is to register a peer with the server
func registration(address string, udpPort string) {
conn, err := net.Dial("tcp", address)
if err != nil {
log.Fatal(err)
fmt.Printf("Error connecting to server for registration: %s\n", err)
os.Exit(1)
}


// Register IP address and the specified UDP port
registerPeer := "REGISTER- " + IP_ADDRESS + ":" + udpPort
conn.Write([]byte(registerPeer))


// Check if server returned REGISTERED
registered := make([]byte, 1024)
n, err := conn.Read(registered)
if err != nil {
fmt.Println("Error reading registration status from server:", err)
}


response := strings.TrimSpace(string(registered[:n]))
if response == "REGISTERED" {
fmt.Println("Registration successful")
} else {
fmt.Println("Registration failed")
}
}


// Function to request and receive a list of IP addresses from the server
func requestIpFromServer(address string) {
conn, err := net.Dial("tcp", address)
if err != nil {
log.Fatal(err)
fmt.Printf("Error connecting to server for IP request: %s\n", err)
os.Exit(1)
}


requestIPs := "REQUEST_PEERS"
conn.Write([]byte(requestIPs))


ipList := make([]byte, 1024)


n, err := conn.Read(ipList)
if err != nil {
fmt.Println("Error with list of IPs:", err)
return
}


// The peer will receive a list of IP addresses separated by commas
ipString := strings.TrimSpace(string(ipList[:n]))
newIpList := strings.Split(ipString, ",") // Split by commas to get individual addresses


fmt.Println("Received from server:", newIpList)


// Add the parsed IP addresses to the NEIGHBORS list
NEIGHBORS = append(NEIGHBORS, newIpList...)
}


// /////////////////////CLIENT CONNECTIONS////////////////////////////////////////////////////////


// ///UDP/////////////
// Requesting a file through UDP
func fileRequest(address string, fileName string, requestAddress string, requestID string) {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
	log.Fatal(err)
	}


	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
	log.Fatal(err)
	fmt.Printf("Error starting UDP connection with peer: %v\n", err)
	}
	defer conn.Close()


	message := fmt.Sprintf("FILENAME-%s-%s-%s", fileName, requestAddress, requestID)
	_, err = conn.Write([]byte(message))
	if err != nil {
	log.Fatal(err)
	fmt.Printf("Error sending file name to peer: %v\n", err)
	}
	fmt.Printf("File request for '%s' sent to %s\n", fileName, address)
}


// ////TCP//////////////////////////
func connectToPeer(address string, fileName string) {
conn, err := net.Dial("tcp", address)
if err != nil {
log.Fatal(err)
fmt.Printf("Error forming TCP connection with peer: %v\n", err)
}
defer conn.Close()


// Send the file name
_, err = conn.Write([]byte(fileName))
if err != nil {
fmt.Printf("Error sending fileName to peer(TCP): %v\n", err)
return
}


// Receive the file
receiveFile(conn, fileName)
}


// Receive the file over TCP
func receiveFile(conn net.Conn, fileName string) {
file, err := os.Create(fileName)
if err != nil {
fmt.Printf("Error creating file %s: %v\n", fileName, err)
return
}
defer file.Close()


fmt.Printf("Receiving file '%s'...\n", fileName)


buf := make([]byte, 1024)
for {
n, err := conn.Read(buf)
if n > 0 {
_, writeErr := file.Write(buf[:n])
if writeErr != nil {
fmt.Printf("Error writing to file %s: %v\n", fileName, writeErr)
return
}
}
if err != nil {
if err == io.EOF {
break
}
fmt.Printf("Error receiving file %s: %v\n", fileName, err)
return
}
}
fmt.Printf("File '%s' received successfully.\n", fileName)
}


// Send the file over TCP
func sendFile(conn net.Conn, fileName string) {
file, err := os.Open(fileName)
if err != nil {
fmt.Printf("Error opening file %s: %v\n", fileName, err)
conn.Write([]byte("ERROR: File not found\n"))
return
}
defer file.Close()


fmt.Printf("Sending file '%s'...\n", fileName)


buf := make([]byte, 1024)
for {
n, err := file.Read(buf)
if n > 0 {
_, writeErr := conn.Write(buf[:n])
if writeErr != nil {
fmt.Printf("Error sending file %s: %v\n", fileName, writeErr)
return
}
}
if err != nil {
if err == io.EOF {
break
}
fmt.Printf("Error reading file %s: %v\n", fileName, err)
return
}
}
fmt.Printf("File '%s' sent successfully.\n", fileName)
}


// /////////////////////SERVER SIDE FUNCTIONS///////////////////////////////////////////


// Listen for UDP file requests
func receiveUDPFileRequests(port string) {
addr, err := net.ResolveUDPAddr("udp", ":"+port)
if err != nil {
log.Fatal(err)
}


ln, err := net.ListenUDP("udp", addr)
if err != nil {
log.Fatal(err)
}
defer ln.Close()


fmt.Printf("Listening for UDP requests on port %s\n", port)


buf := make([]byte, 1024)
for {
n, addr, err := ln.ReadFromUDP(buf)
if err != nil {
log.Fatal(err)
}


request := strings.TrimSpace(string(buf[:n]))
fmt.Printf("Received request: %s from %s\n", request, addr)


if strings.HasPrefix(request, "FILENAME") {
parts := strings.SplitN(request, "-", 4)
if len(parts) == 4 {
filename := parts[1]
requestIP := parts[2]
requestID := parts[3]


fmt.Printf("Request details: filename=%s, requestIP=%s, requestID=%s\n", filename, requestIP, requestID)


if isProcessed(requestID) {
fmt.Printf("Duplicate request ID %s ignored.\n", requestID)
continue
}


if contains(FILES_LIST, filename) {
fmt.Printf("File '%s' found locally. Sending to %s\n", filename, requestIP)
go connectToPeer(requestIP, filename)
} else {
fmt.Printf("File '%s' not found locally. Forwarding request to neighbors.\n", filename)
for _, neighbor := range NEIGHBORS {
if neighbor != requestIP {
go fileRequest(neighbor, filename, requestIP, requestID)
}
}
}
} else {
fmt.Println("Invalid request format.")
}
}
}
}


func main() {
if len(os.Args) < 4 {
fmt.Println("Usage: go run peer.go <tracker_address> <udp_port> <peer_directory>")
os.Exit(1)
}


trackerAddress := os.Args[1]
udpPort := os.Args[2]
peerDirectory := os.Args[3]


// Change working directory to the peer's directory
err := os.Chdir(peerDirectory)
if err != nil {
log.Fatalf("Error changing directory to %s: %v\n", peerDirectory, err)
}
fmt.Printf("Working directory set to %s\n", peerDirectory)


// Populate FILES_LIST from the directory
files, err := os.ReadDir(".")
if err != nil {
log.Fatalf("Error reading directory %s: %v\n", peerDirectory, err)
}
for _, file := range files {
if !file.IsDir() {
FILES_LIST = append(FILES_LIST, file.Name())
}
}


fmt.Println("Files available for sharing:", FILES_LIST)


// Register with the tracker server
registration(trackerAddress, udpPort)


// Start listening for UDP requests
go receiveUDPFileRequests(udpPort)


// Periodically request the list of peers
go func() {
for {
requestIpFromServer(trackerAddress)
time.Sleep(30 * time.Second) // Re-request neighbors every 5 seconds

}
}()


// File request logic with independent delay
go func() {
time.Sleep(10 * time.Second) // Wait 10 seconds before starting file requests


// Request file1.txt from Peer 1
if len(NEIGHBORS) > 0 {
fmt.Printf("Requesting 'file1.txt' from %s\n", NEIGHBORS[0])
fileRequest(NEIGHBORS[0], "file1.txt", IP_ADDRESS+":"+udpPort, "request123")
} else {
fmt.Println("No neighbors to request 'file1.txt'.")
}


// Wait 5 seconds before requesting the next file
time.Sleep(5 * time.Second)


// Request file3.txt from Peer 2
if len(NEIGHBORS) > 1 {
fmt.Printf("Requesting 'file3.txt' from %s\n", NEIGHBORS[1])
fileRequest(NEIGHBORS[1], "file3.txt", IP_ADDRESS+":"+udpPort, "request456")
} else {
fmt.Println("Not enough neighbors to request 'file3.txt'.")
}
}()


select {} // Keep the program running
}


