package mirage

import (
	"fmt"
	"log"
	"encoding/json"
	"errors"

	"code.google.com/p/leveldb-go/leveldb"
	"code.google.com/p/leveldb-go/leveldb/db"
	"github.com/acidlemon/go-dumper"

)

var ErrNotFound = errors.New("Not Found")

type MirageStorage struct {
	storage *leveldb.DB
}

func NewMirageStorage() *MirageStorage {
	storage, err := leveldb.Open("./data", &db.Options{})
	if err != nil {
		fmt.Println("cannot open leveldb")
		log.Fatal(err)
	}

	ms := &MirageStorage{ storage: storage }

	return ms
}

func (ms *MirageStorage) Close() {
	ms.storage.Close()
}

func (ms *MirageStorage) Get(key string) ([]byte, error) {
	data, err := ms.storage.Get([]byte(key), nil)
	if err == db.ErrNotFound {
		err = ErrNotFound
	}

	return data, err
}

func (ms *MirageStorage) Set(key string, value []byte) error {
	err := ms.storage.Set([]byte(key), value, &db.WriteOptions{Sync: true})

	return err
}

func (ms *MirageStorage) AddToSubdomainMap(subdomain string) error {
	subdomainMap, err := ms.GetSubdomainMap()
	if err != nil {
		return errors.New(fmt.Sprintf("failed to get subdomain-map: %s", err.Error()))
	}

	beforeLen := len(subdomainMap)

	subdomainMap[subdomain] = 1 // meanless value

	if beforeLen == len(subdomainMap) {
		// need not to update
		fmt.Println("subdomainMap length is not changed!")
		return nil
	}

	return ms.updateSubdomainMap(subdomainMap)
}

func (ms *MirageStorage) RemoveFromSubdomainMap(subdomain string) error {
	subdomainMap, err := ms.GetSubdomainMap()
	if err != nil {
		return errors.New(fmt.Sprintf("failed to get subdomain-map: %s", err.Error()))
	}

	beforeLen := len(subdomainMap)

	delete(subdomainMap, subdomain)

	if beforeLen == len(subdomainMap) {
		return nil
	}

	return ms.updateSubdomainMap(subdomainMap)
}

func (ms *MirageStorage) GetSubdomainMap() (map[string]int, error) {
	subdomainData, err := ms.Get("subdomain-map")
	if err != nil {
		if err != ErrNotFound {
			return nil, err
		}
		subdomainData = []byte(`{}`)
	}

	// Q. why map?  A. easy to manage subdomains as unique
	var subdomainMap map[string]int
	err = json.Unmarshal(subdomainData, &subdomainMap)
	if err != nil {
		return nil, err
	}

	return subdomainMap, nil
}

func (ms *MirageStorage) updateSubdomainMap(subdomainMap map[string]int) error {
	fmt.Println("willUpdate")
	dump.Dump(subdomainMap)
	subdomainData, err := json.Marshal(subdomainMap)

	err = ms.Set("subdomain-map", subdomainData)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to update subdomain-map: %s", err.Error()))
	}

	return nil
}

