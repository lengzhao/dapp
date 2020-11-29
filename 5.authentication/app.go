package a1000000000000000000000000000000000000000000000000000000000000000

import (
	core "github.com/lengzhao/dapp/zff0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"
)

type tApp struct{}

type tCategory struct {
	ID          string `json:"id,omitempty"`
	Description string `json:"desc,omitempty"`
}

type tUserByOwner struct{}

type tUser struct{}

// User user info
type User struct {
	Addr        core.Address `json:"addr,omitempty"`
	Category    string       `json:"category,omitempty"` // tCategory.ID
	ID          string       `json:"id,omitempty"`
	Description string       `json:"desc,omitempty"`
}

type lifeInfo struct {
	DbType int    `json:"db_type,omitempty"`
	Key    []byte `json:"key,omitempty"`
}

type syncInfo struct {
	DbType int    `json:"db_type,omitempty"`
	Key    []byte `json:"key,omitempty"`
	Value  []byte `json:"value,omitempty"`
}

type ackInfo struct {
	FromChain uint64 `json:"from_chain,omitempty"`
	Key       []byte `json:"key,omitempty"` // transaction key
}

const (
	// OpsNewCategory net Category
	opsNewCategory = iota
	// OpsNewUserByOwner new user by owner
	opsNewUserByOwner
	// OpsNewUser Anyone can enter, but the data is not necessarily correct
	opsNewUser
	// OpsClear clear data
	opsClear
	// opsUpdateLife update data life
	opsUpdateLife
	// opsSync sync data to other chain
	opsSync
	// opsAck ack data from other chain
	opsAck
)

var owner = "01ccaf415a3a6dc8964bf935a1f40e55654a4243ae99c709"

