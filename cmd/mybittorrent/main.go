package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
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
		fmt.Printf("Tracker URL: %s\n", jsonObject.Announce)
		fmt.Printf("Length: %d\n", jsonObject.Info.Length)
		fmt.Printf("Info Hash: %x\n", sha1.Sum(buffer_.Bytes()))
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
