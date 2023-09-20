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

// package main

// import (
// 	"bytes"
// 	"crypto/sha1"
// 	"encoding/binary"
// 	"encoding/hex"
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"log"
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

// type Peers struct {
// 	Ip   net.IP
// 	Port uint64
// }

// const (
// 	Bitfield   = 5
// 	Interested = 2
// 	Unchoke    = 1
// 	Request    = 6
// 	Piece      = 7
// 	BLOCK      = 16 * 1024
// )

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

// func getRequest(jsonObject Torrent, buffer_ bytes.Buffer) (TrackerResponse, error) {
// 	infoHash := sha1.Sum(buffer_.Bytes())
// 	queryParams := url.Values{}
// 	queryParams.Add("info_hash", string(infoHash[:]))
// 	queryParams.Add("peer_id", "00112233445566778899")
// 	queryParams.Add("port", "6881")
// 	queryParams.Add("uploaded", "0")
// 	queryParams.Add("downloaded", "0")
// 	queryParams.Add("left", strconv.Itoa(int(jsonObject.Info.Length)))
// 	queryParams.Add("compact", "1")
// 	encodedParams := queryParams.Encode()
// 	fullURL := fmt.Sprintf("%s?%s", jsonObject.Announce, encodedParams)

// 	response, err := http.Get(fullURL)
// 	if err != nil {
// 		fmt.Println("Error:", err)
// 		return TrackerResponse{}, err
// 	}
// 	defer response.Body.Close()

// 	if response.StatusCode != http.StatusOK {
// 		fmt.Println("Request failed with status code:", response.StatusCode)
// 		return TrackerResponse{}, err
// 	}

// 	var trackerResponse TrackerResponse
// 	err = bencode.Unmarshal(response.Body, &trackerResponse)
// 	if err != nil {
// 		fmt.Println(err)
// 		return TrackerResponse{}, err
// 	}

// 	return trackerResponse, nil
// }

// func sendHandshake(conn net.Conn, peers string, buffer bytes.Buffer) string {
// 	infoHash := sha1.Sum(buffer.Bytes())
// 	peerID := []byte("00112233445566778899")

// 	handshake := new(bytes.Buffer)
// 	handshake.WriteByte(19)
// 	handshake.WriteString("BitTorrent protocol")
// 	handshake.Write([]byte{0, 0, 0, 0, 0, 0, 0, 0}) // 8 reserved bytes
// 	handshake.Write(infoHash[:])
// 	handshake.Write(peerID)

// 	_, err := conn.Write(handshake.Bytes())
// 	if err != nil {
// 		fmt.Println("Error sending handshake:", err)
// 		panic(err)
// 	}

// 	buf := make([]byte, 68)
// 	_, err = io.ReadFull(conn, buf)
// 	if err != nil {
// 		if err == io.EOF {
// 			fmt.Println("Peer closed the connection")
// 		} else {
// 			fmt.Println("Error reading response:", err)
// 		}
// 		panic(err)
// 	}

// 	receivedPeerID := buf[48:]
// 	log.Printf("Handshake completed : %s", hex.EncodeToString(receivedPeerID))
// 	return fmt.Sprintf("Peer ID: %s\n", hex.EncodeToString(receivedPeerID))
// }

// func get_peers(trackerResponse TrackerResponse) []Peers {
// 	peerSize := 6
// 	numPeers := len(trackerResponse.Peers) / peerSize
// 	peersArray := make([]Peers, numPeers)
// 	for i := 0; i < numPeers; i++ {
// 		start := i * 6
// 		end := start + 6
// 		peer := trackerResponse.Peers[start:end]
// 		ip := net.IP(peer[0:4])
// 		port := binary.BigEndian.Uint16([]byte(peer[4:6]))
// 		peersArray[i] = Peers{
// 			Ip:   ip,
// 			Port: uint64(port),
// 		}
// 	}
// 	log.Println(peersArray, "This is peersArray")
// 	return peersArray
// }

// func printPeers(peers []Peers) {
// 	for i := 0; i < len(peers); i++ {
// 		fmt.Printf("%s:%d\n", peers[i].Ip, peers[i].Port)
// 	}
// }

