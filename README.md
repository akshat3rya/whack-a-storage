# Distributed P2P Database

This project is a distributed file storage system implemented in Go. It combines a content-addressable storage (CAS) design with a peer-to-peer (P2P) networking layer to enable decentralized file storage and retrieval. Files are encrypted using AES-CTR mode before being transmitted over the network, ensuring end-to-end security.

## Overview

This project implements a distributed file storage system that:

- **Stores Files in a Distributed Manner:** Files are stored locally and replicated across nodes.
- **Uses Content-Addressable Storage:** Files are saved and retrieved by hashing the file key to generate a unique path.
- **Communicates Over a P2P Network:** Nodes discover and exchange file data using a custom TCP-based transport.
- **Secures Data with Encryption:** Data is encrypted with AES in CTR mode, with the initialization vector (IV) prepended to the ciphertext.

The system is modular, comprising a file server, storage engine, and networking layer that work together seamlessly.

## Features

- **Distributed File Operations:**
  - **Store:** Write files locally and broadcast their existence to peers.
  - **Retrieve:** Fetch files from peers if not found locally.
  - **Delete:** Remove files from local storage.
- **Content-Addressable Storage (CAS):**
  - Uses MD5 to hash keys into a directory structure.
  - Ensures deduplication and integrity of file data.
- **Encryption/Decryption:**
  - Uses AES-CTR mode.
  - Automatically prepends IV to the ciphertext to allow decryption.
- **Custom P2P Transport:**
  - TCP-based transport layer with a simple handshake and stream management.
  - Supports broadcasting messages (e.g., file store or get requests) to all peers.
- **Modular Code Base with Tests:**
  - Organized into distinct packages for storage, networking, and crypto.
  - Comprehensive unit tests for core functions.

## Architecture

The project is organized into four main modules: **FileServer**, **Store**, **P2P Networking**, and **Cryptography**.

### FileServer

The **FileServer** is the central component that orchestrates local storage, encryption/decryption, and network communications.

- **Responsibilities:**
  - **Local File Storage:** Interacts with the `Store` component to write, read, or delete files.
  - **P2P Communication:** Uses the custom TCP transport to broadcast and handle incoming messages.
  - **Message Handling:** Supports two primary messages:
    - `MessageStoreFile` – Notifies peers that a new file is stored.
    - `MessageGetFile` – Requests a file from peers if it is not found locally.
  - **Encryption & Decryption:** Encrypts file data before transmission and decrypts incoming data.

- **Key Methods:**
  - `Store(key string, r io.Reader) error`: Encrypts and writes a file to local storage, then broadcasts its availability.
  - `Get(key string) (io.Reader, error)`: Retrieves a file locally or fetches it from a peer if absent.
  - `broadcast(msg *Message) error`: Encodes and sends messages to all connected peers.
  - `loop()`: Listens for incoming P2P messages and dispatches them to appropriate handlers.

### Store (CAS)

The **Store** component implements content-addressable storage:

- **Key Functions:**
  - `CASPathTransformFunc(key string)`: Converts a file key into a hashed filename and nested directory path. For example, a key like `funnyphoto` is transformed into a hash, and then split into a nested directory structure.
  - `Write`, `Read`, `Delete`, and `Has`: Standard operations for managing files on disk.
- **Testing:**
  - Unit tests in `store_test.go` verify that keys are correctly transformed and that file operations work as expected.

### P2P Networking

The **P2P Networking** layer is encapsulated in the `p2p` package and handles peer discovery and message exchanges over TCP.

- **Transport:**
  - Uses TCP to establish connections between nodes.
  - Implements a simple handshake (`NOPHandshakeFunc`) for connection validation.
- **Message Protocol:**
  - Messages are encoded using Go’s `gob` package.
  - Two important message types:
    - `MessageStoreFile`: Contains the file ID, hashed key, and file size.
    - `MessageGetFile`: Requests a file based on file ID and hashed key.
- **Stream Management:**
  - Distinguishes between normal messages and streaming data by using a special byte flag.
  - Supports reading file streams and ensuring proper closure once data is transferred.

### Cryptography

Encryption and decryption are handled with AES in CTR mode.

- **Functions:**
  - `copyEncrypt(key []byte, src io.Reader, dst io.Writer) (int, error)`: Encrypts data from the source stream, prepends the IV, and writes to the destination.
  - `copyDecrypt(key []byte, src io.Reader, dst io.Writer) (int, error)`: Reads the IV from the source stream and decrypts the subsequent data.
- **Utility Functions:**
  - `generateID()`: Creates a unique identifier for each file server instance.
  - `hashKey(key string)`: Generates an MD5 hash for a given key, used to ensure consistency in file storage.
  - `newEncryptionKey()`: Generates a random key for encryption.

## Getting Started

### Prerequisites

- Go (version 1.18 or later)
- Git

### Installation

1. **Clone the Repository:**

    ```bash
    git clone https://github.com/akshat3rya/whack-a-storage

    cd whack-a-storage
    ```

2. **Build the Project:**

    ```bash
    go build -o file-server .
    ```

### Running the File Servers

The `main.go` file demonstrates how to start multiple file server nodes on different ports. For example, you can launch three nodes on ports 3000, 7000, and 5000:

```bash
# In separate terminal windows, run:
./file-server :3000
./file-server :7000
./file-server :5000
```

The nodes will bootstrap the network by connecting to each other using the addresses provided.

### Usage
### Storing a File

The client calls the Store method on the FileServer, providing a file key and an io.Reader for the file data.
The file is written locally (after being encrypted) and then a broadcast message is sent so that other nodes can replicate the file.

Example 
```
package main

import (
    "bytes"
    "log"
)

func exampleStore(fileServer *FileServer) {
    fileKey := "example_image.png"
    fileData := bytes.NewReader([]byte("This is the file content"))
    
    // Store the file locally and broadcast to peers
    if err := fileServer.Store(fileKey, fileData); err != nil {
        log.Fatalf("Failed to store file: %v", err)
    }
    
    log.Println("File stored successfully!")
}


```
### Retrieving a File
The client calls the Get method with the file key.
If the file is available locally, it is returned immediately.
Otherwise, a broadcast is sent requesting the file. The first peer that has the file streams back the encrypted content, which is then decrypted and stored locally.

``` env
package main

import (
    "fmt"
    "io/ioutil"
    "log"
)

func exampleGet(fileServer *FileServer) {
    fileKey := "example_image.png"
    
    // Retrieve the file
    reader, err := fileServer.Get(fileKey)
    if err != nil {
        log.Fatalf("Failed to get file: %v", err)
    }
    
    data, err := ioutil.ReadAll(reader)
    if err != nil {
        log.Fatalf("Error reading file: %v", err)
    }
    
    fmt.Println("Retrieved file content:", string(data))
}

```


### Testing

The project includes unit tests for key components:

- Storage Tests: store_test.go verifies that key transformation and file operations work correctly.
- Cryptography Tests: crypto_test.go ensures encryption and decryption work as expected.
- P2P Tests: Located under the p2p folder (e.g., tcp_transport_test.go) to validate transport functionality.


