package a1000000000000000000000000000010203040506070809010203040506070804

import core "github.com/lengzhao/dapp/zff0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"

type tApp struct{}

type syncInfo struct {
	ToChain uint64 `json:"to_chain,omitempty"`
	Data    string `json:"data,omitempty"`
}

type ackInfo struct {
	FromChain uint64 `json:"from_chain,omitempty"`
	Key       []byte `json:"key,omitempty"`
}

const (
	// OpsSync sync info to other chains
	OpsSync = byte(iota + 1)
	// OpsAck ack
	OpsAck
)

func run(user, in []byte, cost uint64) {
	switch in[0] {
	case OpsSync:
		log := core.GetLog(tApp{})
		key, _ := core.GetDBData("dbStat", []byte{core.StatTransKey})
		if len(key) == 0 {
			panic("fail to get transaction key")
		}
		var info syncInfo
		core.Decode(core.EncJSON, in[1:], &info)
		// info.ToChain == 0: broadcast
		ok := log.Write(key, core.Encode(core.EncJSON, info))
		if !ok {
			panic("fail to write log")
		}
		core.Event(tApp{}, "sync", key, in[1:])
	case OpsAck:
		var info ackInfo
		var sInfo syncInfo
		var baseInfo core.BaseInfo
		core.Decode(core.EncJSON, in[1:], &info)
		d, _ := core.GetDBData("dbStat", []byte{core.StatBaseInfo})
		core.Decode(0, d, &baseInfo)

		log := core.GetLog(tApp{})
		data := log.Read(info.FromChain, info.Key)
		core.Decode(core.EncJSON, data, &sInfo)
		if sInfo.ToChain != baseInfo.Chain && sInfo.ToChain != 0 {
			panic("error chain")
		}
		core.Event(tApp{}, "ack", info.Key, data)
	default:
		panic("not support")
	}
}
