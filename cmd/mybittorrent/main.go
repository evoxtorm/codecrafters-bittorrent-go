package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	// Available if you need it!
)

// Example:
// - 5:hello -> hello
// - 10:hello12345 -> hello12345
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
		m := make(map[string]interface{})
		i := 1
		for i < len(vals) {
			if i == 2 {
				st_int, err := strconv.ParseInt(vals[i+1], 10, 64)
				if err != nil {
					return "", fmt.Errorf("error while converting string to int")
				}
				m[vals[i]] = st_int
				i += 2
				continue
			}
			m[vals[i]] = vals[i+1]
			i += 2
		}
		jsonData, err := json.Marshal(m)
		if err != nil {
			return "", fmt.Errorf("error while converting string to int")
		}
		return string(jsonData), nil
	}
	return "", nil
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	// fmt.Println("Logs from your program will appear here!")

	command := os.Args[1]

	// fmt.Println(command, "This is first value")

	if command == "decode" {
		// Uncomment this block to pass the first stage
		//
		bencodedValue := os.Args[2]

		// fmt.Println(bencodedValue, "This si bencoded value")
		decoded, err := decodeBencode(bencodedValue)
		// decoded, err := bencode.Decode(bencodedValue)
		if err != nil {
			fmt.Println(err)
			return
		}

		// fmt.Println(decoded, " this is decoded value")

		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
