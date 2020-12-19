package zcdc09c276df3af32f8f60c33d57710374fe126f482fa50db61271bac0886e600

import (
	"encoding/hex"
	"fmt"
	"testing"

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

	core.SetAppAccountForTest(tApp{}, 1e15)

	// new task
	var cost uint64 = 1e10
	var info Task
	info.Title = "task001"
	info.Number = 10
	data := runtime.JSONEncode(info)
	param := append([]byte{opsNewTask}, data...)
	run(hexToBytes(owner), param, cost)
	id1 := GetLastIDOfTask()
	if id1 != 1 {
		t.Error("error task id")
	}
	statDB := core.GetDB(tTaskStatus{})
	id2 := statDB.GetInt(hexToBytes(owner))
	if id1 != id2 {
		t.Error("different id")
	}

	// accept task
	user1 := "02984010319cd34659f7fcb20b31d615d850ab32ca930618"
	taskID := GetLastIDOfTask()
	data2 := runtime.Encode(taskID)
	param1 := append([]byte{opsAcceptTask}, data2...)
	run(hexToBytes(user1), param1, 0)

	actionID := GetLastIDOfAction()
	action := GetAction(actionID)
	if action.Status != StatusAccept {
		t.Error("error action status")
	}
	sActionDB := core.GetDB(tActionStatus{})
	id3 := sActionDB.GetInt(hexToBytes(user1))
	if actionID != id3 {
		t.Error("different action id")
	}

	var commitInfo struct {
		ID      uint64 `json:"id,omitempty"`
		Message string `json:"msg,omitempty"`
	}
	commitInfo.ID = actionID
	commitInfo.Message = "message_01"
	data3 := runtime.JSONEncode(commitInfo)
	param3 := append([]byte{opsCommitAction}, data3...)
	run(hexToBytes(user1), param3, 0)

	action = GetAction(actionID)
	if action.Status != StatusCommit {
		t.Error("error action status")
	}
	if action.Message != "message_01" {
		t.Error("error action message")
	}
	if action.Reward != 0 {
		t.Error("error reward")
	}

	var rewardInfo struct {
		ActionID uint64 `json:"id,omitempty"`
		Reward   uint64 `json:"reward,omitempty"`
	}
	rewardInfo.ActionID = actionID
	rewardInfo.Reward = 1e9
	data4 := runtime.JSONEncode(rewardInfo)
	param4 := append([]byte{opsRewardAction}, data4...)
	run(hexToBytes(owner), param4, 0)

	action = GetAction(actionID)
	if action.Status != StatusClosed {
		t.Error("error action status")
	}
	if action.Message != "message_01" {
		t.Error("error action message")
	}
	if action.Reward != rewardInfo.Reward {
		t.Error("error reward")
	}
	rewDB := core.GetDB(tRewardStatus{})
	aid := rewDB.GetInt(hexToBytes(user1))
	if aid != actionID {
		t.Error("not reward info")
	}

	task := GetTask(taskID)
	if task.Rewarded != rewardInfo.Reward {
		t.Error("error reward1")
	}

	var closeInfo struct {
		ID      uint64 `json:"id,omitempty"`
		Message string `json:"msg,omitempty"`
	}
	closeInfo.ID = taskID
	data5 := runtime.JSONEncode(closeInfo)
	param5 := append([]byte{opsCloseTask}, data5...)
	run(hexToBytes(owner), param5, 0)

	task = GetTask(taskID)
	if task.Rewarded != task.Bounty {
		t.Error("error reward2")
	}
}

func Test_run2(t *testing.T) {
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

	core.SetAppAccountForTest(tApp{}, 1e15)

	// new task
	var cost uint64 = 1e10
	var info Task
	info.Title = "task002"
	info.Number = 10
	info.AcceptRule = core.GetAppName(tApp{})
	info.CommitRule = info.AcceptRule
	info.RewardRule = info.AcceptRule
	data := runtime.JSONEncode(info)
	param := append([]byte{opsNewTask}, data...)
	run(hexToBytes(owner), param, cost)

	// accept task
	userStr := "02984010319cd34659f7fcb20b31d615d850ab32ca930618"
	var user core.Address
	core.Decode(0, hexToBytes(userStr), &user)
	taskID := GetLastIDOfTask()

	ok := AcceptTask(tApp{}, taskID, user)
	if !ok {
		t.Error("error1")
	}

	actionID := GetLastIDOfAction()
	action := GetAction(actionID)
	if action.Status != StatusAccept {
		t.Error("error action status")
	}

	ok = CommitTask(tApp{}, actionID, "msg_bbb")
	if !ok {
		t.Error("error2")
	}

	action = GetAction(actionID)
	if action.Status != StatusCommit {
		t.Error("error action status")
	}
	if action.Message != "msg_bbb" {
		t.Error("error action message")
	}
	if action.Reward != 0 {
		t.Error("error reward")
	}

	var reward uint64 = 1e9
	ok = RewardTask(tApp{}, actionID, reward)
	if !ok {
		t.Error("error2")
	}

	action = GetAction(actionID)
	if action.Status != StatusClosed {
		t.Error("error action status")
	}
	if action.Reward != reward {
		t.Error("error reward")
	}

	task := GetTask(taskID)
	if task.Rewarded != reward {
		t.Error("error reward1")
	}
}