// func handlePeerMessages(conn net.Conn, messageID_ uint8) []byte {
// 	// fmt.Println("Handle peer message started ", messageID_)
// 	// for {
// 	buffer := make([]byte, 4)
// 	// _, err := io.ReadFull(conn, buffer)
// 	_, err := conn.Read(buffer)
// 	if (err) != nil {
// 		fmt.Println("Error reading message length:", err)
// 		conn.Close()
// 		panic(err)

// 	}
// 	recievedMessageID := make([]byte, 1)
// 	messageLength := binary.BigEndian.Uint32(buffer)
// 	// _, err = io.ReadFull(conn, messageID)
// 	_, err = conn.Read(recievedMessageID)
// 	if err != nil {
// 		fmt.Println("Error reading message ID:", err)
// 		conn.Close()
// 		panic(err)
// 	}
// 	var messageId uint8
// 	binary.Read(bytes.NewReader(recievedMessageID), binary.BigEndian, &messageId)
// 	if messageId != messageID_ {
// 		return nil
// 	}
// 	payload := make([]byte, messageLength-1)

// 	size, err := io.ReadFull(conn, payload)
// 	if err != nil {
// 		fmt.Println("Error reading message length:", err)
// 		conn.Close()
// 		panic(err)
// 	}

// 	log.Printf("Size: %d, Message_id: %d\n", size, messageID_)
// 	return payload
// 	// }
// }

// func createConnection(peer string) (net.Conn, error) {
// 	conn, err := net.Dial("tcp", peer)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return conn, nil
// }

// func closeALlConn(conns map[string]net.Conn) {
// 	for _, conn := range conns {
// 		conn.Close()
// 	}
// }

// func main() {
// 	if len(os.Args) < 2 {
// 		fmt.Println("Usage: ./your_bittorrent.sh <command>")
// 		return
// 	}

// 	command := os.Args[1]

// 	switch command {
// 	case "decode":
// 		if len(os.Args) < 3 {
// 			fmt.Println("Usage: ./your_bittorrent.sh decode <bencoded_value>")
// 			return
// 		}
// 		bencodedValue := os.Args[2]
// 		decoded, err := bencode.Decode(bytes.NewReader([]byte(bencodedValue)))
// 		if err != nil {
// 			fmt.Println(err)
// 			return
// 		}
// 		jsonOutput, _ := json.Marshal(decoded)
// 		fmt.Println(string(jsonOutput))

// 	case "info", "peers", "handshake":
// 		if len(os.Args) < 3 {
// 			fmt.Println("Usage: ./your_bittorrent.sh", command, "<torrent_file>")
// 			return
// 		}
// 		filename := os.Args[2]
// 		data, err := os.Open(filename)
// 		if err != nil {
// 			fmt.Println("Error reading file:", err)
// 			return
// 		}
// 		defer data.Close()

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
// 		switch command {
// 		case "info":
// 			chunkSize := 20
// 			chunks := splitString(jsonObject.Info.Pieces, chunkSize)
// 			fmt.Printf("Tracker URL: %s\n", jsonObject.Announce)
// 			fmt.Printf("Length: %d\n", jsonObject.Info.Length)
// 			fmt.Printf("Info Hash: %x\n", sha1.Sum(buffer_.Bytes()))
// 			fmt.Printf("Piece Length: %d\n", jsonObject.Info.PiecesLen)
// 			fmt.Printf("Piece Hashes:\n")
// 			for _, chunk := range chunks {
// 				fmt.Printf("%x\n", chunk)
// 			}

// 		case "peers":
// 			trackerResponse, err := getRequest(jsonObject, buffer_)
// 			if err != nil {
// 				fmt.Println(err)
// 				return
// 			}
// 			peers := get_peers(trackerResponse)
// 			printPeers(peers)

