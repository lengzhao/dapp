package a1000000000000000000000000000010203040506070809010203040506070801

import (
	"encoding/hex"
	"fmt"
	"github.com/lengzhao/database/client"
	"testing"
)

func hexToBytes(in string) []byte {
	out, err := hex.DecodeString(in)
	if err != nil {
		fmt.Println("fail to decode hex:", err)
		panic(err)
	}
	return out
}

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

	run(hexToBytes("02984010319cd34659f7fcb20b31d615d850ab32ca930618"), []byte("parament"), 10)
}
