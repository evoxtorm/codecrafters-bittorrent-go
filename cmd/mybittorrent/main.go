package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
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

// func decodeBencodeNew(r *bufio.Reader) (interface{}, error) {
// 	data, err := bencode.Decode(r)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return data, nil
// }

func decodeBencode(bencodedString string) (interface{}, error) {
	r_string, err := regexp.Compile(`^\d+\:(.*)`)
	if err != nil {
		return "", fmt.Errorf("only strings are supported at the moment")
	}
	if r_string.MatchString(bencodedString) {
		val := r_string.FindStringSubmatch(bencodedString)
		return val[1], nil
	}
	r_int, err := regexp.Compile(`^i(\-?\d+)e$`)
	if err != nil {
		return "", fmt.Errorf("only strings are supported at the moment")
	}
	if r_int.MatchString(bencodedString) {
		val := r_int.FindStringSubmatch(bencodedString)
		st_int, err := strconv.ParseInt(val[1], 10, 64)
		if err != nil {
			return "", fmt.Errorf("error while converting string to int")
		}
		return st_int, nil
	}
	r_list, err := regexp.Compile(`^l\d+\:(.*)i(\-?\d+)e.*`)
	if err != nil {
		return "", fmt.Errorf("only strings are supported at the moment")
	}
	if r_list.MatchString(bencodedString) {
		vals := r_list.FindStringSubmatch(bencodedString)
		s := make([]interface{}, 0, 2)
		for i := 1; i < len(vals); i++ {
			if i == 2 {
				st_int, err := strconv.ParseInt(vals[i], 10, 64)
				if err != nil {
					return "", fmt.Errorf("error while converting string to int")
				}
				s = append(s, st_int)
				continue
			}
			s = append(s, vals[i])
		}
		return s, nil
	}
	r_dict, err := regexp.Compile(`^d\d+\:(.*)\d+\:(.*)\d+\:(.*)i(\-?\d+)e.*`)
	if err != nil {
		return "", fmt.Errorf("error while converting string to int")
	}
	if r_dict.MatchString(bencodedString) {
		vals := r_dict.FindStringSubmatch(bencodedString)
		m := map[string]interface{}{}
		i := 1
		for i < len(vals) {
			if i == 3 {
				st_int, err := strconv.ParseInt(vals[i+1], 10, 64)
				if err != nil {
					continue
				}
				m[vals[i]] = st_int
				i += 2
				continue
			}
			m[vals[i]] = vals[i+1]
			i += 2
		}
		return m, nil
	}
	return "", nil
}

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

func get_request(jsonObject Torrent, buffer_ bytes.Buffer) (bool, error) {
	url_ := jsonObject.Announce
	info_hash := sha1.Sum(buffer_.Bytes())
	queryParams := url.Values{}
	queryParams.Add("info_hash", string(info_hash[:]))
	queryParams.Add("peer_id", "00112233445566778899")
	queryParams.Add("port", "6881")
	queryParams.Add("uploaded", "0")
	queryParams.Add("downloaded", "0")
	queryParams.Add("left", strconv.Itoa(int(jsonObject.Info.Length)))
	queryParams.Add("compact", "1")
	encodedParams := queryParams.Encode()
	fullURL := fmt.Sprintf("%s?%s", url_, encodedParams)

	response, err := http.Get(fullURL)
	if err != nil {
		fmt.Println("Error:", err)
		return false, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		fmt.Println("Request failed with status code:", response.StatusCode)
		return false, err
	}

	var trackerResponse TrackerResponse
	err = bencode.Unmarshal(response.Body, &trackerResponse)
	if err != nil {
		fmt.Println(err)
		return false, err
	}
	peerSize := 6
	numPeers := len(trackerResponse.Peers) / peerSize
	for i := 0; i < numPeers; i++ {
		start := i * 6
		end := start + 6
		peer := trackerResponse.Peers[start:end]
		ip := net.IP(peer[0:4])
		port := binary.BigEndian.Uint16([]byte(peer[4:6]))
		fmt.Printf("%s:%d\n", ip, port)
	}
	return true, nil
}

func main() {
	command := os.Args[1]
	if command == "decode" {
		bencodedValue := os.Args[2]
		decoded, err := decodeBencode(bencodedValue)
		if err != nil {
			fmt.Println(err)
			return
		}
		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
	} else if command == "info" {
		filename := os.Args[2]
		data, err := os.Open(filename)
		if err != nil {
			fmt.Println("Error reading file:", err)
			return
		}
		defer data.Close()
		if err != nil {
			fmt.Println(err)
			return
		}
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
		chunkSize := 20

		chunks := splitString(jsonObject.Info.Pieces, chunkSize)
		fmt.Printf("Tracker URL: %s\n", jsonObject.Announce)
		fmt.Printf("Length: %d\n", jsonObject.Info.Length)
		fmt.Printf("Info Hash: %x\n", sha1.Sum(buffer_.Bytes()))
		fmt.Printf("Piece Length: %d\n", jsonObject.Info.PiecesLen)
		fmt.Printf("Piece Hashes:\n")
		for _, chunk := range chunks {
			if command == "info" {
				fmt.Printf("%x\n", chunk)
			}
		}

	} else if command == "peers" {
		filename := os.Args[2]
		data, err := os.Open(filename)
		if err != nil {
			fmt.Println("Error reading file:", err)
			return
		}
		defer data.Close()
		if err != nil {
			fmt.Println(err)
			return
		}
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
		get_request(jsonObject, buffer_)
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
