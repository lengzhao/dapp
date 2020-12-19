package zcdc09c276df3af32f8f60c33d57710374fe126f482fa50db61271bac0886e600

import (
	core "github.com/lengzhao/dapp/zff0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"
)

type tApp struct{}
type tTask struct{}
type tAction struct{}
type tTaskStatus struct{}
type tActionStatus struct{}
type tRewardStatus struct{}

// TaskHead input by user
type TaskHead struct {
	Title       string                 `json:"title,omitempty"`
	Life        uint64                 `json:"life,omitempty"`
	Description string                 `json:"desc,omitempty"`
	Number      uint32                 `json:"number,omitempty"`
	AcceptRule  core.Hash              `json:"accept_rule,omitempty"`
	CommitRule  core.Hash              `json:"commit_rule,omitempty"`
	RewardRule  core.Hash              `json:"reward_rule,omitempty"`
	Others      map[string]interface{} `json:"others,omitempty"`
}

// Task task
type Task struct {
	TaskHead
	Owner     core.Address `json:"owner,omitempty"`
	Bounty    uint64       `json:"bounty,omitempty"`
	AcceptNum uint32       `json:"accept_num,omitempty"`
	Status    int          `json:"status,omitempty"`
	Rewarded  uint64       `json:"rewarded,omitempty"`
	Message   string       `json:"message,omitempty"`
}

// Action the accept task
type Action struct {
	TaskID  uint64       `json:"task_id,omitempty"`
	User    core.Address `json:"user,omitempty"`
	Index   uint32       `json:"index,omitempty"`
	Status  int          `json:"status,omitempty"`
	Reward  uint64       `json:"reward,omitempty"`
	Message string       `json:"message,omitempty"`
}

// Status of task
const (
	StatusInit = iota
	StatusAccept
	StatusCommit
	StatusRefuse
	StatusClosed
)

const (
	// opsNewTask net task
	opsNewTask = byte(iota + 1)
	opsAcceptTask
	opsCommitAction
	opsRewardAction
	opsCloseTask
	opsCancel
	opsSetOwner
)

var (
	keyOfTask   = []byte{0, 0}
	keyOfAction = []byte{0, 1}
	keyOfOwner  = []byte{0, 2}
	owner       = "01ccaf415a3a6dc8964bf935a1f40e55654a4243ae99c709"
)

