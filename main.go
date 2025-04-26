package main

import (
	"bufio"
	"flag"
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

	logFileName := "logFile.log"
	logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed to open log file: %s", err)
	}
	defer logFile.Close()

	// log.SetOutput(io.MultiWriter(logFile, os.Stderr))
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Llongfile)

	var engineFlag = flag.String("engine", "inmem", "Storage engine to use (inmem or bitcask)")
	var dataDirFlag = flag.String("dataDir", "", "Directory for BitCask data files")
	flag.Parse()

	log.Printf("Starting with storage engine: %s\n", *engineFlag)

	var storageEngine storage.StorageEngine

	switch *engineFlag {
	case "inmem":
		storageEngine = inmem.NewInMemStorageEngine()
	case "bitcask":
		if *dataDirFlag == "" {
			log.Fatal("Please specify a data directory for BitCask using the -dataDir flag")
		}

		log.Printf("Using BitCask storage engine with data directory: %s\n", *dataDirFlag)

		var err error
		storageEngine, err = bitcask.NewBitCaskStorageEngine(*dataDirFlag)
		if err != nil {
			log.Fatal(err)
		}
		defer storageEngine.Close()
	default:
		log.Fatal("usage: specify at least one storage engine: inmem or bitcask")
	}

	kvs := zapstore.NewZapStore(storageEngine)

	reader := bufio.NewReader(os.Stdin)
	if err := runLoop(kvs, reader, os.Stdout); err != nil {
		log.Fatal(err)
	}
}
