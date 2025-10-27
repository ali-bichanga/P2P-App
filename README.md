# How to run
1. Start up the server by running the command go run server.go
2. Create a new directory in the directory that contains all your code files and name it files
3. Create 3 text files file1.txt, file2.tx and file3.txt
4. Input your ip address to the variable at the top of the file peer3.go.
5. Run the command go run peer3.go localhost:8000 8001 files

# Network Design and Architecture
This is a peer to peer network architecture implemented in Go. It includes communication protocols that allow peers to communicate and discover each other in the network as well as file sharing mechanisms.

There is a main tracker server which stores information about which peers are in the network. Peers register with this server when they join the network. They can also request a list of peers to connect with. The peers then connect with each other and share files within the network.

# Communication protocols and peer discovery methods
### TCP 
TCP is used for reliable communication. It is used for the following: 
- When peers first join the network they register with the tracker server. 
- When the peers request a list of peers from the tracker server.
- File transfer to ensure reliable data transfer.

### UDP
It is used to transmit file requests between peers within the network.


## Peer Discovery
Peers continuously request a list of active peers from the tracker server. The peers then connect to some of the active peers and begin making file requests. These requests are transmitted through the network.
When a file is found, the peer who has the file will initiate a TCP connection with the peer who requested the file and begin file transfer.

# File sharing, replication, and fault tolerance mechanisms
Peers request files from their neighbours through UDP. The request includes the filename, requester’s address and a request ID.
If the peer has the requested file, it opens a TCP connection with the peer who initiated the request and sends the file over TCP.

# Performance considerations and optimizations
Each file request has a request ID to ensure security through the network
and to avoid duplicate requests.
If a file is not found at a certain peer, the request is forwarded to the neighboring peers which increases the chances of locating the file.
The peers list has periodic updates to make sure the peers have the most recent information about their active neighbours.

# Future Improvements and limitations
The tracker server should only send a subset of its peers. 
Sometimes the peers do not receive file requests.


