package a1000000000000000000000000000000000000000000000000000000000000000

import core "github.com/lengzhao/dapp/zff0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"

type tApp struct{}

type tTransIn struct {
	Who     []byte `json:"who,omitempty"`
	EthAddr string `json:"eth_addr,omitempty"`
	Cost    uint64 `json:"cost,omitempty"`
	Trans   []byte `json:"trans,omitempty"`
}

type tSign struct{}

type tCash struct {
	EthTrans []byte       `json:"eth_trans,omitempty"`
	Addr     core.Address `json:"addr,omitempty"`
	Cost     uint64       `json:"cost,omitempty"`
}

const (
	// OpsTransIn lock govm
	OpsTransIn = iota
	// OpsSign administrator sign,govm.OpsSign--->eth.relayMint
	OpsSign
	// OpsCash unlock govm by eth.tx, eth.burn--->govm.OpsCash
	OpsCash
)

const limit = 1000000000

var owner = core.Address{2, 152, 64, 16, 49, 156, 211, 70, 89, 247, 252, 178, 11, 49, 214, 21, 216, 80, 171, 50, 202, 147, 6, 24}

func run(user, in []byte, cost uint64) {
	switch in[0] {
	case OpsTransIn:
		var info tTransIn
		if cost < limit {
			panic("request > 1govm")
		}
		core.Decode(core.EncJSON, in[1:], &info)
		if info.EthAddr == "" {
			panic("request eth address")
		}
		info.Trans, _ = core.GetDBData("dbStat", []byte{core.StatTransKey})
		if len(info.Trans) != core.HashLen {
			panic("error trans key")
		}
		if cost/100 > limit {
			core.TransferAccounts(tApp{}, owner, cost/100)
			info.Cost = cost * 99 / 100
		} else {
			core.TransferAccounts(tApp{}, owner, limit)
			info.Cost = cost - limit
		}
		info.Who = user
		db := core.GetDB(tTransIn{})
		data := core.Encode(core.EncJSON, info)
		db.Set(info.Trans, data, core.TimeDay*5)
		core.Event(tApp{}, "OpsTransIn", data)
		db2 := core.GetDB(tApp{})
		id := db2.GetInt([]byte{0}) + 1
		db2.SetInt([]byte{0}, id, core.TimeYear)
		db2.Set(core.Encode(0, id), info.Trans, core.TimeSecond)
	case OpsSign:
		if !isOwner(user) {
			panic("request owner")
		}
		trans := in[1 : 1+core.HashLen]
		sign := in[1+core.HashLen:]
		db := core.GetDB(tSign{})
		db.Set(trans, sign, core.TimeSecond)
		core.Event(tApp{}, "OpsSign", trans)
	case OpsCash:
		if !isOwner(user) {
			panic("request owner")
		}
		info := tCash{}
		core.Decode(core.EncJSON, in[1:], &info)
		if len(info.EthTrans) != 32 {
			panic("request eth trans key")
		}
		if info.Addr.Empty() {
			panic("request govm address")
		}
		db := core.GetDB(tCash{})
		if d, _ := db.Get(info.EthTrans); len(d) > 0 {
			panic("exist eth trans")
		}
		db.Set(info.EthTrans, in[1:], 15*core.TimeDay)

		if info.Cost > limit {
			core.TransferAccounts(tApp{}, info.Addr, info.Cost-limit)
			core.TransferAccounts(tApp{}, owner, limit)
		} else {
			core.TransferAccounts(tApp{}, info.Addr, info.Cost/2)
			core.TransferAccounts(tApp{}, owner, info.Cost/2)
		}
		core.Event(tApp{}, "OpsCash", in[1:])
	default:
		panic("not support")
	}
}

func isOwner(user []byte) bool {
	addr := core.Address{}
	core.Decode(core.EncBinary, user, &addr)
	return addr == owner
}
