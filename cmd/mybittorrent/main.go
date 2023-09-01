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
		val := r_list.FindStringSubmatch(bencodedString)
		return val[1:], nil
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