func run(user, in []byte, cost uint64) {
	switch in[0] {
	case opsNewTask:
		if cost < 1e9 {
			panic("request 1e9 cost")
		}
		var info Task
		core.Decode(core.EncJSON, in[1:], &info.TaskHead)
		if info.Title == "" {
			panic("need title")
		}

		info.Bounty = cost
		if info.Number == 0 {
			info.Number = 1e5
		}
		if info.Life == 0 {
			info.Life = core.TimeDay * 7
		}
		core.Decode(0, user, &info.Owner)
		idDB := core.GetDB(tApp{})
		id := idDB.GetInt(keyOfTask) + 1
		idDB.SetInt(keyOfTask, id, core.TimeYear)
		key := core.Encode(0, id)
		db := core.GetDB(tTask{})
		db.Set(key, core.Encode(core.EncJSON, info), info.Life)
		statDB := core.GetDB(tTaskStatus{})
		statDB.SetInt(user, id, 1)
		statKey := append(user, key...)
		statDB.SetInt(statKey, 1, 1)
		core.Event(tTask{}, "new", key)
	case opsAcceptTask:
		var taskID uint64
		core.Decode(0, in[1:], &taskID)
		var task Task
		key := core.Encode(0, taskID)
		db := core.GetDB(tTask{})
		d1, life := db.Get(key)
		core.Decode(core.EncJSON, d1, &task)
		if !task.AcceptRule.Empty() {
			panic("need AcceptRule")
		}
		task.AcceptNum++

		if task.AcceptNum > task.Number {
			panic("over number")
		}
		if task.Status == StatusClosed {
			panic("closed")
		}
		if task.Rewarded >= task.Bounty {
			panic("not more reward")
		}

		db.Set(key, core.Encode(core.EncJSON, task), life)

		idDB := core.GetDB(tApp{})
		db3 := core.GetDB(tAction{})
		id := idDB.GetInt(keyOfAction) + 1
		idDB.SetInt(keyOfAction, id, core.TimeYear)
		key2 := core.Encode(0, id)

		var action Action
		action.Index = task.AcceptNum
		action.Status = StatusAccept
		action.TaskID = taskID
		core.Decode(0, user, &action.User)
		db3.Set(key2, core.Encode(core.EncJSON, action), life)
		key3 := append(key, core.Encode(0, action.Index)...)
		idDB.SetInt(key3, id, core.TimeSecond)
		statDB := core.GetDB(tActionStatus{})
		statDB.SetInt(user, id, 1)
		statKey := append(user, key...)
		if statDB.GetInt(statKey) > 0 {
			panic("accepted")
		}
		statDB.SetInt(statKey, id, core.TimeDay)
		core.Event(tAction{}, "new", key, key2)
	case opsCommitAction:
		var input struct {
			ID      uint64 `json:"id,omitempty"`
			Message string `json:"msg,omitempty"`
		}
		core.Decode(core.EncJSON, in[1:], &input)

		var action Action
		key := core.Encode(0, input.ID)
		db := core.GetDB(tAction{})
		d, l := db.Get(key)
		if len(d) == 0 {
			panic("not found the action")
		}
		core.Decode(core.EncJSON, d, &action)
		if action.Status != StatusAccept {
			panic("error status")
		}
		var u core.Address
		core.Decode(0, user, &u)
		if u != action.User {
			panic("error user")
		}

		var task Task
		k2 := core.Encode(0, action.TaskID)
		db2 := core.GetDB(tTask{})
		d2, _ := db2.Get(k2)
		core.Decode(core.EncJSON, d2, &task)
		if !task.CommitRule.Empty() {
			panic("CommitRule not empty")
		}
		action.Status = StatusCommit
		if input.Message != "" {
			action.Message = input.Message
		}

		db.Set(key, core.Encode(core.EncJSON, action), l)
		core.Event(tAction{}, "commit", key)
	case opsRewardAction:
		var input struct {
			ActionID uint64 `json:"id,omitempty"`
			Reward   uint64 `json:"reward,omitempty"`
		}
		core.Decode(core.EncJSON, in[1:], &input)
		var action Action
		key := core.Encode(0, input.ActionID)
		db := core.GetDB(tAction{})
		d, l := db.Get(key)
		if len(d) == 0 {
			panic("not found the action")
		}
		core.Decode(core.EncJSON, d, &action)
		if action.Status != StatusCommit {
			panic("error status")
		}

		var task Task
		k2 := core.Encode(0, action.TaskID)
		db2 := core.GetDB(tTask{})
		d2, l2 := db2.Get(k2)
		core.Decode(core.EncJSON, d2, &task)
		if !task.RewardRule.Empty() {
			panic("RewardRule not empty")
		}

		if input.Reward+task.Rewarded > task.Bounty ||
			input.Reward > task.Bounty {
			panic("not more reward")
		}
		var u core.Address
		core.Decode(0, user, &u)
		if u != task.Owner {
			panic("error user")
		}
		task.Rewarded += input.Reward
		core.TransferAccounts(tApp{}, action.User, input.Reward)
		db2.Set(k2, core.Encode(core.EncJSON, task), l2)
		action.Status = StatusClosed
		action.Reward = input.Reward
		db.Set(key, core.Encode(core.EncJSON, action), l)
		statKey := append(u[:], key...)
		statDB := core.GetDB(tRewardStatus{})
		statDB.SetInt(action.User[:], input.ActionID, 1)
		statDB.SetInt(statKey, action.TaskID, 1)
		core.Event(tAction{}, "reward", key)
	case opsCloseTask:
		var input struct {
			ID      uint64 `json:"id,omitempty"`
			Message string `json:"msg,omitempty"`
		}
		core.Decode(core.EncJSON, in[1:], &input)

		var task Task
		key := core.Encode(0, input.ID)
		db := core.GetDB(tTask{})
		d1, life := db.Get(key)
		core.Decode(core.EncJSON, d1, &task)
		var u core.Address
		core.Decode(0, user, &u)
		if u.ToHexString() == owner {
			core.Event(tTask{}, "closeByOwner", key)
		} else if u == task.Owner {
			if !task.RewardRule.Empty() {
				panic("exist RewardRule")
			}
		}

		if task.Bounty > task.Rewarded {
			core.TransferAccounts(tApp{}, u, task.Bounty-task.Rewarded)
		}
		task.Rewarded = task.Bounty
		if input.Message != "" {
			task.Message = input.Message
		}
		db.Set(key, core.Encode(core.EncJSON, task), life)
		core.Event(tTask{}, "close", key)
	case opsCancel:
		var taskID uint64
		core.Decode(0, in[1:], &taskID)
		var task Task
		key := core.Encode(0, taskID)
		db := core.GetDB(tTask{})
		d1, life := db.Get(key)
		if life > core.TimeHour*5 {
			panic("error time")
		}
		core.Decode(core.EncJSON, d1, &task)
		if task.Rewarded >= task.Bounty {
			panic("not more reward")
		}
		var u core.Address
		core.Decode(0, user, &u)
		r1 := task.Bounty - task.Rewarded
		r2 := r1 / 100
		core.TransferAccounts(tApp{}, u, r2)
		core.TransferAccounts(tApp{}, task.Owner, r1-r2)
		task.Rewarded = task.Bounty
		task.Message = "cancel"
		db.Set(key, core.Encode(core.EncJSON, task), life)
		core.Event(tTask{}, "cancel", key)
	case opsSetOwner:
		idDB := core.GetDB(tApp{})
		d, _ := idDB.Get(keyOfOwner)
		if len(d) > 0 {
			panic("exist owner")
		}
		var u core.Address
		core.Decode(0, user, &u)
		if u.ToHexString() != owner {
			panic("not owner")
		}
		var app core.Hash
		core.Decode(0, in[1:], &app)
		if core.GetAppInfo(app) == nil {
			panic("not exist app")
		}
		idDB.Set(keyOfOwner, app[:], core.TimeYear)
		core.Event(tApp{}, "setOwner", app[:])
	default:
		panic("not support")
	}
}

