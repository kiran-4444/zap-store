package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

type KVStore struct {
	hashMap map[string]string
}

func (kvs *KVStore) Set(key string, value string) error {
	if key == "" {
		return fmt.Errorf("KEY CANNOT BE EMPTY")
	}

	if value == "" {
		return fmt.Errorf("VALUE CANNOT BE EMPTY")
	}

	kvs.hashMap[key] = value
	return nil
}

func (kvs *KVStore) Get(key string) (string, error) {
	if _, ok := kvs.hashMap[key]; !ok {
		return "", fmt.Errorf("KEY NOT FOUND")
	}

	return kvs.hashMap[key], nil
}

func (kvs *KVStore) Del(key string) {
	delete(kvs.hashMap, key)
}

func (kvs *KVStore) GetAll() string {
	var result string
	for key, value := range kvs.hashMap {
		result += fmt.Sprintf("%s:%s\n", key, value)
	}
	return result
}

func NewKVStore() *KVStore {
	return &KVStore{
		hashMap: make(map[string]string),
	}
}

func main() {
	kvs := NewKVStore()

	for {
		fmt.Print("> ")
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		// Trim the newline character
		line = line[:len(line)-1]

		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		if line == "exit" {
			fmt.Println("Exiting...")
			break
		}

		words := strings.Split(line, " ")

		if len(words) == 0 {
			continue
		}

		if len(words) == 3 {
			if words[0] != "set" {
				fmt.Println("Invalid command: ", words[0])
				continue
			}

			key := words[1]
			value := words[2]
			var err = kvs.Set(key, value)
			if err != nil {
				fmt.Println(err)
				continue
			}

		}

		if len(words) == 2 {
			if words[0] == "get" {
				key := words[1]
				value, err := kvs.Get(key)
				if err != nil {
					fmt.Println(err)
					continue
				}

				fmt.Println(value)
			} else if words[0] == "del" {
				key := words[1]
				kvs.Del(key)
			} else {
				fmt.Println("Invalid command: ", words[0])
				continue
			}

		}

		if len(words) == 1 {
			if words[0] == "getall" {
				fmt.Println(kvs.GetAll())
			} else {
				fmt.Println("Invalid command: ", words[0])
				continue
			}
		}
	}
}
