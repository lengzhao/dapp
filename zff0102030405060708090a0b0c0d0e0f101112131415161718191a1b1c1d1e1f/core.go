package zff0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/govm-net/govm/runtime"
)

type dbTransInfo struct{}
type dbCoin struct{}
type dbStat struct{}
type dbApp struct{}
type dbDepend struct{}
type logBlockInfo struct{}
type logSync struct{}

// Hash The KEY of the block of transaction
type Hash [HashLen]byte

// Address the wallet address
type Address [AddressLen]byte

// DependItem App's dependency information
type DependItem struct {
	Alias   [4]byte
	AppName Hash
}

// iRuntime The interface that the executive needs to register
type iRuntime interface {
	//Get the hash of the data
	GetHash(in []byte) []byte
	//Encoding data into data streams.
	Encode(typ uint8, in interface{}) []byte
	//The data stream is filled into a variable of the specified type.
	Decode(typ uint8, in []byte, out interface{}) int
	//Signature verification
	Recover(address, sign, msg []byte) bool
	//The write interface of the database
	DbSet(owner interface{}, key, value []byte, life uint64)
	//The reading interface of the database
	DbGet(owner interface{}, key []byte) ([]byte, uint64)
	//get life of the db
	DbGetLife(owner interface{}, key []byte) uint64
	//The write interface of the log
	LogWrite(owner interface{}, key, value []byte, life uint64)
	//The reading interface of the log
	LogRead(owner interface{}, chain uint64, key []byte) ([]byte, uint64)
	//get life of the log
	LogReadLife(owner interface{}, key []byte) uint64
	//Get the app name with the private structure of app
	GetAppName(in interface{}) []byte
	//New app
	NewApp(name []byte, code []byte)
	//Run app,The content returned is allowed to read across the chain
	RunApp(name, user, data []byte, energy, cost uint64)
	//Event interface for notification to the outside
	Event(user interface{}, event string, param ...[]byte)
	//Consume energy
	ConsumeEnergy(energy uint64)
}

// DB Type definition of a database.
type DB struct {
	owner interface{}
	free  bool
}

// Log Type definition of a log. Log data can be read on other chains. Unable to overwrite the existing data.
type Log struct {
	owner interface{}
}

// AppInfo App info in database
type AppInfo struct {
	Account Address
	LineSum uint64
	Life    uint64
	Flag    uint8
}

// BaseInfo stat info of last block
type BaseInfo struct {
	Key           Hash
	Time          uint64
	Chain         uint64
	ID            uint64
	BaseOpsEnergy uint64
	Producer      Address
	ParentID      uint64
	LeftChildID   uint64
	RightChildID  uint64
}

type processer struct {
	BaseInfo
	iRuntime
	pDbTransInfo  *DB
	pDbCoin       *DB
	pDbStat       *DB
	pDbApp        *DB
	pDbDepend     *DB
	pLogSync      *Log
	pLogBlockInfo *Log
}

// time
const (
	TimeMillisecond = 1
	TimeSecond      = 1000 * TimeMillisecond
	TimeMinute      = 60 * TimeSecond
	TimeHour        = 60 * TimeMinute
	TimeDay         = 24 * TimeHour
	TimeYear        = 31558150 * TimeSecond
	TimeMonth       = TimeYear / 12
)

const (
	// HashLen the byte length of Hash
	HashLen = 32
	// AddressLen the byte length of Address
	AddressLen = 24

	maxBlockInterval   = 1 * TimeMinute
	minBlockInterval   = 10 * TimeMillisecond
	blockSizeLimit     = 1 << 20
	blockSyncMin       = 8 * TimeMinute
	blockSyncMax       = 10 * TimeMinute
	defauldbLife       = 6 * TimeMonth
	adminLife          = 10 * TimeYear
	logLockTime        = 3 * TimeDay
	maxDbLife          = 1 << 50
	maxGuerdon         = 5000000000000
	minGuerdon         = 50000
	prefixOfPlublcAddr = 255
	hateRatioMax       = 1 << 30
	minerNum           = 11
)

//Key of the running state
const (
	StatBaseInfo = uint8(iota)
	StatTransKey
	StatGuerdon
	StatBlockSizeLimit
	StatAvgBlockSize
	StatHashPower
	StatBlockInterval
	StatSyncInfo
	StatFirstBlockKey
	StatChangingConfig
	StatBroadcast
	StatParentKey
	StatUser
	StatAdmin
	StatTotalVotes
)