// AcceptTask accept task, return true when success
func AcceptTask(caller interface{}, taskID uint64, user core.Address) bool {
	var task Task
	key := core.Encode(0, taskID)
	db := core.GetDB(tTask{})
	d1, life := db.Get(key)
	core.Decode(core.EncJSON, d1, &task)
	if task.AcceptRule != core.GetAppName(caller) {
		return false
	}
	if task.Rewarded >= task.Bounty {
		return false
	}

	task.AcceptNum++
	if task.AcceptNum > task.Number {
		return false
	}
	core.Event(tAction{}, "new", key)
	db.Set(key, core.Encode(core.EncJSON, task), life)

	idDB := core.GetDB(tApp{})
	db3 := core.GetDB(tAction{})
	id := idDB.GetInt(keyOfAction) + 1
	idDB.SetInt(keyOfAction, id, core.TimeYear)
	key2 := core.Encode(0, id)

	var action Action
	action.Index = task.AcceptNum
	action.Status = StatusAccept
	action.TaskID = taskID
	action.User = user
	db3.Set(key2, core.Encode(core.EncJSON, action), life)

	key3 := append(key, core.Encode(0, action.Index)...)
	idDB.SetInt(key3, id, core.TimeSecond)
	statKey := append(user[:], key...)
	statDB := core.GetDB(tActionStatus{})
	if statDB.GetInt(statKey) > 0 {
		panic("accepted")
	}
	statDB.SetInt(user[:], id, 1)
	statDB.SetInt(statKey, id, core.TimeDay)
	return true
}

// CommitTask commit task
func CommitTask(caller interface{}, actionID uint64, msg string) bool {
	var action Action
	key := core.Encode(0, actionID)
	db := core.GetDB(tAction{})
	d, l := db.Get(key)
	if len(d) == 0 {
		return false
	}
	core.Decode(core.EncJSON, d, &action)
	if action.Status != StatusAccept {
		return false
	}

	var task Task
	k2 := core.Encode(0, action.TaskID)
	db2 := core.GetDB(tTask{})
	d2, _ := db2.Get(k2)
	core.Decode(core.EncJSON, d2, &task)
	if task.CommitRule != core.GetAppName(caller) {
		return false
	}
	action.Status = StatusCommit
	if msg != "" {
		action.Message = msg
	}

	db.Set(key, core.Encode(core.EncJSON, action), l)
	core.Event(tAction{}, "commit", key)
	return true
}

