# dApp

## Description

1. These are govm's dApp samples
2. Introduce how to write and debug dApp

## how to run

1. (Optional)start database server<https://github.com/lengzhao/database>
   1. download project
   2. download the data of database<http://govm.net/dl/database_*.zip>
   3. build and start
2. start local server
   1. go build server/server.go
   2. ./server.go
3. run app
   1. cd 1.hello
   2. go test
4. Deploy the dApp
   1. win_wallet or html_wallet
   2. app->New app->input the path of code->submit

## limit

1. package、import: All imported modules can only be modules on the chain, and reference to external modules is not supported.
2. ~~go/select~~: Since this operation will be concurrent and cause data inconsistency, it is temporarily not supported.
3. ~~range~~: Since the rang of map is random, it will cause uncertainty in the execution order, so it is temporarily not supported.
4. ~~cap/recover~~: This function is not necessary, it may cause discrepancies.
5. var & const: Declaration of variables and constants.
6. func: Used to define functions and methods.
7. return: Used to return from a function.
8. panic: Used to exit the app abnormally.All operations will be rolled back.
9. interface: Used to define the interface.
10. struct: Used to define abstract data types.
11. type: Used to declare custom types.
12. map: Built-in associated data types.
13. case、continue、for、fallthrough、else、if、switch、goto、default: Process control.
