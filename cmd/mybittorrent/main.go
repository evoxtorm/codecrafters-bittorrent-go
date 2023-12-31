// package main

// import (
// 	"bytes"
// 	"crypto/sha1"
// 	"encoding/binary"
// 	"encoding/hex"
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"net"
// 	"net/http"
// 	"net/url"
// 	"os"
// 	"strconv"

// 	bencode "github.com/jackpal/bencode-go"
// )

// type Torrent struct {
// 	Announce string `bencode:"announce"`
// 	Info     Info   `bencode:"info"`
// }

// type Info struct {
// 	Length    int64  `bencode:"length"`
// 	Name      string `bencode:"name"`
// 	PiecesLen int64  `bencode:"piece length"`
// 	Pieces    string `bencode:"pieces"`
// }

// type TrackerResponse struct {
// 	Interval int64  `bencode:"interval"`
// 	Peers    string `bencode:"peers"`
// }

// func splitString(input string, chunkSize int) []string {
// 	var chunks []string

// 	for i := 0; i < len(input); i += chunkSize {
// 		end := i + chunkSize
// 		if end > len(input) {
// 			end = len(input)
// 		}
// 		chunk := input[i:end]
// 		chunks = append(chunks, chunk)
// 	}

// 	return chunks
// }

// func get_request(jsonObject Torrent, buffer_ bytes.Buffer) (bool, error) {
// 	url_ := jsonObject.Announce
// 	info_hash := sha1.Sum(buffer_.Bytes())
// 	queryParams := url.Values{}
// 	queryParams.Add("info_hash", string(info_hash[:]))
// 	queryParams.Add("peer_id", "00112233445566778899")
// 	queryParams.Add("port", "6881")
// 	queryParams.Add("uploaded", "0")
// 	queryParams.Add("downloaded", "0")
// 	queryParams.Add("left", strconv.Itoa(int(jsonObject.Info.Length)))
// 	queryParams.Add("compact", "1")
// 	encodedParams := queryParams.Encode()
// 	fullURL := fmt.Sprintf("%s?%s", url_, encodedParams)

// 	response, err := http.Get(fullURL)
// 	if err != nil {
// 		fmt.Println("Error:", err)
// 		return false, err
// 	}
// 	defer response.Body.Close()

// 	if response.StatusCode != http.StatusOK {
// 		fmt.Println("Request failed with status code:", response.StatusCode)
// 		return false, err
// 	}

// 	var trackerResponse TrackerResponse
// 	err = bencode.Unmarshal(response.Body, &trackerResponse)
// 	if err != nil {
// 		fmt.Println(err)
// 		return false, err
// 	}
// 	peerSize := 6
// 	numPeers := len(trackerResponse.Peers) / peerSize
// 	for i := 0; i < numPeers; i++ {
// 		start := i * 6
// 		end := start + 6
// 		peer := trackerResponse.Peers[start:end]
// 		ip := net.IP(peer[0:4])
// 		port := binary.BigEndian.Uint16([]byte(peer[4:6]))
// 		fmt.Printf("%s:%d\n", ip, port)
// 	}
// 	return true, nil
// }

// func sendHandshake(peers string, buffer bytes.Buffer) {
// 	conn, err := net.Dial("tcp", peers)
// 	if err != nil {
// 		fmt.Printf("Error while making connection: %s\n", err)
// 		return
// 	}
// 	defer conn.Close()

// 	// Prepare infoHash and peerID
// 	infoHash := sha1.Sum(buffer.Bytes())
// 	peerID := []byte("00112233445566778899")

// 	// Build the handshake message
// 	handshake := new(bytes.Buffer)
// 	handshake.WriteByte(19)
// 	handshake.WriteString("BitTorrent protocol")
// 	handshake.Write([]byte{0, 0, 0, 0, 0, 0, 0, 0}) // 8 reserved bytes
// 	handshake.Write(infoHash[:])
// 	handshake.Write(peerID)

// 	// Send the handshake message over the connection
// 	_, err = conn.Write(handshake.Bytes())
// 	if err != nil {
// 		fmt.Println("Error sending handshake:", err)
// 		return
// 	}

// 	// Read and process the response handshake
// 	buf := make([]byte, 68)
// 	_, err = io.ReadFull(conn, buf)
// 	if err != nil {
// 		if err == io.EOF {
// 			fmt.Println("Peer closed the connection")
// 		} else {
// 			fmt.Println("Error reading response:", err)
// 		}
// 		return
// 	}

