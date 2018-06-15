package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/go-xorm/xorm"
	"github.com/neo4l/x/tool"

	chain "github.com/neo4l/eth-chain"
)

func Test_Parse(t *testing.T) {
	hash := "0x7b5bf6346af3b77eb84484f06ef1035ff8206971b5aaaa259738efc7e97425de"
	tops := chain.GetTopics("", hash)
	if len(tops) == 4 && tops[0] == "0xa9059cbb00000000000000000000000000000000000000000000000000000000" {
		fmt.Println("tx: " + tops[1] + "," + tops[2] + "," + tops[3])
		from := strings.Replace(tops[1], "0x000000000000000000000000", "0x", 1)
		to := strings.Replace(tops[2], "0x000000000000000000000000", "0x", 1)
		value := tool.HexToIntStr(tops[3])
		//val := "6316120034000000006700"
		value2 := tool.ToBalance(value, 18)
		value3 := tool.ToValue(value2, 18)
		fmt.Println("tx: " + from + ", " + to + ", " + value + ", " + value2 + ", " + value3)
	}
	//fmt.Println(tops)
}

func Test_Sync(t *testing.T) {
	//buildDBConnect()
	pgEngine, err := xorm.NewEngine(DBDriverName, DBURL)
	if err != nil {
		fmt.Printf("connect pgsql db error, %s\n", err)
		return
	}
	//defer pgsqlEngine.Close()
	err = pgEngine.Sync2(new(BTx))
	if err != nil {
		fmt.Printf("connect pgsql db error, %s\n", err)
		return
	}

	//5,468,417
	for index := 5790795 + 1; index < int(getBcBlockNumber())-15; index++ {
		SyncBlock(pgEngine, int64(index))
		//fmt.Println("result: ", b)
		//return
		// asset := "KEY"
		// from := "0x7b74c19124a9ca92c6141a2ed5f92130fc2791f2"
		// to := "0x6ef149e870e4f7c67f98601ddc58461e79d9db0e"
		// value := "32748.669"
		// status := 20
		// blockNum := int64(5644709)
		// hash := "0x8da3c24bca1c4e2c17ad900be64d728e7d825a790b68138358cbe7f7dcece0f2"
		// blocktime := time.Now()
		// tx := BTx{ Asset:asset,  Fromaddr: from, Toaddr: to, Value: value, Status: status, Blocknum: blockNum, Txhash: hash, Blocktime: blocktime, Createtime: time.Now() }
		// pgEngine.InsertOne(tx)
	}
}
