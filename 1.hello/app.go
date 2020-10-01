package a1000000000000000000000000000010203040506070809010203040506070801

// import dApp on chain
import core "github.com/lengzhao/dapp/zff0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"

type tApp struct {
}

func run(user, in []byte, cost uint64) {
	// write a log
	core.Event(tApp{}, "start_app", user, in)

	// get db by private struct(lowercase start)
	db := core.GetDB(tApp{})
	// write data to db
	db.Set([]byte("hello"), user, core.TimeMonth)
	// get data from db
	val, life := db.Get([]byte("hello"))
	if life < core.TimeHour {
		core.Event(tApp{}, "The data is about to expire")
	}
	core.Event(tApp{}, "value", val)
}