// 	// Extract and print the peer id from the response handshake
// 	receivedPeerID := buf[48:]
// 	fmt.Printf("Peer ID: %s\n", hex.EncodeToString(receivedPeerID))
// }

// func main() {
// 	command := os.Args[1]
// 	if command == "decode" {
// 		bencodedValue := os.Args[2]
// 		decoded, err := bencode.Decode(bytes.NewReader([]byte(bencodedValue)))
// 		if err != nil {
// 			fmt.Println(err)
// 			return
// 		}
// 		jsonOutput, _ := json.Marshal(decoded)
// 		fmt.Println(string(jsonOutput))
// 	} else if command == "info" {
// 		filename := os.Args[2]
// 		data, err := os.Open(filename)
// 		if err != nil {
// 			fmt.Println("Error reading file:", err)
// 			return
// 		}
// 		defer data.Close()
// 		if err != nil {
// 			fmt.Println(err)
// 			return
// 		}
// 		jsonObject := Torrent{}
// 		err = bencode.Unmarshal(data, &jsonObject)
// 		if err != nil {
// 			fmt.Println(err)
// 			return
// 		}

// 		var buffer_ bytes.Buffer

// 		if err := bencode.Marshal(&buffer_, jsonObject.Info); err != nil {
// 			return
// 		}
// 		chunkSize := 20

// 		chunks := splitString(jsonObject.Info.Pieces, chunkSize)
// 		fmt.Printf("Tracker URL: %s\n", jsonObject.Announce)
// 		fmt.Printf("Length: %d\n", jsonObject.Info.Length)
// 		fmt.Printf("Info Hash: %x\n", sha1.Sum(buffer_.Bytes()))
// 		fmt.Printf("Piece Length: %d\n", jsonObject.Info.PiecesLen)
// 		fmt.Printf("Piece Hashes:\n")
// 		for _, chunk := range chunks {
// 			if command == "info" {
// 				fmt.Printf("%x\n", chunk)
// 			}
// 		}

// 	} else if command == "peers" {
// 		filename := os.Args[2]
// 		data, err := os.Open(filename)
// 		if err != nil {
// 			fmt.Println("Error reading file:", err)
// 			return
// 		}
// 		defer data.Close()
// 		if err != nil {
// 			fmt.Println(err)
// 			return
// 		}
// 		jsonObject := Torrent{}
// 		err = bencode.Unmarshal(data, &jsonObject)
// 		if err != nil {
// 			fmt.Println(err)
// 			return
// 		}

// 		var buffer_ bytes.Buffer

// 		if err := bencode.Marshal(&buffer_, jsonObject.Info); err != nil {
// 			return
// 		}
// 		get_request(jsonObject, buffer_)
// 	} else if command == "handshake" {
// 		filename := os.Args[2]
// 		data, err := os.Open(filename)
// 		if err != nil {
// 			fmt.Println("Error reading file:", err)
// 			return
// 		}
// 		defer data.Close()
// 		if err != nil {
// 			fmt.Println(err)
// 			return
// 		}
// 		jsonObject := Torrent{}
// 		err = bencode.Unmarshal(data, &jsonObject)
// 		if err != nil {
// 			fmt.Println(err)
// 			return
// 		}

// 		var buffer_ bytes.Buffer

// 		if err := bencode.Marshal(&buffer_, jsonObject.Info); err != nil {
// 			return
// 		}
// 		peers := os.Args[3]
// 		// val := strings.Split(peers, ":")
// 		// ip := val[0]
// 		// portt := val[1]
// 		sendHandshake(peers, buffer_)
// 		// conn.Close()

// 	} else {
// 		fmt.Println("Unknown command: " + command)
// 		os.Exit(1)
// 	}
// }

package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"

	bencode "github.com/jackpal/bencode-go"
)

type Torrent struct {
	Announce string `bencode:"announce"`
	Info     Info   `bencode:"info"`
}

type Info struct {
	Length    int64  `bencode:"length"`
	Name      string `bencode:"name"`
	PiecesLen int64  `bencode:"piece length"`
	Pieces    string `bencode:"pieces"`
}

