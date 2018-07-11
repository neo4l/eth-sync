package main

import (
	"fmt"
	"testing"

	"github.com/go-xorm/xorm"
	"github.com/neo4l/x/tool"

	chain "github.com/neo4l/eth-chain"
)

func Test_Parse(t *testing.T) {
	hash := "0x51beaf8fd1fcdcd50dfab37d405454acfafd4e477d5105a1217af5eac2b4c604"
	reply, _ := chain.GetTransactionReceipt(BlockChainHost, hash)
	fmt.Println(reply)
	tops := chain.ParseERC20Tx(BlockChainHost, hash)
	if len(tops) == 4 {
		fmt.Println("tx: " + tops[0] + "," + tops[1] + "," + tops[2] + "," + tool.ToBalance(tops[3], 18))
	}

	block, _ := chain.GetBlock(BlockChainHost, tool.IntToHex(5943396), true)
	fmt.Println(block)
}

func Test_Sync(t *testing.T) {
	fmt.Printf("Test_Sync")
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

	for index := 5942666 + 1; index < int(getBcBlockNumber())-15; index++ {
		SyncBlock(pgEngine, int64(index))
	}

}
