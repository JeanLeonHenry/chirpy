package database

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"slices"
	"sync"
)

const ERR_CHIRP_TOO_LONG = "Chirp is too long. Max length is 140."

type Chirp struct {
	Body string `json:"body"`
	Id   int    `json:"id"`
}
type User struct {
	Email string `json:"email"`
	Id    int    `json:"id"`
}

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users  map[int]User  `json:"users"`
}

// NewDB creates a new database connection
// and creates the database file if it doesn't exist
func NewDB(path string) (*DB, error) {
	db := &DB{path: path, mux: &sync.RWMutex{}}
	db.ensureDB()
	return db, nil
}

// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateChirp(body string) (Chirp, error) {
	c := new(Chirp)
	err := json.Unmarshal([]byte(body), c)
	if err != nil {
		return Chirp{}, err
	}
	if len(c.Body) > 140 {
		return Chirp{}, errors.New(ERR_CHIRP_TOO_LONG)
	}
	db.ensureDB()
	dbStructure, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}
	log.Printf("loaded %v chirps", len(dbStructure.Chirps))
	if dbStructure.Chirps == nil {
		dbStructure.Chirps = make(map[int]Chirp)
	}
	var maxId int
	for k := range dbStructure.Chirps {
		if k > maxId {
			maxId = k
		}
	}
	c.Id = maxId + 1
	dbStructure.Chirps[maxId+1] = *c
	err = db.writeDB(dbStructure)
	if err != nil {
		return Chirp{}, err
	}
	return dbStructure.Chirps[maxId+1], err
}

// CreateUser creates a new chirp and saves it to disk
func (db *DB) CreateUser(body string) (User, error) {
	user := new(User)
	err := json.Unmarshal([]byte(body), user)
	if err != nil {
		return User{}, err
	}
	db.ensureDB()
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}
	log.Printf("loaded %v users", len(dbStructure.Users))
	if dbStructure.Users == nil {
		dbStructure.Users = make(map[int]User)
	}
	var maxId int
	for k := range dbStructure.Users {
		if k > maxId {
			maxId = k
		}
	}
	user.Id = maxId + 1
	dbStructure.Users[maxId+1] = *user
	err = db.writeDB(dbStructure)
	if err != nil {
		return User{}, err
	}
	return dbStructure.Users[maxId+1], err
}

// GetChirps returns all chirps in the database
func (db *DB) GetChirps() ([]Chirp, error) {
	data, err := db.loadDB()
	if err != nil {
		return nil, err
	}
	result := []Chirp{}
	for _, chirp := range data.Chirps {
		result = append(result, chirp)
	}
	slices.SortFunc(result, func(c, d Chirp) int { return c.Id - d.Id })
	return result, nil
}

// ensureDB creates a new database file if it doesn't exist
func (db *DB) ensureDB() error {
	db.mux.Lock()
	defer db.mux.Unlock()

	_, err := os.ReadFile(db.path)
	if os.IsNotExist(err) {
		err := os.WriteFile(db.path, []byte{}, 0666)
		if err != nil {
			return err
		}
	}
	return nil
}

// loadDB reads the database file into memory
func (db *DB) loadDB() (DBStructure, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	data, err := os.ReadFile(db.path)
	if err != nil {
		return DBStructure{}, err
	}
	dbStructure := new(DBStructure)
	json.Unmarshal(data, dbStructure)
	return *dbStructure, nil
}

// writeDB writes the database file to disk
func (db *DB) writeDB(dbStructure DBStructure) error {
	db.mux.Lock()
	defer db.mux.Unlock()

	data, err := json.Marshal(dbStructure)
	if err != nil {
		return err
	}
	err = os.WriteFile(db.path, data, 0666)
	return err
}
