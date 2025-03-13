package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"slices"
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
	var sortedKeys []string
	for key := range kvs.hashMap {
		sortedKeys = append(sortedKeys, key)
	}

	slices.Sort(sortedKeys)

	for _, key := range sortedKeys {
		result += fmt.Sprintf("%s:%s\n", key, kvs.hashMap[key])
	}

	return result
}

func NewKVStore() *KVStore {
	return &KVStore{
		hashMap: make(map[string]string),
	}
}

func runLoop(kvs *KVStore, reader *bufio.Reader, out io.Writer) error {
	for {
		fmt.Fprint(out, "> ")
		line, err := reader.ReadString('\n')

		if err != nil {
			return err
		}

		// Trim the newline character
		line = line[:len(line)-1]

		if line == "exit" {
			fmt.Fprintf(out, "Exiting...\n")
			return nil
		}

		if line == "" {
			continue
		}

		words := strings.Split(line, " ")

		if len(words) == 3 {
			if words[0] != "set" {
				fmt.Fprintf(out, "Invalid command: %s\n", words[0])
				continue
			}

			key := words[1]
			value := words[2]
			var err = kvs.Set(key, value)
			if err != nil {
				fmt.Fprintln(out, err)
				continue
			}

		}

		if len(words) == 2 {
			if words[0] == "get" {
				key := words[1]
				value, err := kvs.Get(key)
				if err != nil {
					fmt.Fprintln(out, err)
					continue
				}
				fmt.Fprintln(out, value)
			} else if words[0] == "del" {
				key := words[1]
				kvs.Del(key)
				fmt.Fprintf(out, "Deleted %s\n", key)
			} else {
				fmt.Fprintf(out, "Invalid command: %s\n", words[0])
				continue
			}

		}

		if len(words) == 1 {
			if words[0] == "getall" {
				fmt.Fprintf(out, "%s", kvs.GetAll())
			} else {
				fmt.Fprintf(out, "Invalid command: %s\n", words[0])
				continue
			}
		}

	}
}

func main() {
	kvs := NewKVStore()

	reader := bufio.NewReader(os.Stdin)
	if err := runLoop(kvs, reader, os.Stdout); err != nil {
		log.Fatal(err)
	}

}
