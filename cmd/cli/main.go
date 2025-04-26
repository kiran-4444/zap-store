package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Welcome to ZapStore CLI!")
	fmt.Println("Type 'exit' to quit.")
	for {
		fmt.Print("zapstore=> ")

		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("error reading input:", err)
			continue
		}
		input = strings.TrimSpace(input)

		if input == "exit" {
			break
		}

		args := strings.Fields(input)
		if len(args) < 2 {
			fmt.Println("invalid command")
			continue
		}

		command := strings.ToUpper(args[0])
		key := args[1]

		switch command {
		case "GET":
			resp, err := http.Get(fmt.Sprintf("http://localhost:8080/get?key=%s", key))
			if err != nil {
				fmt.Println("error:", err)
				continue
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			fmt.Println(string(body))

		case "SET":
			if len(args) < 3 {
				fmt.Println("usage: SET key value")
				continue
			}
			value := args[2]
			payload := map[string]string{
				"key":   key,
				"value": value,
			}
			data, _ := json.Marshal(payload)

			resp, err := http.Post("http://localhost:8080/set", "application/json", bytes.NewReader(data))
			if err != nil {
				fmt.Println("error:", err)
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				fmt.Println("OK")
			} else {
				body, _ := io.ReadAll(resp.Body)
				fmt.Println("error:", string(body))
			}

		case "DELETE":
			req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("http://localhost:8080/delete?key=%s", key), nil)
			if err != nil {
				fmt.Println("error:", err)
				continue
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				fmt.Println("error:", err)
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				fmt.Println("OK")
			} else {
				body, _ := io.ReadAll(resp.Body)
				fmt.Println("error:", string(body))
			}

		default:
			fmt.Println("unknown command:", command)
		}
	}
}