const (
	// OpsTransfer pTransfer
	OpsTransfer = uint8(iota)
	// OpsMove Move out of coin, move from this chain to adjacent chains
	OpsMove
	// OpsNewChain create new chain
	OpsNewChain
	// OpsNewApp create new app
	OpsNewApp
	// OpsRunApp run app
	OpsRunApp
	// OpsUpdateAppLife update app life
	OpsUpdateAppLife
	// OpsRegisterMiner Registered as a miner
	OpsRegisterMiner
	// OpsRegisterAdmin Registered as a admin
	OpsRegisterAdmin
	// OpsVote vote admin
	OpsVote
	// OpsUnvote unvote
	OpsUnvote
	// OpsReportError error block
	OpsReportError
)

var (
	gBS processer
	// gPublicAddr The address of a public account for the preservation of additional rewards.
	gPublicAddr = Address{prefixOfPlublcAddr, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23}
)

// Empty Check whether Hash is empty
func (h Hash) Empty() bool {
	return h == (Hash{})
}

// MarshalJSON marshal by base64
func (h Hash) MarshalJSON() ([]byte, error) {
	if h.Empty() {
		return json.Marshal(nil)
	}
	return json.Marshal(h[:])
}

// UnmarshalJSON UnmarshalJSON
func (h *Hash) UnmarshalJSON(b []byte) error {
	var v []byte
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}
	copy(h[:], v)
	return nil
}

// ToHexString encode to hex string
func (h Hash) ToHexString() string {
	return hex.EncodeToString(h[:])
}

// Empty Check where Address is empty
func (a Address) Empty() bool {
	return a == (Address{})
}

// MarshalJSON marshal by base64
func (a Address) MarshalJSON() ([]byte, error) {
	if a.Empty() {
		return json.Marshal(nil)
	}
	return json.Marshal(a[:])
}

// UnmarshalJSON UnmarshalJSON
func (a *Address) UnmarshalJSON(b []byte) error {
	var v []byte
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}
	copy(a[:], v)
	return nil
}

// Decode decode hex string
func (a *Address) Decode(hexStr string) {
	d, err := hex.DecodeString(hexStr)
	assert(err == nil)
	Decode(0, d, a)
}

// ToHexString encode to hex string
func (a Address) ToHexString() string {
	return hex.EncodeToString(a[:])
}

func assert(cond bool) {
	if !cond {
		panic("error")
	}
}

func init() {
	bit := 32 << (^uint(0) >> 63)
	assert(bit == 64)
	gBS.pDbTransInfo = GetDB(dbTransInfo{})
	gBS.pDbCoin = GetDB(dbCoin{})
	gBS.pDbCoin.free = true
	gBS.pDbStat = GetDB(dbStat{})
	gBS.pDbStat.free = true
	gBS.pDbApp = GetDB(dbApp{})
	gBS.pDbApp.free = true
	gBS.pDbDepend = GetDB(dbDepend{})
	gBS.pDbDepend.free = true
	gBS.pLogBlockInfo = GetLog(logBlockInfo{})
	gBS.pLogSync = GetLog(logSync{})

	var addrType, address, dbFlag string
	var commandLine = flag.NewFlagSet("core", flag.ContinueOnError)
	commandLine.StringVar(&addrType, "at", "tcp", "address type of server")
	commandLine.StringVar(&address, "addr", "127.0.0.1:17778", "address of server")
	commandLine.StringVar(&dbFlag, "f", "000000000000000000090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f", "db flag(block key)")

	commandLine.Parse(os.Args[1:])

	f, err := hex.DecodeString(dbFlag)
	if err != nil {
		log.Panic("error db flag:", dbFlag, err)
	}
	runt := runtime.NewRuntime(addrType, address)
	runt.SetInfo(1, f)
	gBS.iRuntime = runt
	stream, _ := gBS.DbGet(dbStat{}, []byte{StatBaseInfo})
	if len(stream) == 0 {
		fmt.Print("fail to get base info from dbStat")
		return
	}
	Decode(0, stream, &gBS.BaseInfo)
	if bytes.Compare(f, gBS.Key[:]) != 0 {
		log.Printf("warning. different db flag, cmd:%x, db:%x\n", f, gBS.Key)
	}
}

