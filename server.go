package main


import (
"fmt"
"net"
"strings"
"sync"
)


var (
activePeers = make(map[string]bool)
mutex sync.Mutex
)


func handlePeerConnection(conn net.Conn) {
defer conn.Close()


buf := make([]byte, 1024)
n, err := conn.Read(buf)
if err != nil {
fmt.Printf("Error reading from peer: %v\n", err)
return
}


request := strings.TrimSpace(string(buf[:n]))
fmt.Printf("Received request: %s\n", request)


if strings.HasPrefix(request, "REGISTER") {
parts := strings.SplitN(request, "-", 2)
if len(parts) == 2 {
address := strings.TrimSpace(parts[1])
registerPeer(address, conn)
}
} else if request == "REQUEST_PEERS" {
sendActivePeers(conn)
}
}


func registerPeer(address string, conn net.Conn) {
mutex.Lock()
defer mutex.Unlock()


activePeers[address] = true
fmt.Printf("Peer registered: %s\n", address)


conn.Write([]byte("REGISTERED"))
}


func sendActivePeers(conn net.Conn) {
mutex.Lock()
defer mutex.Unlock()


var peerList []string
for peer := range activePeers {
peerList = append(peerList, peer)
}


peerData := strings.Join(peerList, ",")
conn.Write([]byte(peerData))
}


func main() {
listener, err := net.Listen("tcp", ":8000")
if err != nil {
fmt.Printf("Error starting server: %v\n", err)
return
}
defer listener.Close()
fmt.Println("Tracker server is running on port 8000")


for {
conn, err := listener.Accept()
if err != nil {
fmt.Printf("Error accepting connection: %v\n", err)
continue
}


go handlePeerConnection(conn)
}
}