// RewardTask reward task
func RewardTask(caller interface{}, actionID uint64, reward uint64) bool {
	var action Action
	key := core.Encode(0, actionID)
	db := core.GetDB(tAction{})
	d, l := db.Get(key)
	if len(d) == 0 {
		return false
	}
	core.Decode(core.EncJSON, d, &action)
	if action.Status != StatusCommit {
		return false
	}

	var task Task
	k2 := core.Encode(0, action.TaskID)
	db2 := core.GetDB(tTask{})
	d2, l2 := db2.Get(k2)
	core.Decode(core.EncJSON, d2, &task)
	if task.RewardRule != core.GetAppName(caller) {
		return false
	}
	if reward+task.Rewarded > task.Bounty || reward > task.Bounty {
		return false
	}
	task.Rewarded += reward
	core.TransferAccounts(tApp{}, action.User, reward)
	db2.Set(k2, core.Encode(core.EncJSON, task), l2)
	action.Status = StatusClosed
	action.Reward = reward
	db.Set(key, core.Encode(core.EncJSON, action), l)
	statKey := append(action.User[:], key...)
	statDB := core.GetDB(tRewardStatus{})
	statDB.SetInt(action.User[:], actionID, 1)
	statDB.SetInt(statKey, action.TaskID, 1)
	core.Event(tAction{}, "reward", key)
	return true
}

// CloseTask close task
func CloseTask(caller interface{}, taskID uint64, msg string) bool {
	var task Task
	key := core.Encode(0, taskID)
	db := core.GetDB(tTask{})
	d, l := db.Get(key)
	if len(d) == 0 {
		return false
	}
	core.Decode(core.EncJSON, d, &task)
	if task.Status == StatusClosed {
		return false
	}
	idDB := core.GetDB(tApp{})
	d, _ = idDB.Get(keyOfOwner)
	if len(d) == 0 {
		return false
	}
	var app core.Hash
	core.Decode(0, d, &app)
	if app != core.GetAppName(caller) {
		return false
	}
	if task.Bounty > task.Rewarded {
		core.TransferAccounts(tApp{}, task.Owner, task.Bounty-task.Rewarded)
	}
	task.Rewarded = task.Bounty
	task.Status = StatusClosed
	if msg != "" {
		task.Message = msg
	}

	db.Set(key, core.Encode(core.EncJSON, task), l)
	core.Event(tTask{}, "close", key)
	return true
}

// SetOwner set owner
func SetOwner(caller interface{}, newOwner core.Hash) bool {
	idDB := core.GetDB(tApp{})
	d, _ := idDB.Get(keyOfOwner)
	if len(d) == 0 {
		return false
	}
	var app core.Hash
	core.Decode(0, d, &app)
	if app != core.GetAppName(caller) {
		return false
	}
	if core.GetAppInfo(newOwner) == nil {
		return false
	}
	idDB.Set(keyOfOwner, newOwner[:], core.TimeYear)
	core.Event(tApp{}, "setOwner", newOwner[:])
	return true
}

// GetTask get task by id
func GetTask(id uint64) Task {
	var out Task
	db := core.GetDB(tTask{})
	d, _ := db.Get(core.Encode(0, id))
	if len(d) > 0 {
		core.Decode(core.EncJSON, d, &out)
	}
	return out
}

// GetLastIDOfTask get last id of task
func GetLastIDOfTask() uint64 {
	idDB := core.GetDB(tApp{})
	return idDB.GetInt(keyOfTask)
}

// GetAction get action by id
func GetAction(id uint64) Action {
	var out Action
	db := core.GetDB(tAction{})
	d, _ := db.Get(core.Encode(0, id))
	if len(d) > 0 {
		core.Decode(core.EncJSON, d, &out)
	}
	return out
}

// GetLastIDOfAction get last id of Action
func GetLastIDOfAction() uint64 {
	idDB := core.GetDB(tApp{})
	return idDB.GetInt(keyOfAction)
}