// GetHash get data hash
func GetHash(data []byte) Hash {
	hashKey := Hash{}
	if len(data) == 0 {
		return hashKey
	}
	gBS.ConsumeEnergy(gBS.BaseOpsEnergy * 10)
	hash := gBS.GetHash(data)
	n := Decode(0, hash, &hashKey)
	assert(n == HashLen)
	return hashKey
}

// encoding type
const (
	EncBinary = uint8(iota)
	EncJSON
	EncGob
)

// Encode Encoding data into data streams.
func Encode(typ uint8, in interface{}) []byte {
	return gBS.Encode(typ, in)
}

// Decode The data stream is filled into a variable of the specified type.
func Decode(typ uint8, in []byte, out interface{}) int {
	return gBS.Decode(typ, in, out)
}

// Recover recover sign
func Recover(address, sign, msg []byte) bool {
	gBS.ConsumeEnergy(gBS.BaseOpsEnergy * 10)
	return gBS.Recover(address, sign, msg)
}

/*-------------------------DB----------------------------------*/

// Set Storage data. the record will be deleted when life=0 or value=nil
func (d *DB) Set(key, value []byte, life uint64) {
	// assert(life <= maxDbLife)
	assert(len(key) > 0)
	assert(len(key) < 200)
	assert(len(value) < 40960)
	size := uint64(len(key) + len(value))
	if d.free {
		gBS.ConsumeEnergy(gBS.BaseOpsEnergy)
	} else if life == 0 || len(value) == 0 {
		value = nil
		life = 0
		gBS.ConsumeEnergy(gBS.BaseOpsEnergy)
	} else if size > 200 {
		// assert(life <= 50*TimeYear)
		t := gBS.BaseOpsEnergy * size * (life + TimeHour - 1) / (TimeHour * 50)
		gBS.ConsumeEnergy(t)
	} else {
		l := gBS.DbGetLife(d.owner, key)
		if l < gBS.Time {
			l = 0
		} else {
			l -= gBS.Time
		}
		var t uint64
		if life > l {
			t = gBS.BaseOpsEnergy * (life + TimeHour - l) / TimeHour
		} else {
			t = gBS.BaseOpsEnergy
		}
		gBS.ConsumeEnergy(t)
	}
	life += gBS.Time
	gBS.DbSet(d.owner, key, value, life)
}

// SetInt Storage uint64 data
func (d *DB) SetInt(key []byte, value uint64, life uint64) {
	v := Encode(0, value)
	d.Set(key, v, life)
}

// Get Read data from database
func (d *DB) Get(key []byte) ([]byte, uint64) {
	assert(len(key) > 0)
	gBS.ConsumeEnergy(gBS.BaseOpsEnergy)
	out, life := gBS.DbGet(d.owner, key)
	if life <= gBS.Time {
		return nil, 0
	}
	return out, (life - gBS.Time)
}

// GetInt read uint64 data from database
func (d *DB) GetInt(key []byte) uint64 {
	v, _ := d.Get(key)
	if v == nil {
		return 0
	}
	var val uint64
	n := Decode(0, v, &val)
	assert(n == len(v))
	return val
}

// GetDB Through the private structure in app, get a DB of app, the parameter must be a structure, not a pointer.
// such as: owner = tAppInfo{}
func GetDB(owner interface{}) *DB {
	out := DB{}
	out.owner = owner
	return &out
}

// Write Write log,if exist the key,return false.the key and value can't be nil.
func (l *Log) Write(key, value []byte) bool {
	assert(len(key) > 0)
	assert(len(value) > 0)
	assert(len(value) < 1024)

	life := gBS.LogReadLife(l.owner, key)
	if life+logLockTime >= gBS.Time {
		return false
	}
	life = TimeYear

	t := 10 * gBS.BaseOpsEnergy * uint64(len(key)+len(value)) * life / TimeDay
	gBS.ConsumeEnergy(t)
	life += gBS.Time
	gBS.LogWrite(l.owner, key, value, life)
	return true
}

// Read Read log
func (l *Log) Read(chain uint64, key []byte) []byte {
	assert(len(key) > 0)
	if chain == 0 {
		chain = gBS.Chain
	}
	dist := getLogicDist(chain, gBS.Chain)
	gBS.ConsumeEnergy(gBS.BaseOpsEnergy * (1 + dist*10))
	minLife := gBS.Time - blockSyncMax*dist
	maxLife := minLife + TimeYear
	out, life := gBS.LogRead(l.owner, chain, key)
	if life < minLife || life > maxLife {
		return nil
	}
	return out
}

