package main

import (
	"log"
	"strconv"
	"time"

	chain "github.com/neo4l/eth-chain"
	"github.com/neo4l/x/redis"
	"github.com/neo4l/x/tool"

	_ "database/sql"

	"github.com/go-xorm/xorm"
	_ "github.com/lib/pq"

	"fmt"
)

var (
	syncedTxCount, totalTxCount int64
	DefaultStartSyncBlockNumber int64 = 5942666

	BlockChainHost             = "http://127.0.0.1:8545"
	RedisSynchronizedBlockFlag = "bc:synchronizedBlockNumber"

	IsSaveToDB = true

	DBDriverName = "postgres" //mysql
	DBURL        = "host= port= user= password= dbname= sslmode=disable"
	RedisHost    = ""
	RedisPasswd  = ""

	tokenMap    map[string]string
	pgsqlEngine *xorm.Engine
	redisClient *redis.RedisClient
)

func init() {
	tokenMap = make(map[string]string)
	tokenMap["0x4cd988afbad37289baaf53c13e98e2bd46aaea8c"] = "KEY"
	tokenMap["0xb9bb08ab7e9fa0a1356bd4a39ec0ca267e03b0b3"] = "PAI"
	tokenMap["0x0f1ed66c251bcb52ecf7e67ac64bb72482048adb"] = "SER"
}

func main() {

	log.Printf("Sync Run err: %s", Run())
}

func Try(fun func(), handler func(interface{})) {
	defer func() {
		if err := recover(); err != nil {
			handler(err)
		}
	}()
	fun()
}
func Run() error {
	log.Println("Method: sync job run....")
	buildDBConnect()

	//redisClient.Set(RedisSynchronizedBlockFlag, strconv.FormatInt(DefaultStartSyncBlockNumber, 10))

	for true {
		fmt.Printf("---------------------------start exec sync job: %s---------------------------\n", time.Now().Format("2006-01-02 15:04:05"))
		Try(func() {
			SyncData(pgsqlEngine)
		}, func(e interface{}) {
			//print(e)
			fmt.Println("run exception: ", e)
		})
		fmt.Printf("---------------------------end exec sync job: %s---------------------------\n", time.Now().Format("2006-01-02 15:04:05"))
		time.Sleep(10 * time.Second)
	}
	return nil
}

func buildDBConnect() error {

	closeDBConnect()

	pgEngine, err := xorm.NewEngine(DBDriverName, DBURL)
	if err != nil {
		fmt.Printf("connect pgsql db error\n")
		return err
	}
	//defer pgsqlEngine.Close()
	err = pgEngine.Sync2(new(BTx))
	if err != nil {
		fmt.Printf("connect pgsql db error\n")
		return err
	}
	pgsqlEngine = pgEngine
	fmt.Println("build db connect successful")

	redisClient = redis.NewClient(RedisHost, RedisPasswd, 0)
	return nil
}

func closeDBConnect() {
	fmt.Println("method: closeDBConnect...")

	if pgsqlEngine != nil {
		pgsqlEngine.Close()
		pgsqlEngine = nil
	}
	if redisClient != nil {
		redisClient.Close()
		redisClient = nil
	}
}

func SyncData(pgEngine *xorm.Engine) {
	//log.Println("Method: syncData....")

	bcBlockNumber := getBcBlockNumber()
	procBlockNumber, err := getProcBlockNumber()

	log.Printf("Sync block: start %d/%d (%d/%d).............", procBlockNumber, bcBlockNumber, syncedTxCount, totalTxCount)

	if bcBlockNumber == 0 || err != nil || procBlockNumber >= bcBlockNumber-15 {
		return
	}

	for index := procBlockNumber; index < bcBlockNumber-15; index++ {
		if SyncBlock(pgEngine, index) {
			//redis.Set("bc:synchronizedBlockNumber", strconv.FormatInt(9750, 10))
			_, err := redisClient.Set(RedisSynchronizedBlockFlag, strconv.FormatInt(index, 10))
			if err != nil {
				break
			}
			log.Printf("Sync block: success %d/%d .............", index, bcBlockNumber)
		} else {
			log.Printf("Sync block: fail %d/%d .............", index, bcBlockNumber)
			break
		}
	}
}

func SyncBlock(engine *xorm.Engine, blockNumber int64) bool {
	log.Printf("Method: SyncBlock %d....", blockNumber)
	blockData, err := chain.GetBlock(BlockChainHost, tool.IntToHex(blockNumber), true)
	if err != nil {
		log.Printf("GetBlock: %s", err)
		return false
	}
	if clearTx(engine, blockNumber) != nil {
		buildDBConnect()
		return false
	}
	var txArray = make([]BTx, 0)
	txCount := len(blockData.Transactions)
	if txCount > 0 {
		//log.Printf("BlockObject: %d %s, %s", len(blockData.Transactions), blockData, err)

		timestamp := time.Unix(tool.HexToIntWithoutError(blockData.Timestamp), 0)
		for i := 0; i < txCount; i++ {
			tx, ok := blockData.Transactions[i].(map[string]interface{})
			if !ok {
				return false
			}
			txObj := ParseTx(tx, timestamp)
			if txObj != nil {
				txArray = append(txArray, *txObj)
			}
		}

		if len(txArray) == 0 {
			return true
		}
		totalTxCount = totalTxCount + int64(len(txArray))
		if IsSaveToDB {
			cnt, err := engine.Insert(txArray)

			syncedTxCount = syncedTxCount + cnt
			log.Printf("SaveTx: %d, %s", len(txArray), err)

			return err == nil && cnt == int64(len(txArray))
		}
	}
	return true
}