// 		case "handshake":
// 			if len(os.Args) < 4 {
// 				fmt.Println("Usage: ./your_bittorrent.sh handshake <torrent_file> <peer_ip:peer_port>")
// 				return
// 			}
// 			peers := os.Args[3]
// 			conn, err := createConnection(peers)
// 			if err != nil {
// 				return
// 			}
// 			fmt.Println(sendHandshake(conn, peers, buffer_))
// 			conn.Close()
// 		}
// 	case "download_piece":
// 		filename := os.Args[4]
// 		data, err := os.Open(filename)
// 		if err != nil {
// 			fmt.Println("Error reading file:", err)
// 			return
// 		}
// 		pieceIndex, err := strconv.Atoi(os.Args[5])
// 		if err != nil {
// 			return
// 		}
// 		defer data.Close()

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
// 		trackerResponse, err := getRequest(jsonObject, buffer_)
// 		if err != nil {
// 			fmt.Println(err)
// 			return
// 		}
// 		connections := map[string]net.Conn{}
// 		peers := get_peers(trackerResponse)
// 		peerObjVal := peers[0]
// 		peerStr := fmt.Sprintf("%s:%d", peerObjVal.Ip, peerObjVal.Port)
// 		connections[peerStr], err = createConnection(peerStr)
// 		defer closeALlConn(connections)
// 		if err != nil {
// 			fmt.Println(err, "Error while creating connection")
// 			return
// 		}
// 		sendHandshake(connections[peerStr], peerStr, buffer_)
// 		handlePeerMessages(connections[peerStr], Bitfield)
// 		// interestedMessage := []byte{0, 0, 0, 1, 2} // Message length (1 byte) + Message ID (1 byte) + Payload (empty)
// 		interestedMessage := make([]byte, 4+1+len([]byte{}))
// 		binary.BigEndian.PutUint32(interestedMessage[0:4], uint32(1+len([]byte{})))
// 		interestedMessage[4] = Interested
// 		copy(interestedMessage[5:], []byte{})
// 		connections[peerStr].Write(interestedMessage)
// 		handlePeerMessages(connections[peerStr], Unchoke)
// 		piecesHex := jsonObject.Info.Pieces
// 		pieces := make([]string, len(piecesHex)/20)
// 		for i := 0; i < len(piecesHex)/20; i++ {
// 			piece := piecesHex[i*20 : (i*20)+20]
// 			pieces[i] = piece
// 		}
// 		piecesHash := pieces[pieceIndex]

// 		log.Printf("This is piece hash: %x and piece id: %d\n", piecesHash, pieceIndex)

// 		count := 0
// 		for i := int64(0); i < int64(jsonObject.Info.PiecesLen); i = i + BLOCK {
// 			requestMessage := make([]byte, 12)
// 			binary.BigEndian.PutUint32(requestMessage[0:4], uint32(pieceIndex))
// 			binary.BigEndian.PutUint32(requestMessage[4:8], uint32(i))
// 			binary.BigEndian.PutUint32(requestMessage[8:], BLOCK)

// 			messageData := make([]byte, 4+1+len(requestMessage))
// 			binary.BigEndian.PutUint32(messageData[0:4], uint32(1+len(requestMessage)))
// 			messageData[4] = Request
// 			copy(messageData[5:], requestMessage)
// 			_, err = connections[peerStr].Write(messageData)
// 			if err != nil {
// 				fmt.Println("Error sending request message: ", err)
// 				return
// 			}
// 			count++
// 		}
// 		combinedBlockPiece := make([]byte, jsonObject.Info.PiecesLen)
// 		for i := int(0); i < int(count); i++ {
// 			// fmt.Println("This the piece number: ", Piece)
// 			data := handlePeerMessages(connections[peerStr], Piece)
// 			pieceInd := binary.BigEndian.Uint32(data[0:4])
// 			if pieceInd != uint32(pieceIndex) {
// 				fmt.Println(err)
// 				return
// 			}
// 			begin := binary.BigEndian.Uint32(data[4:8])
// 			blockData := data[8:]
// 			copy(combinedBlockPiece[begin:], blockData)
// 		}
// 		sum := sha1.Sum(combinedBlockPiece)
// 		// fmt.Println(string(sum[:]) == piecesHash, "this is hash")
// 		if string(sum[:]) == piecesHash {
// 			file_val := os.Args[3]
// 			// fmt.Println(file_val, "this is arg3")
// 			err := os.WriteFile(file_val, combinedBlockPiece, os.ModePerm)
// 			if err != nil {
// 				fmt.Println(err)
// 				return
// 			}
// 			fmt.Printf("Piece %d downloaded to %s.\n", pieceIndex, file_val)
// 		} else {
// 			panic("Not matched ")
// 		}
// 		connections[peerStr].Close()

