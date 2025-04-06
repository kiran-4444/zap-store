package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"zap-store/internal/storage"
	"zap-store/internal/storage/bitcask"
	"zap-store/internal/storage/inmem"
	"zap-store/internal/zapstore"
)

func runLoop(kvs *zapstore.ZapStore, reader *bufio.Reader, out io.Writer) error {
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

		} else if len(words) == 2 {
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

		} else {
			fmt.Fprintf(out, "Invalid command: %s\n", line)
			continue
		}
	}
}

func main() {
	var args = os.Args
	if len(args) < 2 {
		log.Fatal("usage: specify atleast one storage engine: inmem or bitcask")
	}
	var storageEngine storage.StorageEngine
	if args[1] == "inmem" {
		storageEngine = inmem.NewInMemStorageEngine()
	} else if args[1] == "bitcask" {
		if len(args) != 3 {
			log.Fatal("usage: specify data directory for bitcask")
		}
		var err error
		storageEngine, err = bitcask.NewBitCaskStorageEngine(args[2])
		defer storageEngine.Close()

		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal("usage: specify atleast one storage engine: inmem or bitcask")
	}

	kvs := zapstore.NewZapStore(storageEngine)

	reader := bufio.NewReader(os.Stdin)
	if err := runLoop(kvs, reader, os.Stdout); err != nil {
		log.Fatal(err)
	}
}
