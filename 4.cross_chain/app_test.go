package a1000000000000000000000000000010203040506070809010203040506070804

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/govm-net/govm/counter"
	"github.com/govm-net/govm/runtime"
	core "github.com/lengzhao/dapp/zff0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"
	"github.com/lengzhao/database/client"
)

func hexToBytes(in string) []byte {
	out, err := hex.DecodeString(in)
	if err != nil {
		fmt.Println("fail to decode hex:", err)
		panic(err)
	}
	return out
}

const user = "02984010319cd34659f7fcb20b31d615d850ab32ca930618"

// Test_run It will get the block info, so the server needs to connect to the db server
func Test_run(t *testing.T) {
	var chain uint64 = 1
	flg := hexToBytes("000000000000000000090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f")
	c := client.New("tcp", "127.0.0.1:17778", 1)
	err := c.OpenFlag(chain, flg)
	if err != nil {
		t.Error("fail to open Flag,", err)
		f := c.GetLastFlag(chain)
		c.Cancel(chain, f)
		return
	}
	defer c.Cancel(chain, flg)

	// set transaction key
	counter.SetEnergy(1e10)
	baseInfo := core.BaseInfo{}
	d, _ := core.GetDBData("dbStat", []byte{core.StatBaseInfo})
	if len(d) > 0 {
		core.Decode(0, d, &baseInfo)
	}
	core.Decode(0, flg, &baseInfo.Key)
	core.SetStatForTest(core.StatBaseInfo, core.Encode(0, baseInfo))
	core.SetStatForTest(core.StatTransKey, flg)

	info := syncInfo{0, "alias of user"}
	data := runtime.JSONEncode(info)
	param := append([]byte{1}, data...)
	run(hexToBytes(user), param, 0)

	info1 := ackInfo{0, flg}
	data1 := runtime.JSONEncode(info1)
	param1 := append([]byte{2}, data1...)
	run(hexToBytes(user), param1, 0)
}