func ParseTx(txData map[string]interface{}, blocktime time.Time) *BTx {
	//log.Printf("GetBcTx: %s", txData)
	hash := txData["hash"].(string)
	tx := chain.ParseERC20Tx(BlockChainHost, hash)
	if len(tx) == 4 {
		asset := tokenMap[tx[0]]
		if asset == "" {
			return nil
		}
		value := tool.ToBalance(tx[3], 18)
		blockNum := tool.HexToIntWithoutError(txData["blockNumber"].(string))
		fmt.Println("tx: ", asset, tx[1], tx[2], value, blockNum, hash)
		return &BTx{Asset: asset, Fromaddr: tx[1], Toaddr: tx[2], Value: value, Status: 20, Blocknum: blockNum, Txhash: hash, Blocktime: blocktime, Createtime: time.Now()}
	}
	//fmt.Println("ParseERC20Tx fail")
	return nil
}

func ParseTx2(txData map[string]interface{}, blocktime time.Time) *BTx {
	//log.Printf("GetBcTx: %s", txData)
	if txData["to"] == nil {
		return nil
	}
	to := txData["to"].(string)
	asset := tokenMap[to]
	if to != "" && asset != "" {
		//log.Printf("GetBcTx: %s", tool.ToJson(txData))
		hash := txData["hash"].(string)
		blockNum := tool.HexToIntWithoutError(txData["blockNumber"].(string))
		tx := chain.ParseERC20Tx(BlockChainHost, hash)
		if len(tx) == 4 {
			value := tool.ToBalance(tx[3], 18)
			fmt.Println("tx: ", asset, tx[1], tx[2], value, blockNum, hash)
			return &BTx{Asset: asset, Fromaddr: tx[1], Toaddr: tx[2], Value: value, Status: 20, Blocknum: blockNum, Txhash: hash, Blocktime: blocktime, Createtime: time.Now()}
		}
		fmt.Println("ParseERC20Tx fail")
	}
	return nil
}

func getBcBlockNumber() int64 {
	bcBlockNumberStr, err := chain.GetLatestBlockNumber(BlockChainHost)
	if err != nil {
		log.Printf("Can't find current block number from chain, %s", err)
		return 0
	}
	bcBlockNumber, err := tool.HexToInt(bcBlockNumberStr)
	if err != nil {
		log.Printf("Can't find current block number from chain, %s", err)
		return 0
	}
	return bcBlockNumber
}

func getProcBlockNumber() (int64, error) {

	dbBlockNumber, err := redisClient.Get(RedisSynchronizedBlockFlag)
	if err != nil {
		log.Printf("Get synchronizedBlockNumber error1, %s", err)
		return 0, err
	}

	var procBlockNumber int64
	if tool.IsEmpty(dbBlockNumber) {
		redisClient.Set(RedisSynchronizedBlockFlag, strconv.FormatInt(DefaultStartSyncBlockNumber, 10))
		procBlockNumber = DefaultStartSyncBlockNumber
	} else {
		synchronizedBlockNumber := tool.AToInt64WithoutErr(dbBlockNumber)
		procBlockNumber = synchronizedBlockNumber + 1
	}
	return procBlockNumber, nil
}

type BTx struct {
	Id         int64     `xorm:"BIGINT"`
	Asset      string    `xorm:"VARCHAR(20)"`
	Fromaddr   string    `xorm:"VARCHAR(66)"`
	Toaddr     string    `xorm:"VARCHAR(66)"`
	Value      string    `xorm:"NUMERIC"`
	Status     int       `xorm:"default 0 INTEGER"`
	Blocknum   int64     `xorm:"default 0 BIGINT"`
	Txhash     string    `xorm:"unique VARCHAR(66)"`
	Ext1       string    `xorm:"VARCHAR(200)"`
	Blocktime  time.Time `xorm:"DATETIME"`
	Createtime time.Time `xorm:"DATETIME"`
	Updatetime time.Time `xorm:"DATETIME"`
}

func clearTx(engine *xorm.Engine, blockNum int64) error {
	results, err := engine.Query("select id from b_tx where blocknum = ?", blockNum)
	//log.Printf("ClearTx: Query %d, %s", len(results), err)
	if err == nil && len(results) > 0 {
		affected, err := engine.Exec("delete from b_tx where blocknum = ?", blockNum)
		log.Printf("ClearTx: %d, %s, %s", blockNum, affected, err)
		return err
	}
	return err
}
