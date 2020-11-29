package main

import (
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"sync"

	"github.com/lengzhao/database/client"
	"github.com/lengzhao/database/server"
)

func main() {
	var addrType, address, dbServer string
	flag.StringVar(&addrType, "at", "tcp", "address type of db server")
	flag.StringVar(&address, "addr", "127.0.0.1:17778", "listen address")
	flag.StringVar(&dbServer, "db_server", "127.0.0.1:17777", "address of db server")

	flag.Parse()
	db := server.NewRPCObj(".")
	server.RegisterAPI(db, func(dir string, id uint64) server.DBApi {
		return NewProxy(id, addrType, dbServer)
	})

	rpc.Register(db)
	rpc.HandleHTTP()
	lis, err := net.Listen(addrType, address)
	if err != nil {
		log.Fatalln("fatal error: ", err)
	}
	http.Serve(lis, nil)
}

type memKey struct {
	TbName string
	Key    string
}

// DBNWProxy proxy of database, not write data to database
type DBNWProxy struct {
	chain uint64
	mu    sync.Mutex
	flag  []byte
	cache map[memKey][]byte
	dbc   *client.Client
}

// NewProxy new proxy of database
func NewProxy(chain uint64, addType, address string) server.DBApi {
	out := new(DBNWProxy)
	out.chain = chain
	out.cache = make(map[memKey][]byte)
	out.dbc = client.New(addType, address, 1)
	return out
}

// Close close
func (db *DBNWProxy) Close() {
	db.dbc.Close()
}

// OpenFlag open flag
func (db *DBNWProxy) OpenFlag(flag []byte) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if len(db.flag) != 0 {
		log.Println("fail to open flag, exist flag")
		return fmt.Errorf("exist flag")
	}
	db.flag = flag
	db.cache = make(map[memKey][]byte)
	return nil
}

// GetLastFlag return opened flag or nil
func (db *DBNWProxy) GetLastFlag() []byte {
	return db.flag
}

// Commit not write, only reset cache
func (db *DBNWProxy) Commit(flag []byte) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.cache = make(map[memKey][]byte)
	db.flag = nil
	return nil
}

// Cancel only reset cache
func (db *DBNWProxy) Cancel(flag []byte) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.cache = make(map[memKey][]byte)
	db.flag = nil
	return nil
}

// Rollback only reset cache
func (db *DBNWProxy) Rollback(flag []byte) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.cache = make(map[memKey][]byte)
	db.flag = nil
	return nil
}

// SetWithFlag set with flag, only write to cache
func (db *DBNWProxy) SetWithFlag(flag, tbName, key, value []byte) error {
	if bytes.Compare(db.flag, flag) != 0 {
		return fmt.Errorf("different flag")
	}
	mk := memKey{}
	mk.TbName = hex.EncodeToString(tbName)
	mk.Key = hex.EncodeToString(key)
	db.mu.Lock()
	defer db.mu.Unlock()
	db.cache[mk] = value
	return nil
}

// Set only write to cache
func (db *DBNWProxy) Set(tbName, key, value []byte) error {
	mk := memKey{}
	mk.TbName = hex.EncodeToString(tbName)
	mk.Key = hex.EncodeToString(key)
	db.mu.Lock()
	defer db.mu.Unlock()
	db.cache[mk] = value
	return nil
}

// Get get from cache,if not exist, read from database server
func (db *DBNWProxy) Get(tbName, key []byte) []byte {
	mk := memKey{}
	mk.TbName = hex.EncodeToString(tbName)
	mk.Key = hex.EncodeToString(key)
	db.mu.Lock()
	v, ok := db.cache[mk]
	db.mu.Unlock()
	if ok {
		return v
	}
	return db.dbc.Get(db.chain, tbName, key)
}

// Exist if exist return true
func (db *DBNWProxy) Exist(tbName, key []byte) bool {
	mk := memKey{}
	mk.TbName = hex.EncodeToString(tbName)
	mk.Key = hex.EncodeToString(key)
	db.mu.Lock()
	_, ok := db.cache[mk]
	db.mu.Unlock()
	if ok {
		return true
	}
	return db.dbc.Exist(db.chain, tbName, key)
}

// GetNextKey get next key for visit
func (db *DBNWProxy) GetNextKey(tbName, preKey []byte) []byte {
	return db.dbc.GetNextKey(db.chain, tbName, preKey)
}