// 	default:
// 		fmt.Println("Unknown command:", command)
// 		os.Exit(1)
// 	}
// }


package main

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
	bencode "github.com/jackpal/bencode-go"
)
/*
	References:
	- https://www.bittorrent.org/beps/bep_0003.html#metainfo-files
	- https://www.sohamkamani.com/golang/json/
	- https://github.com/veggiedefender/torrent-client/blob/master/torrentfile/torrentfile.go
*/
type bencodeInfo struct {
	Pieces      string `bencode:"pieces"`
	PieceLength uint64 `bencode:"piece length"`
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
}
type bencodeTorrent struct {
	Announce string      `bencode:"announce"`
	Info     bencodeInfo `bencode:"info"`
}
type bencodeTracker struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}
type Peer struct {
	IP   net.IP
	Port uint16
}
type TorrentFile struct {
	Filename    string
	Announce    string
	InfoHash    [20]byte
	PeerId      string
	Length      int
	PieceLength int
	TotalHashes int
	Hashes      [][20]byte
}
func (torrent *TorrentFile) TorrentInfo() *TorrentFile {
	file, err := os.Open(torrent.Filename)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	defer file.Close()
	bencodeObject := bencodeTorrent{}
	err = bencode.Unmarshal(file, &bencodeObject)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	var buf bytes.Buffer
	if err := bencode.Marshal(&buf, bencodeObject.Info); err != nil {
		fmt.Println(err)
		return nil
	}
	torrent.Announce = bencodeObject.Announce
	torrent.InfoHash = sha1.Sum(buf.Bytes())
	torrent.PeerId = "00112233445566778899"
	torrent.Length = bencodeObject.Info.Length
	const hashLength = 20
	pieces := []byte(bencodeObject.Info.Pieces)
	totalHashes := len(pieces) / hashLength
	hashes := make([][hashLength]byte, totalHashes)
	for i := 0; i < totalHashes; i++ {
		copy(hashes[i][:], pieces[i*hashLength:(i+1)*hashLength])
	}
	torrent.TotalHashes = totalHashes
	torrent.Hashes = hashes
	torrent.PieceLength = int(bencodeObject.Info.PieceLength)
	return torrent
}
func (torrent *TorrentFile) GetPeers() []Peer {
	base, err := url.Parse(torrent.Announce)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	const port = 6881
	params := url.Values{
		"info_hash":  []string{string(torrent.InfoHash[:])},
		"peer_id":    []string{string(torrent.PeerId[:])},
		"port":       []string{strconv.Itoa(int(port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(torrent.Length)},
	}
	base.RawQuery = params.Encode()
	http_client := &http.Client{Timeout: 15 * time.Second}
	r, err := http_client.Get(base.String())
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer r.Body.Close()
	trackerResponse := bencodeTracker{}
	err = bencode.Unmarshal(r.Body, &trackerResponse)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	const peerSize = 6
	numPeers := len(trackerResponse.Peers) / peerSize
	peers := make([]Peer, numPeers)
	for i := 0; i < numPeers; i++ {
		offset := i * peerSize
		peers[i].IP = net.IP(trackerResponse.Peers[offset : offset+4])
		peers[i].Port = binary.BigEndian.Uint16([]byte(trackerResponse.Peers[offset+4 : offset+6]))
	}
	return peers
}
func DoHandshake(connection net.Conn, torrent TorrentFile) {
	const PROTO_BANNER = "BitTorrent protocol"
	const PROTO_LENGTH byte = byte(len(PROTO_BANNER))
	const DEMO_PEER_ID = "00112233445566778899"
	handshake := new(bytes.Buffer)
	// serialize banner length
	err := binary.Write(handshake, binary.BigEndian, PROTO_LENGTH)
	if err != nil {
		log.Fatal(err)
		return
	}
	// serialize banner
	_, err = handshake.Write([]byte(PROTO_BANNER))
	if err != nil {
		log.Fatal(err)
		return
	}
	// serialize reserved bitfield
	_, err = handshake.Write(make([]byte, 8))
	if err != nil {
		log.Fatal(err)
		return
	}
	// send info hash
	_, err = handshake.Write(torrent.InfoHash[:])
	if err != nil {
		log.Fatal(err)
		return
	}
	// serialize peer id
	_, err = handshake.Write([]byte(DEMO_PEER_ID))
	if err != nil {
		log.Fatal(err)
		return
	}
	// send handshake
	_, err = connection.Write(handshake.Bytes())
	if err != nil {
		log.Fatal(err)
		return
	}
	// receive handshake
	recv_buffer := make([]byte, len(handshake.Bytes()))
	_, err = connection.Read(recv_buffer)
	if err != nil {
		log.Fatal(err)
		return
	}
}
func (peer *Peer) DoHandshake(torrent *TorrentFile) []byte {
	peer_addr := fmt.Sprintf("%s:%d", peer.IP, peer.Port)
	// establish TCP connection
	sock, err := net.Dial("tcp", peer_addr)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer sock.Close()
	const PROTO_BANNER = "BitTorrent protocol"
	const PROTO_LENGTH byte = byte(len(PROTO_BANNER))
	const DEMO_PEER_ID = "00112233445566778899"
	handshake := new(bytes.Buffer)
	// serialize banner length
	err = binary.Write(handshake, binary.BigEndian, PROTO_LENGTH)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	// serialize banner
	_, err = handshake.Write([]byte(PROTO_BANNER))
	if err != nil {
		log.Fatal(err)
		return nil
	}
	// serialize reserved bitfield
	_, err = handshake.Write(make([]byte, 8))
	if err != nil {
		log.Fatal(err)
		return nil
	}
	// send info hash
	_, err = handshake.Write(torrent.InfoHash[:])
	if err != nil {
		log.Fatal(err)
		return nil
	}
	// serialize peer id
	_, err = handshake.Write([]byte(DEMO_PEER_ID))
	if err != nil {
		log.Fatal(err)
		return nil
	}
	// send handshake
	_, err = sock.Write(handshake.Bytes())
	if err != nil {
		log.Fatal(err)
		return nil
	}
	// receive handshake
	recv_buffer := make([]byte, len(handshake.Bytes()))
	_, err = sock.Read(recv_buffer)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	return recv_buffer
}
func (peer *Peer) CreateConnection() net.Conn {
	// Connect to a TCP server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", peer.IP, peer.Port))
	if err != nil {
		log.Fatal(err)
		return nil
	}
	return conn
}
func CloseConnections(connections map[string]net.Conn) {
	for _, conn := range connections {
		defer conn.Close()
	}
}
// Example:
// - 5:hello -> hello
// - 10:hello12345 -> hello12345
func decodeBencode(bencodedString *bufio.Reader) (interface{}, error) {
	c, err := bencodedString.ReadByte()
	if err != nil {
		return nil, err
	}
	switch c {
	// integer
	case 'i':
		{
			iBuf, err := bencodedString.ReadBytes('e')
			if err != nil {
				return nil, err
			}
			iBuf = iBuf[:len(iBuf)-1]
			num, err := strconv.ParseInt(string(iBuf), 10, 64)
			if err != nil {
				return "", err
			}
			return num, nil
		}
	// list
	case 'l':
		{
			list := []interface{}{}
			for {
				c, err := bencodedString.ReadByte()
				if err != nil {
					return nil, err
				}
				if c == 'e' {
					return list, nil
				}
				bencodedString.UnreadByte()
				value, err2 := decodeBencode(bencodedString)
				if err2 != nil {
					return nil, err2
				}
				list = append(list, value)
			}
		}
	// dictionary
	case 'd':
		{
			dict := map[string]interface{}{}
			for {
				c, err := bencodedString.ReadByte()
				if err != nil {
					return nil, err
				}
				if c == 'e' {
					return dict, nil
				}
				bencodedString.UnreadByte()
				value, err2 := decodeBencode(bencodedString)
				if err2 != nil {
					return nil, err2
				}
				key, ok := value.(string)
				if !ok {
					return nil, errors.New("invalid key format")
				}
				value, err2 = decodeBencode(bencodedString)
				if err2 != nil {
					return nil, err2
				}
				dict[key] = value
			}
		}
	// string
	default:
		{
			bencodedString.UnreadByte()
			stringLengthBuffer, err := bencodedString.ReadBytes(':')
			if err != nil {
				return nil, err
			}
			stringLengthBuffer = stringLengthBuffer[:len(stringLengthBuffer)-1]
			stringLength, err := strconv.ParseInt(string(stringLengthBuffer), 10, 64)
			if err != nil {
				return nil, err
			}
			buf := make([]byte, stringLength)
			n, err2 := bencodedString.Read(buf)
			if n != int(stringLength) || err2 != nil {
				return nil, err2
			}
			return string(buf), nil
		}
	}
}
func WaitFor(connection net.Conn, expected_message_id uint8) []byte {
	log.Printf("[+] Connected: %s\n", connection.RemoteAddr())
	log.Printf("[!] Waiting for %d\n", expected_message_id)
	msg_len := make([]byte, 4)
	_, err := connection.Read(msg_len)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	log.Printf("[+] Received: %x\n", msg_len)
	messageLength := binary.BigEndian.Uint32(msg_len)
	log.Printf("[+] messageLength %v\n", messageLength)
	message_id_byte := make([]byte, 1)
	_, err = connection.Read(message_id_byte)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	var message_id uint8
	binary.Read(bytes.NewReader(message_id_byte), binary.BigEndian, &message_id)
	log.Printf("[!] Received: %x - Expected: %x\n", message_id, expected_message_id)
	if message_id != expected_message_id {
		return nil
	}
	// we already consumed 1 byte for message id
	payload := make([]byte, messageLength-1)
	size, err := io.ReadFull(connection, payload)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	log.Printf("Payload: %d, size = %d\n", len(payload), size)
	log.Printf("Message ID: %d\n", message_id)
	log.Printf("Return for MessageId: %d\n", message_id)
	return payload
}
func SendPayload(connection net.Conn, message_id uint8, payload []byte) {
	_, err := connection.Write(createPeerMessage(message_id, payload))
	if err != nil {
		log.Fatal(err)
		return
	}
}
func createPeerMessage(messageId uint8, payload []byte) []byte {
	// Peer messages consist of a message length prefix (4 bytes), message id (1 byte) and a payload (variable size).
	messageData := make([]byte, 4+1+len(payload))
	binary.BigEndian.PutUint32(messageData[0:4], uint32(1+len(payload)))
	messageData[4] = messageId
	copy(messageData[5:], payload)
	return messageData
}
func main() {
	log.SetFlags(0)
	log.SetOutput(os.Stderr)
	command := os.Args[1]
	if command == "decode" {
		bencodedValue := os.Args[2]
		decoded, err := decodeBencode(bufio.NewReader(strings.NewReader(bencodedValue)))
		if err != nil {
			fmt.Println(err)
			return
		}
		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
	} else if command == "info" {
		filename := os.Args[2]
		torrent := new(TorrentFile)
		torrent.Filename = filename
		torrent = torrent.TorrentInfo()
		fmt.Printf("Tracker URL: %s\n", torrent.Announce)
		fmt.Printf("Length: %d\n", torrent.Length)
		fmt.Printf("Info Hash: %x\n", torrent.InfoHash)
		fmt.Printf("Piece Length: %d\n", torrent.PieceLength)
		fmt.Printf("Piece Hashes:\n")
		for _, hash := range torrent.Hashes {
			fmt.Printf("%x\n", hash)
		}
	} else if command == "peers" {
		filename := os.Args[2]
		torrent := new(TorrentFile)
		torrent.Filename = filename
		torrent = torrent.TorrentInfo()
		base, err := url.Parse(torrent.Announce)
		if err != nil {
			fmt.Println(err)
			return
		for _, peer := range torrent.GetPeers() {
			fmt.Printf("%s:%d\n", peer.IP, peer.Port)
		}
		const port = 6881
	} else if command == "handshake" {
		filename := os.Args[2]
		params := url.Values{
			"info_hash":  []string{string(torrent.InfoHash[:])},
			"peer_id":    []string{string(torrent.PeerId[:])},
			"port":       []string{strconv.Itoa(int(port))},
			"uploaded":   []string{"0"},
			"downloaded": []string{"0"},
			"compact":    []string{"1"},
			"left":       []string{strconv.Itoa(torrent.Length)},
		}
		torrent := new(TorrentFile)
		torrent.Filename = filename
		torrent = torrent.TorrentInfo()
		base.RawQuery = params.Encode()
		peer_addr := os.Args[3]
		tmp := strings.Split(peer_addr, ":")
		http_client := &http.Client{Timeout: 15 * time.Second}
		r, err := http_client.Get(base.String())
		if err != nil {
			fmt.Println(err)
		if len(tmp) != 2 {
			log.Fatal("invalid peer address")
			return
		}
		defer r.Body.Close()
		trackerResponse := bencodeTracker{}
		err = bencode.Unmarshal(r.Body, &trackerResponse)
		peer_ip := net.ParseIP(tmp[0])
		peer_port, err := strconv.Atoi(tmp[1])
		if err != nil {
			fmt.Println(err)
			log.Fatal(err)
			return
		}
		const peerSize = 6
		numPeers := len(trackerResponse.Peers) / peerSize
		peers := make([]Peer, numPeers)
		for i := 0; i < numPeers; i++ {
			offset := i * peerSize
			peers[i].IP = net.IP(trackerResponse.Peers[offset : offset+4])
			peers[i].Port = binary.BigEndian.Uint16([]byte(trackerResponse.Peers[offset+4 : offset+6]))
		peer := Peer{
			IP:   peer_ip,
			Port: uint16(peer_port),
		}
		for _, peer := range peers {
			fmt.Printf("%s:%d\n", peer.IP, peer.Port)
		}
		fmt.Printf("IP: %s - Port: %d\n", peer.IP, peer.Port)
	} else if command == "handshake" {
		filename := os.Args[2]
		recv_buffer := peer.DoHandshake(torrent)
		fmt.Printf("Peer ID: %x\n", recv_buffer[48:])
	} else if command == "download_piece" {
		const MSG_UNCHOKE = 1
		const MSG_INTERESTED = 2
		const MSG_BITFIELD = 5
		const MSG_REQUEST = 6
		const MSG_PIECE = 7
		const BLOCK_SIZE = 16 * 1024
		outfile := os.Args[3]
		filename := os.Args[4]
		torrent := new(TorrentFile)
		torrent.Filename = filename
		torrent = torrent.TorrentInfo()
		peer := os.Args[3]
		tmp := strings.Split(peer, ":")
		peerIP := tmp[0]
		peerPort := tmp[1]
		/*
			All non-keepalive messages start with a single byte which gives their type.
		fmt.Printf("IP: %s - Port: %s\n", peerIP, peerPort)
			The possible values are:
		// establish TCP connection
		sock, err := net.Dial("tcp", peer)
				0 - choke
				1 - unchoke
				2 - interested
				3 - not interested
				4 - have
				5 - bitfield
				6 - request
				7 - piece
				8 - cancel
		*/
		pieceId, err := strconv.Atoi(os.Args[5])
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
			return
		}
		defer sock.Close()
		connections := make(map[string]net.Conn)
		peers := torrent.GetPeers()
		const PROTO_BANNER = "BitTorrent protocol"
		const PROTO_LENGTH byte = byte(len(PROTO_BANNER))
		const DEMO_PEER_ID = "00112233445566778899"
		peer := peers[0]
		peer_addr := fmt.Sprintf("%s:%d", peer.IP, peer.Port)
		handshake := new(bytes.Buffer)
		connections[peer_addr] = peer.CreateConnection()
		defer CloseConnections(connections)
		// serialize banner length
		err = binary.Write(handshake, binary.BigEndian, PROTO_LENGTH)
		if err != nil {
			log.Fatal(err)
			return
		}
		log.Printf("[!] Performing handshake...\n")
		DoHandshake(connections[peer_addr], *torrent)
		log.Printf("[+] Handshake completed...\n")
		// serialize banner
		_, err = handshake.Write([]byte(PROTO_BANNER))
		if err != nil {
			log.Fatal(err)
			return
		}
		/*
			Here are the peer messages you'll need to exchange once the handshake is complete:
		// serialize reserved bitfield
		_, err = handshake.Write(make([]byte, 8))
		if err != nil {
			log.Fatal(err)
			return
		}
				Wait for a bitfield message from the peer indicating which pieces it has
					The message id for this message is 5.
					You can ignore the payload for now, the tracker we use for this challenge ensures that all peers have all pieces available.
		*/
		// send info hash
		_, err = handshake.Write(torrent.InfoHash[:])
		if err != nil {
			log.Fatal(err)
			return
		log.Printf("[!] Waiting for MSG_BITFIELD...\n")
		_ = WaitFor(connections[peer_addr], MSG_BITFIELD)
		log.Printf("[+] MSG_BITFIELD received...\n")
		/*
			Send an interested message
				The message id for interested is 2.
				The payload for this message is empty.
		*/
		log.Printf("[!] Sending MSG_INTERESTED...\n")
		SendPayload(connections[peer_addr], MSG_INTERESTED, []byte{})
		log.Printf("[+] MSG_INTERESTED sent...\n")
		/*
			Wait until you receive an unchoke message back
				The message id for unchoke is 1.
				The payload for this message is empty.
		*/
		log.Printf("[!] Waiting for MSG_UNCHOKE...\n")
		_ = WaitFor(connections[peer_addr], MSG_UNCHOKE)
		log.Printf("[+] MSG_UNCHOKE received...\n")
		/*
			Break the piece into blocks of 16 kiB (16 * 1024 bytes) and send a request message for each block
				The message id for request is 6.
				The payload for this message consists of:
					index: the zero-based piece index
					begin: the zero-based byte offset within the piece
						This'll be 0 for the first block, 2^14 for the second block, 2*2^14 for the third block etc.
					length: set this to 2^14 (16 * 1024)
		*/
		pieceHash := torrent.Hashes[pieceId]
		log.Printf("PieceHash for id: %d --> %x\n", pieceId, pieceHash)
		count := 0
		for byteOffset := 0; byteOffset < int(torrent.PieceLength); byteOffset = byteOffset + BLOCK_SIZE {
			payload := make([]byte, 12)
			binary.BigEndian.PutUint32(payload[0:4], uint32(pieceId))
			binary.BigEndian.PutUint32(payload[4:8], uint32(byteOffset))
			binary.BigEndian.PutUint32(payload[8:], BLOCK_SIZE)
			SendPayload(connections[peer_addr], MSG_REQUEST, payload)
			count++
		}
		// serialize peer id
		_, err = handshake.Write([]byte(DEMO_PEER_ID))
		if err != nil {
			log.Fatal(err)
			return
		/*
				Wait for a piece message for each block you've requested
					The message id for piece is 7.
					The payload for this message consists of:
						index: the zero-based piece index
						begin: the zero-based byte offset within the piece
						block: the data for the piece, usually 2^14 bytes long
			After receiving blocks and combining them into pieces, you'll want to check the integrity of each piece by comparing it's hash with the piece hash value found in the torrent file.
		*/
		combinedBlockToPiece := make([]byte, torrent.PieceLength)
		for i := 0; i < count; i++ {
			data := WaitFor(connections[peer_addr], MSG_PIECE)
			index := binary.BigEndian.Uint32(data[0:4])
			if index != uint32(pieceId) {
				panic(fmt.Sprintf("something went wrong [expected: %d -- actual: %d]", pieceId, index))
			}
			begin := binary.BigEndian.Uint32(data[4:8])
			block := data[8:]
			copy(combinedBlockToPiece[begin:], block)
		}
		// send handshake
		_, err = sock.Write(handshake.Bytes())
		if err != nil {
			log.Fatal(err)
			return
		sum := sha1.Sum(combinedBlockToPiece)
		if sum != pieceHash {
			panic("unequal pieces")
		}
		recv_buffer := make([]byte, len(handshake.Bytes()))
		_, err = sock.Read(recv_buffer)
		err = os.WriteFile(outfile, combinedBlockToPiece, os.ModePerm)
		if err != nil {
			log.Fatal(err)
			return
		}
		fmt.Printf("Peer ID: %x\n", recv_buffer[48:])
		fmt.Printf("Piece %d downloaded to %s.\n", pieceId, outfile)
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}