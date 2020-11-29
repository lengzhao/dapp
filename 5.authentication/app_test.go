package a1000000000000000000000000000000000000000000000000000000000000000

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/govm-net/govm/runtime"
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

	// add category
	category := "alias"
	info := tCategory{category, "alias of user"}
	data := runtime.JSONEncode(info)
	param := append([]byte{opsNewCategory}, data...)
	run(hexToBytes(owner), param, 0)
}

func TestNewUser(t *testing.T) {
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

	// add category
	category := "alias"
	info := tCategory{category, "alias of user"}
	data := runtime.JSONEncode(info)
	param := append([]byte{opsNewCategory}, data...)
	run(hexToBytes(owner), param, 0)

	// new user
	var user User
	userID := "02984010319cd34659f7fcb20b31d615d850ab32ca930618"
	user.Addr.Decode(userID)
	user.Category = category
	user.ID = "dev"
	user.Description = "description"
	data1 := runtime.JSONEncode(user)
	param1 := append([]byte{opsNewUser}, data1...)
	run(hexToBytes(userID), param1, 0)

	userInfo := GetUser(category, user.Addr)
	if userInfo.ID == "" {
		t.Error("fail to get user")
	}
}

func TestNewUserByOwner(t *testing.T) {
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

	// add category
	category := "alias"
	info := tCategory{category, "alias of user"}
	data := runtime.JSONEncode(info)
	param := append([]byte{opsNewCategory}, data...)
	run(hexToBytes(owner), param, 0)

	// new user
	var user User
	userID := "02984010319cd34659f7fcb20b31d615d850ab32ca930618"
	user.Addr.Decode(userID)
	user.Category = category
	user.ID = "dev"
	user.Description = "description"
	data1 := runtime.JSONEncode(user)
	param1 := append([]byte{opsNewUserByOwner}, data1...)
	run(hexToBytes(owner), param1, 0)
	userInfo := GetTrustedUser(category, user.Addr)
	if userInfo.ID == "" {
		t.Error("fail to get user")
	}
}

func TestClear(t *testing.T) {
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

	// add category
	category := "alias"
	info := tCategory{category, "alias of user"}
	data := runtime.JSONEncode(info)
	param := append([]byte{opsNewCategory}, data...)
	run(hexToBytes(owner), param, 0)

	// new user
	var user User
	userID := "02984010319cd34659f7fcb20b31d615d850ab32ca930618"
	user.Addr.Decode(userID)
	user.Category = category
	user.ID = "dev"
	user.Description = "description"
	data1 := runtime.JSONEncode(user)
	param1 := append([]byte{opsNewUser}, data1...)
	run(hexToBytes(userID), param1, 0)

	userInfo := GetUser(category, user.Addr)
	if userInfo.ID == "" {
		t.Error("fail to get user")
	}

	keyMsg := append(user.Addr[:], []byte(category)...)
	key := runtime.GetHash(keyMsg)
	param2 := append([]byte{opsClear, DbUser}, key...)
	run(hexToBytes(owner), param2, 0)

	userInfo2 := GetUser(category, user.Addr)
	if userInfo2.ID != "" {
		t.Error("exist the user")
	}
}