func run(user, in []byte, cost uint64) {
	switch in[0] {
	case opsNewCategory:
		// new category. The owner can modify the information
		var info tCategory
		core.Decode(core.EncJSON, in[1:], &info)
		if info.ID == "" {
			panic("request id")
		}

		db := core.GetDB(tCategory{})
		data, _ := db.Get([]byte(info.ID))
		if len(data) != 0 {
			addr := core.Address{}
			core.Decode(0, user, &addr)
			if addr.ToHexString() != owner {
				panic("exist category")
			}
		}
		data = core.Encode(core.EncJSON, info)
		db.Set([]byte(info.ID), data, core.TimeYear)
		core.Event(tApp{}, "OpsNewCategory", data)
	case opsNewUserByOwner:
		addr := core.Address{}
		core.Decode(0, user, &addr)
		if addr.ToHexString() != owner {
			panic("not owner")
		}
		var info User
		core.Decode(core.EncJSON, in[1:], &info)
		if info.Category == "" || info.ID == "" {
			panic("empty value")
		}
		data, _ := core.GetDB(tCategory{}).Get([]byte(info.Category))
		if len(data) == 0 {
			panic("not found Category")
		}
		keyMsg := append(info.Addr[:], []byte(info.Category)...)
		key := core.GetHash(keyMsg)
		db := core.GetDB(tUserByOwner{})
		data, _ = db.Get(key[:])
		if len(data) > 0 {
			panic("exist user")
		}
		data = core.Encode(core.EncJSON, info)
		db.Set(key[:], data, core.TimeYear)
		core.Event(tApp{}, "OpsNewUserByOwner", data)
	case opsNewUser:
		var info User
		core.Decode(core.EncJSON, in[1:], &info)
		core.Decode(core.EncBinary, user, &info.Addr)
		data, _ := core.GetDB(tCategory{}).Get([]byte(info.Category))
		if len(data) == 0 {
			panic("not found Category")
		}
		if info.ID == "" {
			panic("request id")
		}
		keyMsg := append(info.Addr[:], []byte(info.Category)...)
		key := core.GetHash(keyMsg)
		db := core.GetDB(tUser{})
		data, _ = db.Get(key[:])
		if len(data) > 0 {
			panic("exist user")
		}
		data = core.Encode(core.EncJSON, info)
		db.Set(key[:], data, core.TimeYear)
		core.Event(tApp{}, "OpsNewUser", data)
	case opsClear:
		addr := core.Address{}
		dbID := in[1]
		core.Decode(0, user, &addr)
		if addr.ToHexString() != owner {
			panic("not owner")
		}
		var db *core.DB
		switch dbID {
		case DbCategory:
			db = core.GetDB(tCategory{})
		case DbUser:
			db = core.GetDB(tUser{})
		case DbUserByOwner:
			db = core.GetDB(tUserByOwner{})
		default:
			panic("not support")
		}
		key := in[2:]
		db.Set(key, []byte("{}"), core.TimeMillisecond)
	case opsUpdateLife:
		var info lifeInfo
		core.Decode(core.EncJSON, in[1:], &info)
		if len(info.Key) == 0 {
			panic("request key")
		}
		var db *core.DB
		switch info.DbType {
		case DbCategory:
			db = core.GetDB(tCategory{})
		case DbUser:
			db = core.GetDB(tUser{})
		case DbUserByOwner:
			db = core.GetDB(tUserByOwner{})
		default:
			panic("not support db type")
		}
		data, _ := db.Get(info.Key)
		if len(data) == 0 {
			panic("not exist the key")
		}
		db.Set(info.Key, data, core.TimeYear)
	case opsSync:
		var info syncInfo
		core.Decode(core.EncJSON, in[1:], &info)
		if len(info.Key) == 0 {
			panic("request key")
		}
		var db *core.DB
		switch info.DbType {
		case DbCategory:
			db = core.GetDB(tCategory{})
		case DbUser:
			db = core.GetDB(tUser{})
		case DbUserByOwner:
			db = core.GetDB(tUserByOwner{})
		default:
			panic("not support db type")
		}
		data, _ := db.Get(info.Key)
		if len(data) == 0 {
			panic("not exist the key")
		}
		info.Value = data

		key, _ := core.GetDBData("dbStat", []byte{core.StatTransKey})
		log := core.GetLog(tApp{})
		log.Write(key, core.Encode(core.EncJSON, info))
		core.Event(tApp{}, "sync", key)
	case opsAck:
		var info ackInfo
		core.Decode(core.EncJSON, in[1:], &info)
		if len(info.Key) == 0 {
			panic("request transaction key")
		}
		log := core.GetLog(tApp{})
		data := log.Read(info.FromChain, info.Key)
		if len(data) == 0 {
			panic("not exist the log")
		}
		var sInfo syncInfo
		core.Decode(core.EncJSON, data, &sInfo)
		var db *core.DB
		switch sInfo.DbType {
		case DbCategory:
			db = core.GetDB(tCategory{})
		case DbUser:
			db = core.GetDB(tUser{})
		case DbUserByOwner:
			db = core.GetDB(tUserByOwner{})
		default:
			panic("not support db type")
		}
		db.Set(sInfo.Key, sInfo.Value, core.TimeYear)
	default:
		panic("not support")
	}
}

// db type
const (
	DbCategory = iota
	DbUser
	DbUserByOwner
)

// Check if exist the key,return true
func Check(dbType int, key []byte) bool {
	var db *core.DB
	switch dbType {
	case DbCategory:
		db = core.GetDB(tCategory{})
	case DbUser:
		db = core.GetDB(tUser{})
	case DbUserByOwner:
		db = core.GetDB(tUserByOwner{})
	default:
		return false
	}

	_, life := db.Get(key)
	if life > core.TimeSecond {
		return true
	}
	return false
}

// GetUser get user info, entered by the user
func GetUser(category string, user core.Address) User {
	var out User
	db := core.GetDB(tUser{})
	keyMsg := append(user[:], []byte(category)...)
	key := core.GetHash(keyMsg)
	data, _ := db.Get(key[:])
	if len(data) > 0 {
		core.Decode(core.EncJSON, data, &out)
	}

	return out
}

// GetTrustedUser get user info, write by owner
func GetTrustedUser(category string, addr core.Address) User {
	var out User
	db := core.GetDB(tUserByOwner{})
	keyMsg := append(addr[:], []byte(category)...)
	key := core.GetHash(keyMsg)
	data, _ := db.Get(key[:])
	if len(data) > 0 {
		core.Decode(core.EncJSON, data, &out)
	}

	return out
}