// GetLog Through the private structure in app, get a Log of app, the parameter must be a structure, not a pointer.
func GetLog(owner interface{}) *Log {
	out := Log{}
	out.owner = owner
	return &out
}

func getLogicDist(c1, c2 uint64) uint64 {
	var dist uint64
	for {
		if c1 == c2 {
			break
		}
		if c1 > c2 {
			c1 /= 2
		} else {
			c2 /= 2
		}
		dist++
	}
	return dist
}

/***************************** app **********************************/

// GetAppName Get the app name based on the private object
func GetAppName(in interface{}) Hash {
	gBS.ConsumeEnergy(gBS.BaseOpsEnergy)
	out := Hash{}
	name := gBS.GetAppName(in)
	n := Decode(0, name, &out)
	assert(n == len(name))
	return out
}

// GetAppAccount  Get the owner Address of the app
func GetAppAccount(in interface{}) Address {
	app := GetAppName(in)
	assert(!app.Empty())
	info := GetAppInfo(app)
	return info.Account
}

// GetAppInfo get app information
func GetAppInfo(name Hash) *AppInfo {
	out := AppInfo{}
	val, _ := gBS.pDbApp.Get(name[:])
	if len(val) == 0 {
		return nil
	}
	Decode(0, val, &out)
	return &out
}

/*-------------------------------------Coin------------------------*/

// TransferAccounts pTransfer based on the app private object
func TransferAccounts(owner interface{}, payee Address, value uint64) {
	payer := GetAppAccount(owner)
	assert(!payee.Empty())
	assert(!payer.Empty())
	adminTransfer(payer, payee, value)
}

func getAccount(addr Address) (uint64, uint64) {
	v, l := gBS.pDbCoin.Get(addr[:])
	if len(v) == 0 {
		return 0, 0
	}
	var val uint64
	n := Decode(0, v, &val)
	assert(n == len(v))
	return val, l
}

func adminTransfer(payer, payee Address, value uint64) {
	if payer == payee {
		return
	}
	if value == 0 {
		return
	}

	payeeV, payeeL := getAccount(payee)
	payeeV += value
	if payeeV < value {
		return
	}
	if !payer.Empty() {
		v := gBS.pDbCoin.GetInt(payer[:])
		assert(v >= value)
		v -= value
		if v == 0 {
			gBS.pDbCoin.SetInt(payer[:], 0, 0)
		} else {
			gBS.pDbCoin.SetInt(payer[:], v, maxDbLife)
		}
	}
	if !payee.Empty() {
		if payeeV == value {
			gBS.ConsumeEnergy(gBS.BaseOpsEnergy * 1000)
			payeeL = maxDbLife
		}
		gBS.pDbCoin.SetInt(payee[:], payeeV, payeeL)
	}

	Event(dbCoin{}, "pTransfer", payer[:], payee[:], Encode(0, value), Encode(0, payeeV))
}

type tSyncInfo struct {
	ToParentID       uint64
	ToLeftChildID    uint64
	ToRightChildID   uint64
	FromParentID     uint64
	FromLeftChildID  uint64
	FromRightChildID uint64
}

// MoveCost move app cost to other chain(child chain or parent chain)
func MoveCost(user interface{}, chain, cost uint64) {
	assert(chain > 0)
	if gBS.Chain > chain {
		assert(gBS.Chain/2 == chain)
	} else {
		assert(gBS.Chain == chain/2)
		if chain%2 == 0 {
			assert(gBS.LeftChildID > 0)
		} else {
			assert(gBS.RightChildID > 0)
		}
	}
	gBS.ConsumeEnergy(500 * gBS.BaseOpsEnergy)
	addr := GetAppAccount(user)
	adminTransfer(addr, Address{}, cost)
	stru := syncMoveInfo{addr, cost}
	addSyncInfo(chain, SyncOpsMoveCoin, Encode(0, stru))
}

// syncMoveInfo sync Information of move out
type syncMoveInfo struct {
	User  Address
	Value uint64
}

/*------------------------------app--------------------------------------*/

const (
	// AppFlagRun the app can be call
	AppFlagRun = uint8(1 << iota)
	// AppFlagImport the app code can be included
	AppFlagImport
	// AppFlagPlublc App funds address uses the plublc address, except for app, others have no right to operate the address.
	AppFlagPlublc
	// AppFlagGzipCompress gzip compress
	AppFlagGzipCompress
)

