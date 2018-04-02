package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
	"github.com/speps/go-hashids"
	"hash/adler32"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	apiShortEndpoint = "/v1/short"
	hashMinLength    = 4
)

var (
	listenHostPort = flag.String("listen", "127.0.0.1:8080", "Listen interface")
	urlScheme      = flag.String("scheme", "http", "URL scheme http or https")
	hashSalt       = flag.String("salt", "tFZ2cQ7U8OQlSWOyZIoFdRusvkFvJh3A", "Hash salt string")
	dbFilename     = flag.String("dbpath", "goshort.db", "Path to the database file")
	dbBucket       = flag.String("dbbucket", "goshort", "DB `bucket` name")
)

// Global DB connector valiable
var db *bolt.DB

// Request object
type shortRequest struct {
	LongURL string `json:"url"`
}

// Response object
type shortResponse struct {
	ShortURL string `json:"shortURL"`
}

// Hash LongURL. Result is short string
func (s *shortRequest) Hash() (string, error) {
	// Make integer from LongURL using adler32 hash function
	urlNumbered := int(adler32.Checksum([]byte(s.LongURL)))

	hd := hashids.NewData()
	hd.Salt = *hashSalt
	hd.MinLength = hashMinLength

	h := hashids.NewWithData(hd)

	// Convert integer to short hash string
	hashedURL, err := h.Encode([]int{urlNumbered})
	if err != nil {
		return "", err
	}

	return hashedURL, nil
}

// Write to the DB
func writeToDatabase(db *bolt.DB, key, value string) error {
	if err := db.Update(func(tx *bolt.Tx) error {
		// Make sure bucket exists. Create It if not
		b, err := tx.CreateBucketIfNotExists([]byte(*dbBucket))
		if err != nil {
			return err
		}

		// Write key, value pair to the DB
		err = b.Put([]byte(key), []byte(value))
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}

// Read from DB by key
func readFromDatabase(db *bolt.DB, key string) string {
	var value string

	if err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(*dbBucket))
		if b == nil {
			return nil
		}

		data := b.Get([]byte(key))

		value = string(data)
		return nil
	}); err != nil {
		log.Fatal(err)
	}

	return value
}

// RedirectHandler handles GET HTTP requests and redirects to longURL
func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	// Read variables from mux router
	vars := mux.Vars(r)

	longURL := readFromDatabase(db, vars["hashID"])
	if longURL == "" {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Not Found")
		return
	}
	http.Redirect(w, r, longURL, 301)
	return

}

// APIShortHandler handles POST HTTP requests
func APIShortHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Read JSON payload as binary data
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Can't read request body"))
		return
	}

	// Unmarshal binary data to the shortRequest object (JSON)
	longURL := &shortRequest{}
	err = json.Unmarshal(body, longURL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Can't unmarshal JSON"))
		return
	}

	// Hash longURL
	hash, err := longURL.Hash()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Can't hash URL"))
	}

	// Make short URL based on request
	shortURL := fmt.Sprintf("%s://%s/%s", *urlScheme, r.Host, hash)
	resp := shortResponse{
		ShortURL: shortURL,
	}

	// Return short URL to user
	respJSON, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Can't marshal as JSON"))
		return
	}

	// Save to the DB
	err = writeToDatabase(db, hash, longURL.LongURL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Can't save to DB"))
		return
	}

	w.Write(respJSON)
}

func main() {
	// Parse flags
	flag.Parse()

	// Open the database.
	var err error
	db, err = bolt.Open(*dbFilename, 0666, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Handle requests
	r := mux.NewRouter()
	r.HandleFunc("/{hashID}", RedirectHandler).Methods("GET")
	r.HandleFunc(apiShortEndpoint, APIShortHandler).Methods("POST")
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(*listenHostPort, r))

	// Close the database
	if err := db.Close(); err != nil {
		log.Fatal(err)
	}
}
