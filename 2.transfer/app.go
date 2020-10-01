package a1000000000000000000000000000010203040506070809010203040506070809

// import dApp on chain
import core "github.com/lengzhao/dapp/zff0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"

type tApp struct {
}

func run(user, in []byte, cost uint64) {
	// write a log
	core.Event(tApp{}, "start_app", user, in)

	addr := core.Address{}
	// decode user to Address
	core.Decode(core.EncBinary, user, &addr)
	// transfer from dApp to user 10govm
	core.TransferAccounts(tApp{}, addr, 1e10)
}
