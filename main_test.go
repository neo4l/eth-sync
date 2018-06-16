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

	for index := 5790795 + 1; index < int(getBcBlockNumber())-15; index++ {
		SyncBlock(pgEngine, int64(index))

	}
}
