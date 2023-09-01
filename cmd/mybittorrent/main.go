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
	r_string, err := regexp.Compile(`\d+\:(.*)`)
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
	return "", nil
	// if unicode.IsDigit(rune(bencodedString[0])) {
	// 	var firstColonIndex int

	// 	for i := 0; i < len(bencodedString); i++ {
	// 		if bencodedString[i] == ':' {
	// 			firstColonIndex = i
	// 			break
	// 		}
	// 	}

	// 	lengthStr := bencodedString[:firstColonIndex]

	// 	length, err := strconv.Atoi(lengthStr)
	// 	if err != nil {
	// 		return "", err
	// 	}

	// 	return bencodedString[firstColonIndex+1 : firstColonIndex+1+length], nil
	// } else {
	// 	return "", fmt.Errorf("Only strings are supported at the moment")
	// }
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
