package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"zap-store/internal/storage"
	"zap-store/internal/storage/bitcask"
	"zap-store/internal/storage/inmem"
	"zap-store/internal/zapstore"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func setHandler(kvs *zapstore.ZapStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := kvs.Set(req.Key, req.Value); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)

	}
}

func getHandler(kvs *zapstore.ZapStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "missing key parameter", http.StatusBadRequest)
			return
		}
		value, err := kvs.Get(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.Write([]byte(value))
	}
}

func deleteHandler(kvs *zapstore.ZapStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "missing key parameter", http.StatusBadRequest)
			return
		}
		err := kvs.Delete(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func StartServer(kv *zapstore.ZapStore) {
	mux := http.NewServeMux()
	mux.Handle("/set", loggingMiddleware(setHandler(kv)))
	mux.Handle("/get", loggingMiddleware(getHandler(kv)))
	mux.Handle("/delete", loggingMiddleware(deleteHandler(kv)))

	fmt.Println("Server started at :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
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
	log.SetOutput(os.Stderr)
	// log.SetFlags(log.LstdFlags | log.Llongfile)
	log.SetFlags(log.Ltime | log.Ldate | log.LUTC | log.Lmicroseconds)

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

	StartServer(kvs)

}