// UpdateAppLife update app life
func UpdateAppLife(AppName Hash, life uint64) {
	app := GetAppInfo(AppName)
	assert(app != nil)
	assert(app.Life >= gBS.Time)
	assert(life < 10*TimeYear)
	assert(life > 0)
	app.Life += life
	assert(app.Life > life)
	assert(app.Life < gBS.Time+10*TimeYear)
	deps, _ := gBS.pDbDepend.Get(AppName[:])
	gBS.pDbApp.Set(AppName[:], Encode(0, app), app.Life-gBS.Time)
	t := gBS.BaseOpsEnergy * (life + TimeDay) / TimeHour
	gBS.ConsumeEnergy(t)
	if len(deps) == 0 {
		return
	}
	gBS.pDbDepend.Set(AppName[:], deps, app.Life-gBS.Time)
	for len(deps) > 0 {
		item := Hash{}
		n := Decode(0, deps, &item)
		deps = deps[n:]
		itemInfo := GetAppInfo(item)
		assert(itemInfo != nil)
		assert(itemInfo.Life >= app.Life)
	}
}

/*------------------------------api---------------------------------------*/

// Event send event
func Event(user interface{}, event string, param ...[]byte) {
	gBS.Event(user, event, param...)
}

// GetDBData get data by name.
// name list:dbTransInfo,dbCoin,dbStat,dbApp,logBlockInfo,logSync
func GetDBData(name string, key []byte) ([]byte, uint64) {
	var db *DB
	switch name {
	case "dbTransInfo":
		db = gBS.pDbTransInfo
	case "dbCoin":
		db = gBS.pDbCoin
	case "dbStat":
		db = gBS.pDbStat
	case "dbApp":
		db = gBS.pDbApp
	case "logBlockInfo":
		return gBS.pLogBlockInfo.Read(0, key), 0
	case "logSync":
		return gBS.pLogSync.Read(0, key), 0
	default:
		return nil, 0
	}
	return db.Get(key)
}

// ops of sync
const (
	SyncOpsMoveCoin = iota
	SyncOpsNewChain
	SyncOpsMiner
	SyncOpsBroadcast
	SyncOpsBroadcastAck
	SyncOpsHateRatio
)

type syncHead struct {
	BlockID uint64
	Ops     uint8
}

func getSyncKey(typ byte, index uint64) []byte {
	var key = []byte{typ}
	key = append(key, Encode(0, index)...)
	return key
}

func addSyncInfo(chain uint64, ops uint8, data []byte) {
	var info tSyncInfo
	stream, _ := gBS.pDbStat.Get([]byte{StatSyncInfo})
	if len(stream) > 0 {
		Decode(0, stream, &info)
	}

	var key []byte
	switch chain {
	case gBS.Chain / 2:
		key = getSyncKey('p', info.ToParentID)
		info.ToParentID++
	case 2 * gBS.Chain:
		key = getSyncKey('l', info.ToLeftChildID)
		info.ToLeftChildID++
	case 2*gBS.Chain + 1:
		key = getSyncKey('r', info.ToRightChildID)
		info.ToRightChildID++
	default:
		assert(false)
	}
	head := syncHead{gBS.ID, ops}
	d := Encode(0, head)
	d = append(d, data...)
	gBS.pLogSync.Write(key, d)
	gBS.pDbStat.Set([]byte{StatSyncInfo}, Encode(0, info), maxDbLife)
	Event(logSync{}, "addSyncInfo", []byte{ops}, data)
}

/*-----------------------Just for test------------------------------*/

// SetCostForTest SetCostForTest
func SetCostForTest(account Address, value uint64) {
	adminTransfer(Address{}, account, value)
}

// SetAppAccountForTest SetAppAccountForTest
func SetAppAccountForTest(in interface{}, value uint64) {
	app := GetAppName(in)
	assert(!app.Empty())
	info := GetAppInfo(app)
	if info == nil {
		info = &AppInfo{}
		gBS.Decode(0, app[:], &info.Account)
		info.Account[0] = prefixOfPlublcAddr
		gBS.pDbApp.Set(app[:], gBS.Encode(0, info), TimeYear)
	}
	adminTransfer(Address{}, info.Account, value)
}
