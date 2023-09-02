// package main

// import (
// 	"encoding/json"
// 	"fmt"
// 	"os"
// 	"regexp"
// 	"strconv"
// 	// Available if you need it!
// )

// // Example:
// // - 5:hello -> hello
// // - 10:hello12345 -> hello12345
// func decodeBencode(bencodedString string) (interface{}, error) {
// 	r_string, err := regexp.Compile(`^\d+\:(.*)`)
// 	if err != nil {
// 		return "", fmt.Errorf("only strings are supported at the moment")
// 	}
// 	if r_string.MatchString(bencodedString) {
// 		val := r_string.FindStringSubmatch(bencodedString)
// 		return val[1], nil
// 	}
// 	r_int, err := regexp.Compile(`^i(\-?\d+)e$`)
// 	if err != nil {
// 		return "", fmt.Errorf("only strings are supported at the moment")
// 	}
// 	if r_int.MatchString(bencodedString) {
// 		val := r_int.FindStringSubmatch(bencodedString)
// 		st_int, err := strconv.ParseInt(val[1], 10, 64)
// 		if err != nil {
// 			return "", fmt.Errorf("error while converting string to int")
// 		}
// 		return st_int, nil
// 	}
// 	r_list, err := regexp.Compile(`^l\d+\:(.*)i(\-?\d+)e.*`)
// 	if err != nil {
// 		return "", fmt.Errorf("only strings are supported at the moment")
// 	}
// 	if r_list.MatchString(bencodedString) {
// 		vals := r_list.FindStringSubmatch(bencodedString)
// 		s := make([]interface{}, 0, 2)
// 		for i := 1; i < len(vals); i++ {
// 			if i == 2 {
// 				st_int, err := strconv.ParseInt(vals[i], 10, 64)
// 				if err != nil {
// 					return "", fmt.Errorf("error while converting string to int")
// 				}
// 				s = append(s, st_int)
// 				continue
// 			}
// 			s = append(s, vals[i])
// 		}
// 		return s, nil
// 	}
// 	r_dict, err := regexp.Compile(`^d\d+\:(.*)\d+\:(.*)\d+\:(.*)i(\-?\d+)e.*`)
// 	if err != nil {
// 		return "", fmt.Errorf("error while converting string to int")
// 	}
// 	if r_dict.MatchString(bencodedString) {
// 		vals := r_dict.FindStringSubmatch(bencodedString)
// 		m := map[string]interface{}{}
// 		i := 1
// 		for i < len(vals) {
// 			if i == 3 {
// 				st_int, err := strconv.ParseInt(vals[i+1], 10, 64)
// 				if err != nil {
// 					continue
// 				}
// 				m[vals[i]] = st_int
// 				i += 2
// 				continue
// 			}
// 			m[vals[i]] = vals[i+1]
// 			i += 2
// 		}
// 		return m, nil
// 	}
// 	return "", nil
// }

// func main() {
// 	// You can use print statements as follows for debugging, they'll be visible when running tests.
// 	// fmt.Println("Logs from your program will appear here!")

// 	command := os.Args[1]

// 	// fmt.Println(command, "This is first value")

// 	if command == "decode" {
// 		// Uncomment this block to pass the first stage
// 		//
// 		bencodedValue := os.Args[2]

// 		// fmt.Println(bencodedValue, "This si bencoded value")
// 		decoded, err := decodeBencode(bencodedValue)
// 		// decoded, err := bencode.Decode(bencodedValue)
// 		if err != nil {
// 			fmt.Println(err)
// 			return
// 		}

// 		// fmt.Println(decoded, " this is decoded value")

// 		jsonOutput, _ := json.Marshal(decoded)
// 		fmt.Println(string(jsonOutput))
// 	} else {
// 		fmt.Println("Unknown command: " + command)
// 		os.Exit(1)
// 	}
// }

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Torrent struct {
	Announce string `json:"announce"`
	Info     Info   `json:"info"`
}

type Info struct {
	Length    int64  `json:"length"`
	Name      string `json:"name"`
	PiecesLen int64  `json:"piece length"`
	Pieces    string `json:"pieces"`
}

func decodeBencode(bencodedString *bufio.Reader) (interface{}, error) {
	c, err := bencodedString.ReadByte()
	if err != nil {
		return nil, err
	}
	switch c {
	case 'i':
		return decodeInteger(bencodedString)
	case 'l':
		return decodeList(bencodedString)
	case 'd':
		return decodeDictionary(bencodedString)
	default:
		bencodedString.UnreadByte()
		return decodeString(bencodedString)
	}
}

func decodeInteger(r *bufio.Reader) (int64, error) {
	iBuf, err := readBytesUntil(r, 'e')
	if err != nil {
		return 0, err
	}
	num, err := strconv.ParseInt(string(iBuf), 10, 64)
	if err != nil {
		return 0, err
	}
	return num, nil
}

func decodeList(r *bufio.Reader) ([]interface{}, error) {
	var list []interface{}
	for {
		c, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		if c == 'e' {
			return list, nil
		}
		r.UnreadByte()
		value, err := decodeBencode(r)
		if err != nil {
			return nil, err
		}
		list = append(list, value)
	}
}

func decodeDictionary(r *bufio.Reader) (map[string]interface{}, error) {
	dict := make(map[string]interface{})
	for {
		c, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		if c == 'e' {
			return dict, nil
		}
		r.UnreadByte()
		key, err := decodeString(r)
		if err != nil {
			return nil, err
		}
		value, err := decodeBencode(r)
		if err != nil {
			return nil, err
		}
		dict[key.(string)] = value
	}
}

func decodeString(r *bufio.Reader) (interface{}, error) {
	stringLengthBuffer, err := readBytesUntil(r, ':')
	if err != nil {
		return nil, err
	}
	stringLength, err := strconv.ParseInt(string(stringLengthBuffer), 10, 64)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, stringLength)
	_, err = r.Read(buf)
	if err != nil {
		return nil, err
	}
	return string(buf), nil
}

func readBytesUntil(r *bufio.Reader, delim byte) ([]byte, error) {
	buf, err := r.ReadBytes(delim)
	if err != nil {
		return nil, err
	}
	return buf[:len(buf)-1], nil
}

func main() {
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
		data, err := os.ReadFile(filename)
		if err != nil {
			fmt.Println("Error reading file:", err)
			return
		}
		decoded, err := decodeBencode(bufio.NewReader(strings.NewReader(string(data))))
		if err != nil {
			fmt.Println(err)
			return
		}
		jsonRepr, _ := json.Marshal(decoded)
		var jsonObject Torrent
		if err := json.Unmarshal(jsonRepr, &jsonObject); err != nil {
			fmt.Println("Error unmarshaling JSON:", err)
			return
		}
		fmt.Printf("Tracker URL: %s\n", jsonObject.Announce)
		fmt.Printf("Length: %d\n", jsonObject.Info.Length)
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