type TrackerResponse struct {
	Interval int64  `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

type Peers struct {
	Ip   net.IP
	Port uint64
}

const (
	Bitfield   = 5
	Interested = 2
	Unchoke    = 1
	// Request    = 6
	Piece = 7
	BLOCK = 16 * 1024
)

func splitString(input string, chunkSize int) []string {
	var chunks []string

	for i := 0; i < len(input); i += chunkSize {
		end := i + chunkSize
		if end > len(input) {
			end = len(input)
		}
		chunk := input[i:end]
		chunks = append(chunks, chunk)
	}

	return chunks
}

func getRequest(jsonObject Torrent, buffer_ bytes.Buffer) (TrackerResponse, error) {
	infoHash := sha1.Sum(buffer_.Bytes())
	queryParams := url.Values{}
	queryParams.Add("info_hash", string(infoHash[:]))
	queryParams.Add("peer_id", "00112233445566778899")
	queryParams.Add("port", "6881")
	queryParams.Add("uploaded", "0")
	queryParams.Add("downloaded", "0")
	queryParams.Add("left", strconv.Itoa(int(jsonObject.Info.Length)))
	queryParams.Add("compact", "1")
	encodedParams := queryParams.Encode()
	fullURL := fmt.Sprintf("%s?%s", jsonObject.Announce, encodedParams)

	response, err := http.Get(fullURL)
	if err != nil {
		fmt.Println("Error:", err)
		return TrackerResponse{}, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		fmt.Println("Request failed with status code:", response.StatusCode)
		return TrackerResponse{}, err
	}

	var trackerResponse TrackerResponse
	err = bencode.Unmarshal(response.Body, &trackerResponse)
	if err != nil {
		fmt.Println(err)
		return TrackerResponse{}, err
	}

	return trackerResponse, nil
}

func sendHandshake(conn net.Conn, peers string, buffer bytes.Buffer) string {
	infoHash := sha1.Sum(buffer.Bytes())
	peerID := []byte("00112233445566778899")

	handshake := new(bytes.Buffer)
	handshake.WriteByte(19)
	handshake.WriteString("BitTorrent protocol")
	handshake.Write([]byte{0, 0, 0, 0, 0, 0, 0, 0}) // 8 reserved bytes
	handshake.Write(infoHash[:])
	handshake.Write(peerID)

	_, err := conn.Write(handshake.Bytes())
	if err != nil {
		fmt.Println("Error sending handshake:", err)
		panic(err)
	}

	buf := make([]byte, 68)
	_, err = io.ReadFull(conn, buf)
	if err != nil {
		if err == io.EOF {
			fmt.Println("Peer closed the connection")
		} else {
			fmt.Println("Error reading response:", err)
		}
		panic(err)
	}

	receivedPeerID := buf[48:]
	log.Printf("Handshake completed : %s", hex.EncodeToString(receivedPeerID))
	return fmt.Sprintf("Peer ID: %s\n", hex.EncodeToString(receivedPeerID))
}

func get_peers(trackerResponse TrackerResponse) []Peers {
	peerSize := 6
	numPeers := len(trackerResponse.Peers) / peerSize
	peersArray := make([]Peers, numPeers)
	for i := 0; i < numPeers; i++ {
		start := i * 6
		end := start + 6
		peer := trackerResponse.Peers[start:end]
		ip := net.IP(peer[0:4])
		port := binary.BigEndian.Uint16([]byte(peer[4:6]))
		peersArray[i] = Peers{
			Ip:   ip,
			Port: uint64(port),
		}
	}
	log.Println(peersArray, "This is peersArray")
	return peersArray
}

func printPeers(peers []Peers) {
	for i := 0; i < len(peers); i++ {
		fmt.Printf("%s:%d\n", peers[i].Ip, peers[i].Port)
	}
}

func handlePeerMessages(conn net.Conn, messageID_ uint8) ([]byte, error) {
	// fmt.Println("Handle peer message started ", messageID_)
	// for {

	var messageLength uint32
	var messageID byte
	if err := binary.Read(conn, binary.BigEndian, &messageLength); err != nil {
		return nil, fmt.Errorf("error while reading message length: %s", err.Error())
	}
	if err := binary.Read(conn, binary.BigEndian, &messageID); err != nil {
		return nil, fmt.Errorf("error while reading message ID: %s", err.Error())
	}
	if messageID_ != messageID {
		return nil, fmt.Errorf("unexpected message ID: (actual=%d, expected=%d)", messageID, messageID_)
	}
	log.Printf("received message %d\n", messageID_)
	if messageLength > 1 {
		log.Printf("message %d has attached payload of size %d\n", messageID_, messageLength-1)
		payload := make([]byte, messageLength-1)
		if _, err := io.ReadAtLeast(conn, payload, len(payload)); err != nil {
			return nil, fmt.Errorf("error while reading payload: %s", err.Error())
		}
		return payload, nil
	}
	return nil, nil
}

func createConnection(peer string) (net.Conn, error) {
	conn, err := net.Dial("tcp", peer)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func closeALlConn(conns map[string]net.Conn) {
	for _, conn := range conns {
		conn.Close()
	}
}

func sendPeerMessage(connection net.Conn, messageId uint8, payload []byte) {
	messageData := make([]byte, 4+1+len(payload))
	binary.BigEndian.PutUint32(messageData[0:4], uint32(1+len(payload)))
	messageData[4] = messageId
	copy(messageData[5:], payload)
	_, err := connection.Write(messageData)
	if err != nil {
		log.Fatal(err)
		return
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./your_bittorrent.sh <command>")
		return
	}

	command := os.Args[1]

	switch command {
	case "decode":
		if len(os.Args) < 3 {
			fmt.Println("Usage: ./your_bittorrent.sh decode <bencoded_value>")
			return
		}
		bencodedValue := os.Args[2]
		decoded, err := bencode.Decode(bytes.NewReader([]byte(bencodedValue)))
		if err != nil {
			fmt.Println(err)
			return
		}
		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))

	case "info", "peers", "handshake":
		if len(os.Args) < 3 {
			fmt.Println("Usage: ./your_bittorrent.sh", command, "<torrent_file>")
			return
		}
		filename := os.Args[2]
		data, err := os.Open(filename)
		if err != nil {
			fmt.Println("Error reading file:", err)
			return
		}
		defer data.Close()

		jsonObject := Torrent{}
		err = bencode.Unmarshal(data, &jsonObject)
		if err != nil {
			fmt.Println(err)
			return
		}

		var buffer_ bytes.Buffer

		if err := bencode.Marshal(&buffer_, jsonObject.Info); err != nil {
			return
		}
		switch command {
		case "info":
			chunkSize := 20
			chunks := splitString(jsonObject.Info.Pieces, chunkSize)
			fmt.Printf("Tracker URL: %s\n", jsonObject.Announce)
			fmt.Printf("Length: %d\n", jsonObject.Info.Length)
			fmt.Printf("Info Hash: %x\n", sha1.Sum(buffer_.Bytes()))
			fmt.Printf("Piece Length: %d\n", jsonObject.Info.PiecesLen)
			fmt.Printf("Piece Hashes:\n")
			for _, chunk := range chunks {
				fmt.Printf("%x\n", chunk)
			}

		case "peers":
			trackerResponse, err := getRequest(jsonObject, buffer_)
			if err != nil {
				fmt.Println(err)
				return
			}
			peers := get_peers(trackerResponse)
			printPeers(peers)

		case "handshake":
			if len(os.Args) < 4 {
				fmt.Println("Usage: ./your_bittorrent.sh handshake <torrent_file> <peer_ip:peer_port>")
				return
			}
			peers := os.Args[3]
			conn, err := createConnection(peers)
			if err != nil {
				return
			}
			fmt.Println(sendHandshake(conn, peers, buffer_))
			conn.Close()
		}
	case "download_piece":
		filename := os.Args[4]
		data, err := os.Open(filename)
		if err != nil {
			fmt.Println("Error reading file:", err)
			return
		}
		pieceIndex, err := strconv.Atoi(os.Args[5])
		if err != nil {
			return
		}
		defer data.Close()

		jsonObject := Torrent{}
		err = bencode.Unmarshal(data, &jsonObject)
		if err != nil {
			fmt.Println(err)
			return
		}

		var buffer_ bytes.Buffer

		if err := bencode.Marshal(&buffer_, jsonObject.Info); err != nil {
			return
		}
		trackerResponse, err := getRequest(jsonObject, buffer_)
		if err != nil {
			fmt.Println(err)
			return
		}
		connections := map[string]net.Conn{}
		peers := get_peers(trackerResponse)
		peerObjVal := peers[0]
		peerStr := fmt.Sprintf("%s:%d", peerObjVal.Ip, peerObjVal.Port)
		connections[peerStr], err = createConnection(peerStr)
		defer closeALlConn(connections)
		if err != nil {
			fmt.Println(err, "Error while creating connection")
			return
		}
		sendHandshake(connections[peerStr], peerStr, buffer_)
		handlePeerMessages(connections[peerStr], Bitfield)
		// interestedMessage := []byte{0, 0, 0, 1, 2} // Message length (1 byte) + Message ID (1 byte) + Payload (empty)
		interestedMessage := make([]byte, 4+1+len([]byte{}))
		binary.BigEndian.PutUint32(interestedMessage[0:4], uint32(1+len([]byte{})))
		interestedMessage[4] = Interested
		copy(interestedMessage[5:], []byte{})
		connections[peerStr].Write(interestedMessage)
		handlePeerMessages(connections[peerStr], Unchoke)
		piecesHex := jsonObject.Info.Pieces
		pieces := make([]string, len(piecesHex)/20)
		for i := 0; i < len(piecesHex)/20; i++ {
			piece := piecesHex[i*20 : (i*20)+20]
			pieces[i] = piece
		}
		piecesHash := pieces[pieceIndex]

		log.Printf("This is piece hash: %x and piece id: %d\n", piecesHash, pieceIndex)
		if pieceIndex > len(jsonObject.Info.Pieces) {
			panic(fmt.Errorf("this torrent only has %d pieces but %d-th requested", len(jsonObject.Info.Pieces)-1, pieceIndex))
		}
		pieceLength := jsonObject.Info.PiecesLen
		if pieceIndex >= int(jsonObject.Info.Length/jsonObject.Info.PiecesLen) {
			pieceLength = jsonObject.Info.Length - (jsonObject.Info.PiecesLen * int64(pieceIndex))
		}
		lastBlockSize := pieceLength % BLOCK
		numBlocks := (pieceLength - lastBlockSize) / BLOCK
		log.Printf("*******Piece length : %d and lastBlockSize: %d and numblocks:  %d*****\n", pieceLength, lastBlockSize, numBlocks)
		if lastBlockSize > 0 {
			numBlocks++
		} else {
			log.Printf("piece %d has size of %d and is aligned with blocksize of %d\n", pieceIndex, jsonObject.Info.PiecesLen, BLOCK)
		}
		combinedBlockPiece := make([]byte, pieceLength)
		for i := int64(0); i < pieceLength; i += BLOCK {
			length := BLOCK
			if i+BLOCK > pieceLength {
				log.Printf("reached last block, changing size to %d\n", lastBlockSize)
				length = int(pieceLength) - int(i)
			}
			requestMessage := make([]byte, 12)
			binary.BigEndian.PutUint32(requestMessage[0:4], uint32(pieceIndex))
			binary.BigEndian.PutUint32(requestMessage[4:8], uint32(i))
			binary.BigEndian.PutUint32(requestMessage[8:], uint32(length))

			messageData := make([]byte, 4+1+len(requestMessage))
			binary.BigEndian.PutUint32(messageData[0:4], uint32(1+len(requestMessage)))
			messageData[4] = byte(6)
			copy(messageData[5:], requestMessage)
			_, err = connections[peerStr].Write(messageData)
			if err != nil {
				fmt.Println("Error sending request message: ", err)
				return
			}
			// sendPeerMessage(connections[peerStr], 6, requestMessage)
			data, err := handlePeerMessages(connections[peerStr], Piece)
			if err != nil {
				panic(err)
			}
			_ = binary.BigEndian.Uint32(data[0:4])
			begin := binary.BigEndian.Uint32(data[4:8])
			blockData := data[8:]
			copy(combinedBlockPiece[begin:], blockData)

		}
		h := sha1.Sum(combinedBlockPiece)
		log.Printf("Calculated hash: %s, piece hash : %s\n", h, piecesHash)
		if string(h[:]) == piecesHash {
			file_val := os.Args[3]
			err := os.WriteFile(file_val, combinedBlockPiece, os.ModePerm)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Printf("Piece %d downloaded to %s.\n", pieceIndex, file_val)
		} else {
			panic("Not matched ")
		}
		connections[peerStr].Close()

	default:
		fmt.Println("Unknown command:", command)
		os.Exit(1)
	}
}
