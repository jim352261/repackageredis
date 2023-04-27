package redis

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	redsynclib "github.com/go-redsync/redsync/v4/redis/goredis/v8"
)

// Cacher 先構建一個Cacher實例，然後將配置參數傳入該實例的StartAndGC方法來初始化實例和程序進程退出後的清理工作。
type Cacher struct {
	syncRedis *redsync.Redsync
	pool      *redis.Client
	prefix    string
	Log       *log.Logger
	ctx       ContextTraceInfo
}

// ContextTraceInfo context 用的struct
type ContextTraceInfo struct {
	Context context.Context
	Field   string
}

// Options redis配置參數
type Options struct {
	Addr         string // redis服務的地址，默認為 127.0.0.1:6379
	Password     string // redis鑒權密碼
	Db           int    // 數據庫
	Debug        bool
	MaxRetries   int    // 放棄前會重試幾次
	PoolSize     int    // 池子大小
	MaxActive    int    // 最大活動連接數，值為0時表示不限制 (預計拿掉)
	MaxIdle      int    // 最大空閑連接數 (預計拿掉)
	MinIdle      int    // 最小空閒連接數
	MaxConnAge   int    // redis連接的最大存活時間 默認不會關閉過期連結
	DialTimeout  int    // redis連接的超時時間，超過該時間則關閉連接。單位為秒。默認值是3秒。
	IdleTimeout  int    // 空閑連接的超時時間，超過該時間則關閉連接。單位為秒。默認值是5分鐘。值為0時表示不關閉空閑連接。此值應該總是大於redis服務的超時時間。
	PoolTimeout  int    // 連接池的超時時間，超過該時間則關閉連接。單位為秒。默認值是4秒。值為0時表示不關閉空閑連接。此值應該總是大於redis服務的超時時間。
	ReadTimeout  int    // socket read timeout 超過時間會導致指令失敗(ex. BLPOP超過秒數) 預設三秒
	WriteTimeout int    // socket read timeout 超過時間會導致指令失敗 預設為ReadTimeout
	Prefix       string // 鍵名前綴
	Wait         bool   // 取不到連線池時是否等待
	Log          *log.Logger
}

// New 根據配置參數創建redis工具實例
func New(options Options) (r *Cacher, err error) {
	r = &Cacher{}
	err = r.StartAndGC(options)
	if err != nil {
		err = fmt.Errorf("start and GC error: %s", err)
		return nil, err
	}
	reply, err := r.Do("PING").Value()
	if err != nil {
		err = fmt.Errorf("PONG error: %s", err)
		return nil, err
	}
	if reply != `PONG` {
		err = fmt.Errorf("PONG error: %s", err)
		return nil, err
	}
	return r, err
}

// StartAndGC 使用 Options 初始化redis，並在程序進程退出時關閉連接池。
func (c *Cacher) StartAndGC(options interface{}) error {
	switch opts := options.(type) {
	case Options:
		if opts.Addr == "" {
			return errors.New("miss Addr")
		}
		redisOption := &redis.Options{
			Addr: opts.Addr,
			DB:   opts.Db,
		}
		if opts.Password != "" {
			redisOption.Password = opts.Password
		}
		if opts.MaxRetries != 0 {
			redisOption.MaxRetries = opts.MaxRetries
		}
		if opts.MaxConnAge != 0 {
			redisOption.MaxConnAge = (time.Duration(opts.MaxConnAge) * time.Second)
		}
		if opts.DialTimeout != 0 {
			redisOption.DialTimeout = (time.Duration(opts.DialTimeout) * time.Second)
		}
		if opts.IdleTimeout != 0 {
			redisOption.IdleTimeout = (time.Duration(opts.IdleTimeout) * time.Second)
		}
		if opts.PoolSize != 0 {
			redisOption.PoolSize = opts.PoolSize
		}
		if opts.MinIdle != 0 {
			redisOption.MinIdleConns = opts.MinIdle
		}
		if opts.ReadTimeout != 0 {
			redisOption.ReadTimeout = (time.Duration(opts.ReadTimeout) * time.Second)
		}
		if opts.WriteTimeout != 0 {
			redisOption.WriteTimeout = (time.Duration(opts.WriteTimeout) * time.Second)
		}
		client := redis.NewClient(redisOption)

		syncPool := redsynclib.NewPool(client)
		rs := redsync.New(syncPool)
		c.syncRedis = rs

		// pool := &redis.Pool{
		// 	MaxActive:   opts.MaxActive,
		// 	MaxIdle:     opts.MaxIdle,
		// 	IdleTimeout: time.Duration(opts.IdleTimeout) * time.Second,

		// 	Dial: func() (redis.Conn, error) {
		// 		conn, err := redis.Dial("tcp", opts.Addr)
		// 		if err != nil {
		// 			return nil, err
		// 		}
		// 		if opts.Password != "" {
		// 			if _, err := conn.Do("AUTH", opts.Password); err != nil {
		// 				conn.Close()
		// 				return nil, err
		// 			}
		// 		}
		// 		if _, err := conn.Do("SELECT", opts.Db); err != nil {
		// 			conn.Close()
		// 			return nil, err
		// 		}

		// 		if opts.Debug {
		// 			if opts.Log == nil {
		// 				conn.Close()
		// 				return nil, fmt.Errorf(`empty debug log object`)
		// 			}
		// 			conn = redis.NewLoggingConn(conn, opts.Log, "redis")
		// 		}

		// 		return conn, err
		// 	},

		// 	TestOnBorrow: func(conn redis.Conn, t time.Time) error {
		// 		if time.Since(t) < (time.Second * 30) {
		// 			return nil
		// 		}
		// 		_, err := conn.Do("PING")
		// 		return err
		// 	},
		// 	Wait: opts.Wait,
		// }
		c.prefix = opts.Prefix
		c.pool = client

		c.Log = opts.Log

		return nil
	default:
		return errors.New("unsupported options")
	}
}

func (c *Cacher) clone() *Cacher {
	clone := *c
	return &clone
}

// GracefulStop GracefulStop
func (c *Cacher) GracefulStop() {
	c.pool.Close()
}

// WithContext 添加context 進去
func (c *Cacher) WithContext(ctx context.Context, field string) *Cacher {
	if ctx == nil {
		panic("nil context")
	}
	clone := c.clone()
	clone.ctx = ContextTraceInfo{
		Context: ctx,
		Field:   field,
	}

	return clone
}

// func newTracingConn(ctx ContextTraceInfo, c redis.Conn) redis.Conn {
// 	return c
// }

// Do 執行redis命令並返回結果。執行時從連接池獲取連接並在執行完命令後關閉連接。
func (c *Cacher) Do(commandName string, args ...interface{}) *Cmd {
	cmd := &Cmd{}

	// conn := newTracingConn(c.Context, c.pool.Get())
	// conn := c.pool.Get()
	// defer conn.Close()

	if c.ctx.Context != nil {
		traceID := c.ctx.Context.Value(c.ctx.Field)
		log.Println("[Context] ", traceID, commandName, args)
	}
	argsNew := make([]interface{}, 1+len(args))
	argsNew[0] = commandName
	copy(argsNew[1:], args)
	contextDefault := context.Background()
	if c.ctx.Context != nil {
		contextDefault = c.ctx.Context
	}
	goRedisCmd := c.pool.Do(contextDefault, argsNew...)
	cmd.cmd = goRedisCmd
	cmd.val = goRedisCmd.Val()
	cmd.Err = goRedisCmd.Err()
	if cmd.Err != nil && cmd.Err.Error() == "redis: nil" {
		redigoNilErr := ErrNil
		cmd.Err = redigoNilErr
	}
	// cmd.val, cmd.Err = conn.Do(commandName, args...)

	return cmd
}

// Get 獲取鍵值。一般不直接使用該值，而是配合下面的工具類方法獲取具體類型的值，或者直接使用github.com/gomodule/redigo/redis包的工具方法。
func (c *Cacher) Get(key string) *Cmd {
	return c.Do("GET", c.getKey(key))
}

// Set 存並設置有效時長。時長的單位為秒。
// 基礎類型直接保存，其他用json.Marshal後轉成string保存。
func (c *Cacher) Set(key string, val interface{}, expire int64) *Cmd {
	value, err := c.encode(val)
	if err != nil {
		return &Cmd{
			Err: err,
		}
	}

	if expire > 0 {
		return c.Do("SETEX", c.getKey(key), expire, value)
	}
	return c.Do("SET", c.getKey(key), value)
}

// Expire  將該key設定expire時間
func (c *Cacher) Expire(key string, expire int64) *Cmd {
	return c.Do("EXPIRE", c.getKey(key), expire)
}

func (c *Cacher) ExpireAt(key string, expireAt int64) *Cmd {
	return c.Do("EXPIREAT", c.getKey(key), expireAt)
}

// TTL 搜尋該key expire時間
func (c *Cacher) TTL(key string) *Cmd {
	return c.Do("TTL", c.getKey(key))
}

// Keys 搜尋keys
func (c *Cacher) Keys(key string) *Cmd {
	return c.Do("KEYS", c.getKey(key))
}

// Del 刪除鍵
func (c *Cacher) Del(key string) *Cmd {
	return c.Do("DEL", c.getKey(key))
}

// IncrBy 將 key 所儲存的值加上給定的增量值（increment）。
func (c *Cacher) IncrBy(key string, amount int64) *Cmd {
	return c.Do("INCRBY", c.getKey(key), amount)
}

// DecrBy key 所儲存的值減去給定的減量值（decrement）。
func (c *Cacher) DecrBy(key string, amount int64) *Cmd {
	return c.Do("DECRBY", c.getKey(key), amount)
}

// SetNX 設定NX
func (c *Cacher) SetNX(key string, val interface{}, expire int64) *Cmd {
	value, err := c.encode(val)
	if err != nil {
		return &Cmd{
			Err: err,
		}
	}
	cmd := &Cmd{}
	cmd = c.Do("SETNX", c.getKey(key), value)
	if cmd.Err != nil {
		return cmd
	}
	exists, _ := cmd.Bool()
	if exists && expire > 0 {
		c.Do("EXPIRE", c.getKey(key), int64(expire))
	}

	return cmd
}

// HMSet 將一個map存到Redis hash，同時設置有效期，單位：秒
// Example:
//
// ```golang
// m := make(map[string]interface{})
// m["name"] = "corel"
// m["age"] = 23
// err := c.HMSet("user", m, 10)
// ```
func (c *Cacher) HMSet(key string, expire int, val ...interface{}) *Cmd {
	//直接使用HSet
	// cmd := &Cmd{}
	// conn := c.pool.Get()
	// defer conn.Close()
	// cmd.Err = conn.Send("HMSET", redis.Args{}.Add(c.getKey(key)).AddFlat(val)...)
	// if cmd.Err != nil {
	// 	return cmd
	// }
	// if expire > 0 {
	// 	cmd.Err = conn.Send("EXPIRE", c.getKey(key), int64(expire))
	// }
	// if cmd.Err != nil {
	// 	return cmd
	// }
	// conn.Flush()
	// cmd.val, cmd.Err = conn.Receive()
	cmd := c.HSet(key, val...)
	if expire != 0 {
		cmd = c.Do("EXPIRE", key, expire)
	}

	return cmd
}

/** Redis hash 是一個string類型的field和value的映射表，hash特別適合用於存儲對象。 **/

// HSet 將哈希表 key 中的字段 field 的值設為 val
// Example:
//
// ```golang
// _, err := c.HSet("user", "age", 23)
// ```
func (c *Cacher) HSet(key string, val ...interface{}) *Cmd {
	// value, err := c.encode(val)
	// if err != nil {
	// 	return &Cmd{
	// 		Err: err,
	// 	}
	// }
	args := make([]interface{}, 1, 1+len(val))
	args[0] = c.getKey(key)
	args = appendArgs(args, val)

	return c.Do("HSET", args...)
}

// HGet 獲取存儲在哈希表中指定字段的值
// Example:
//
// ```golang
// val, err := c.HGet("user", "age")
// ```
func (c *Cacher) HGet(key, field string) *Cmd {
	return c.Do("HGET", c.getKey(key), field)
}

// HGetAll HGetAll("key", &val)
func (c *Cacher) HGetAll(key string) *Cmd {

	return c.Do("HGETALL", c.getKey(key))
}

// HDel , HDEL KEY_NAME FIELD1.. FIELDN
func (c *Cacher) HDel(key string, fileds ...interface{}) *Cmd {
	args := make([]interface{}, 1+len(fileds))
	args[0] = key
	copy(args[1:], fileds)

	return c.Do("HDEL", args...)
}

// HExists , 確認該欄位是否存在
func (c *Cacher) HExists(key, field string) *Cmd {
	return c.Do("HEXISTS", c.getKey(key), field)
}

// HLen , 確認該hash的長度
func (c *Cacher) HLen(key string) *Cmd {
	return c.Do("HLEN", c.getKey(key))
}

// HKeys , 確認該hash內的所有fields名稱
func (c *Cacher) HKeys(key string) *Cmd {
	return c.Do("HKEYS", c.getKey(key))
}

// HIncrby , 對該hash內的field進行加值
func (c *Cacher) HIncrby(key, field string, number int) *Cmd {
	return c.Do("HINCRBY", c.getKey(key), field, number)
}

// HSetNX , 為hash新增 field ，如果已存在 則新增失敗
func (c *Cacher) HSetNX(key, field string, value interface{}) *Cmd {
	return c.Do("HSETNX", c.getKey(key), field, value)
}

/**
Redis列表是簡單的字符串列表，按照插入順序排序。你可以添加一個元素到列表的頭部（左邊）或者尾部（右邊）
**/

// BLPop 它是 LPOP 命令的阻塞版本，當給定列表內沒有任何元素可供彈出的時候，連接將被 BLPOP 命令阻塞，直到等待超時或發現可彈出元素為止。
// 超時參數 timeout 接受一個以秒為單位的數字作為值。超時參數設為 0 表示阻塞時間可以無限期延長(block indefinitely) 。
func (c *Cacher) BLPop(key string, timeout int) *Cmd {
	values, err := c.Do("BLPOP", c.getKey(key), timeout).Values()
	if err != nil {
		return &Cmd{
			Err: err,
		}
	}
	if len(values) != 2 {
		return &Cmd{
			Err: fmt.Errorf("redigo: unexpected number of values, got %d", len(values)),
		}
	}

	return &Cmd{
		val: values[1],
	}
}

// BRPop 它是 RPOP 命令的阻塞版本，當給定列表內沒有任何元素可供彈出的時候，連接將被 BRPOP 命令阻塞，直到等待超時或發現可彈出元素為止。
// 超時參數 timeout 接受一個以秒為單位的數字作為值。超時參數設為 0 表示阻塞時間可以無限期延長(block indefinitely) 。
func (c *Cacher) BRPop(key string, timeout int) *Cmd {
	values, err := c.Do("BRPOP", c.getKey(key), timeout).Values()
	if err != nil {
		return &Cmd{
			Err: err,
		}
	}
	if len(values) != 2 {
		return &Cmd{
			Err: fmt.Errorf("redigo: unexpected number of values, got %d", len(values)),
		}
	}

	return &Cmd{
		val: values[1],
	}
}

// LPop 移出並獲取列表中的第一個元素（表頭，左邊）
func (c *Cacher) LPop(key string) *Cmd {
	return c.Do("LPOP", c.getKey(key))
}

// RPop 移出並獲取列表中的最後一個元素（表尾，右邊）
func (c *Cacher) RPop(key string) *Cmd {
	return c.Do("RPOP", c.getKey(key))
}

// LPush 將一個值插入到列表頭部
func (c *Cacher) LPush(key string, member ...interface{}) *Cmd {

	memberSlice := make([]interface{}, 0)
	for _, memberv := range member {
		value, err := c.encode(memberv)
		if err != nil {
			return &Cmd{
				Err: err,
			}
		}
		memberSlice = append(memberSlice, value)
	}
	argsNew := make([]interface{}, 1+len(member))
	argsNew[0] = c.getKey(key)
	copy(argsNew[1:], memberSlice)
	return c.Do("LPUSH", argsNew...)
}

// LTrim 將列表進行修剪 只保留指定的區間
func (c *Cacher) LTrim(key string, start, stop int32) *Cmd {

	return c.Do("LTRIM", c.getKey(key), start, stop)
}

// RPush 將一個值插入到列表尾部
func (c *Cacher) RPush(key string, member ...interface{}) *Cmd {
	memberSlice := make([]interface{}, 0)
	for _, memberv := range member {
		value, err := c.encode(memberv)
		if err != nil {
			return &Cmd{
				Err: err,
			}
		}
		memberSlice = append(memberSlice, value)
	}
	argsNew := make([]interface{}, 1+len(member))
	argsNew[0] = c.getKey(key)
	copy(argsNew[1:], memberSlice)

	return c.Do("RPUSH", argsNew...)
}

// LLen 獲取列表的長度
func (c *Cacher) LLen(key string) *Cmd {
	return c.Do("LLEN", c.getKey(key))
}

// LRange 返回列表 key 中指定區間內的元素，區間以偏移量 start 和 stop 指定。
// 下標(index)參數 start 和 stop 都以 0 為底，也就是說，以 0 表示列表的第一個元素，以 1 表示列表的第二個元素，以此類推。
// 你也可以使用負數下標，以 -1 表示列表的最後一個元素， -2 表示列表的倒數第二個元素，以此類推。
// 和編程語言區間函數的區別：end 下標也在 LRANGE 命令的取值範圍之內(閉區間)。
func (c *Cacher) LRange(key string, start, end int) *Cmd {
	return c.Do("LRANGE", c.getKey(key), start, end)
}

/**
Redis 有序集合和集合一樣也是string類型元素的集合,且不允許重覆的成員。
不同的是每個元素都會關聯一個double類型的分數。redis正是通過分數來為集合中的成員進行從小到大的排序。
有序集合的成員是唯一的,但分數(score)卻可以重覆。
集合是通過哈希表實現的，所以添加，刪除，查找的覆雜度都是O(1)。
**/

// ZAdd 將一個 member 元素及其 score 值加入到有序集 key 當中。
func (c *Cacher) ZAdd(key string, score int64, member string) *Cmd {
	return c.Do("ZADD", c.getKey(key), score, member)
}

// ZRem 移除有序集 key 中的一個成員，不存在的成員將被忽略。
func (c *Cacher) ZRem(key string, member string) *Cmd {
	return c.Do("ZREM", c.getKey(key), member)
}

// ZScore 返回有序集 key 中，成員 member 的 score 值。 如果 member 元素不是有序集 key 的成員，或 key 不存在，返回 nil 。
func (c *Cacher) ZScore(key string, member string) *Cmd {
	return c.Do("ZSCORE", c.getKey(key), member)
}

// ZRank 返回有序集中指定成員的排名。其中有序集成員按分數值遞增(從小到大)順序排列。score 值最小的成員排名為 0
func (c *Cacher) ZRank(key, member string) *Cmd {
	return c.Do("ZRANK", c.getKey(key), member)
}

// ZRevrank 返回有序集中成員的排名。其中有序集成員按分數值遞減(從大到小)排序。分數值最大的成員排名為 0 。
func (c *Cacher) ZRevrank(key, member string) *Cmd {
	return c.Do("ZREVRANK", c.getKey(key), member)
}

// ZRange 返回有序集中，指定區間內的成員。其中成員的位置按分數值遞增(從小到大)來排序。具有相同分數值的成員按字典序(lexicographical order )來排列。
// 以 0 表示有序集第一個成員，以 1 表示有序集第二個成員，以此類推。或 以 -1 表示最後一個成員， -2 表示倒數第二個成員，以此類推。
func (c *Cacher) ZRange(key string, from, to int64) *Cmd {
	return c.Do("ZRANGE", c.getKey(key), from, to)
}

func (c *Cacher) ZRangeWithScore(key string, from, to int64) *Cmd {
	return c.Do("ZRANGE", c.getKey(key), from, to, "WITHSCORES")
}

// ZRevrange 返回有序集中，指定區間內的成員。其中成員的位置按分數值遞減(從大到小)來排列。具有相同分數值的成員按字典序(lexicographical order )來排列。
// 以 0 表示有序集第一個成員，以 1 表示有序集第二個成員，以此類推。或 以 -1 表示最後一個成員， -2 表示倒數第二個成員，以此類推。
func (c *Cacher) ZRevrange(key string, from, to int64) *Cmd {
	return c.Do("ZREVRANGE", c.getKey(key), from, to, "WITHSCORES")
}

// ZRangeByScore 返回有序集合中指定分數區間的成員列表。有序集成員按分數值遞增(從小到大)次序排列。
// 具有相同分數值的成員按字典序來排列
func (c *Cacher) ZRangeByScore(key string, from, to, offset int64, count int) *Cmd {
	return c.Do("ZRANGEBYSCORE", c.getKey(key), from, to, "WITHSCORES", "LIMIT", offset, count)
}

// ZRevrangeByScore 返回有序集中指定分數區間內的所有的成員。有序集成員按分數值遞減(從大到小)的次序排列。
// 具有相同分數值的成員按字典序來排列
func (c *Cacher) ZRevrangeByScore(key string, from, to, offset int64, count int) *Cmd {
	return c.Do("ZREVRANGEBYSCORE", c.getKey(key), from, to, "WITHSCORES", "LIMIT", offset, count)
}

// ZCard 返回有序集合中的成員數
// 具有相同分數值的成員按字典序來排列
func (c *Cacher) ZCard(key string) *Cmd {
	return c.Do("ZCARD", c.getKey(key))
}

// SAdd 將一個 member 元素加入到集合 key 當中，已經存在集合內的將被忽略。
func (c *Cacher) SAdd(key, member string) *Cmd {
	return c.Do("SADD", c.getKey(key), member)
}

// SRem 移除集合中的成員
func (c *Cacher) SRem(key, member string) *Cmd {
	return c.Do("SREM", c.getKey(key), member)
}

// SCard 返回集合中的成員數。
func (c *Cacher) SCard(key string) *Cmd {
	return c.Do("SCARD", c.getKey(key))
}

// SPop 返回集合中指定分數區間內的所有的成員。
func (c *Cacher) SPop(key string, count int) *Cmd {
	return c.Do("SPOP", c.getKey(key), count)
}

// SISMember 確認該成員是否在該集合內
func (c *Cacher) SisMembers(key, member string) *Cmd {
	return c.Do("SISMEMBER", c.getKey(key), member)
}

// SMembers 返回集合內的所有的成員
func (c *Cacher) SMembers(key string) *Cmd {
	return c.Do("SMEMBERS", c.getKey(key))
}

/**
Redis 發布訂閱(pub/sub)是一種消息通信模式：發送者(pub)發送消息，訂閱者(sub)接收消息。
Redis 客戶端可以訂閱任意數量的頻道。
當有新消息通過 PUBLISH 命令發送給頻道 channel 時， 這個消息就會被發送給訂閱它的所有客戶端。
**/

// Publish 將信息發送到指定的頻道，返回接收到信息的訂閱者數量
func (c *Cacher) Publish(channel, message string) error {
	cmd := c.Do("PUBLISH", channel, message)

	return cmd.Err
}

// Subscribe 訂閱給定的一個或多個頻道的信息。
// 支持redis服務停止或網絡異常等情況時，自動重新訂閱。
// 一般的程序都是啟動後開啟一些固定channel的訂閱，也不會動態的取消訂閱，這種場景下可以使用本方法。
// 覆雜場景的使用可以直接參考 https://godoc.org/github.com/gomodule/redigo/redis#hdr-Publish_and_Subscribe
func (c *Cacher) Subscribe(onMessage func(channel string, data []byte) error, channels ...string) error {
	// conn := c.pool.Get()
	// psc := redis.PubSubConn{Conn: conn}
	// err := psc.Subscribe(redis.Args{}.AddFlat(channels)...)
	// // 如果訂閱失敗，休息1秒後重新訂閱（比如當redis服務停止服務或網絡異常）
	// if err != nil {
	// 	fmt.Println(err)
	// 	time.Sleep(time.Second)
	// 	return c.Subscribe(onMessage, channels...)
	// }
	pubSub := c.pool.Subscribe(c.ctx.Context, channels...)
	// 處理消息
	ch := pubSub.Channel()
	go func() {
		for msg := range ch {
			//如果redis斷線 會自己重連
			go onMessage(msg.Channel, []byte(msg.Payload))

		}
	}()

	return nil
}

// getKey 將健名加上指定的前綴。
func (c *Cacher) getKey(key string) string {
	return c.prefix + key
}

// encode 序列化要保存的值
func (c *Cacher) encode(val interface{}) (interface{}, error) {
	var value interface{}
	switch v := val.(type) {
	case string, int, uint, int8, int16, int32, int64, float32, float64, bool, []byte:
		value = v
	default:
		b, err := json.Marshal(val)
		if err != nil {
			return nil, err
		}
		value = string(b)
	}

	return value, nil
}

// // 目前沒用到
// // decode 反序列化保存的struct對象
// func (c *Cacher) decode(reply interface{}, err error, val interface{}) error {
// 	str, err := String(reply, err)
// 	if err != nil {
// 		return err
// 	}

// 	err = json.Unmarshal([]byte(str), val)

// 	return err
// }

// init 註冊到cache
// func init() {
// 	cache.Register("redis", &Cacher{})
// }

// Script lua檔用struct
type Script struct {
	keyCount int
	src      string
	hash     string
}

// NewScript 產生新的Script
func NewScript(keyCount int, src string) *Script {
	h := sha1.New()
	io.WriteString(h, src)

	return &Script{keyCount, src, hex.EncodeToString(h.Sum(nil))}
}
func (s *Script) args(spec string, keysAndArgs []interface{}) []interface{} {
	var args []interface{}
	if s.keyCount < 0 {
		args = make([]interface{}, 1+len(keysAndArgs))
		args[0] = spec
		copy(args[1:], keysAndArgs)
	} else {
		args = make([]interface{}, 2+len(keysAndArgs))
		args[0] = spec
		args[1] = s.keyCount
		copy(args[2:], keysAndArgs)
	}

	return args
}

// DoScript 將lua檔執行
func (s *Script) DoScript(c *Cacher, keysAndArgs ...interface{}) (interface{}, error) {
	// TODO: e, ok := err.(Error), rror 無法正確轉出
	// v, err := c.Conn.Do("EVALSHA", s.args(s.hash, keysAndArgs)...)
	// if e, ok := err.(Error); ok && strings.HasPrefix(string(e), "NOSCRIPT ") {
	// conn := c.pool.Get()
	// defer conn.Close()

	if c.ctx.Context != nil {
		traceID := c.ctx.Context.Value(c.ctx.Field)
		log.Println("[Context] ", traceID, keysAndArgs)
	}
	v := c.Do("EVAL", s.args(s.src, keysAndArgs)...)
	// }

	return v.val, v.Err
}

// Scan 搜尋。
func (c *Cacher) Scan(cursor, count int, match string) *Cmd {
	return c.Do("SCAN", cursor, "MATCH", match, "COUNT", count)
}

// Mutex 包覆原本物件
type Mutex struct {
	mutexObject *redsync.Mutex
}

// MutexOption 包覆原本的option
type MutexOption struct {
	mutexOption redsync.Option
}

// NewMutex 產生新的Mutex
func (c *Cacher) NewMutex(mutexName string, options ...MutexOption) *Mutex {

	response := &Mutex{}
	response.mutexObject = c.syncRedis.NewMutex(mutexName)
	for _, o := range options {
		o.mutexOption.Apply(response.mutexObject)
	}

	return response
}

// Lock 上鎖
func (m *Mutex) Lock() error {
	err := m.mutexObject.Lock()
	return err
}

// UnLock 解鎖並回傳bool
func (m *Mutex) UnLock() (bool, error) {
	unlockBool, err := m.mutexObject.Unlock()
	return unlockBool, err
}

// OptionFunc is a function that configures a mutex.
type OptionFunc func(*Mutex)

// WithExpiry can be used to set the expiry of a mutex to the given value.
func WithExpiry(expiry time.Duration) MutexOption {
	option := redsync.WithExpiry(expiry)
	res := MutexOption{
		mutexOption: option,
	}
	return res
}

// WithTries can be used to set the number of times lock acquire is attempted.
func WithTries(tries int) MutexOption {
	option := redsync.WithTries(tries)
	res := MutexOption{
		mutexOption: option,
	}
	return res
}

// WithRetryDelay can be used to set the amount of time to wait between retries.
func WithRetryDelay(delay time.Duration) MutexOption {
	option := redsync.WithRetryDelay(delay)
	res := MutexOption{
		mutexOption: option,
	}
	return res
}

// WithRetryDelayFunc can be used to override default delay behavior.
func WithRetryDelayFunc(delayFunc redsync.DelayFunc) MutexOption {
	option := redsync.WithRetryDelayFunc(delayFunc)
	res := MutexOption{
		mutexOption: option,
	}
	return res
}

// WithDriftFactor can be used to set the clock drift factor.
func WithDriftFactor(factor float64) MutexOption {
	option := redsync.WithDriftFactor(factor)
	res := MutexOption{
		mutexOption: option,
	}
	return res
}

// WithGenValueFunc can be used to set the custom value generator.
func WithGenValueFunc(genValueFunc func() (string, error)) MutexOption {
	option := redsync.WithGenValueFunc(genValueFunc)
	res := MutexOption{
		mutexOption: option,
	}
	return res
}

// WithValue can be used to assign the random value without having to call lock. This allows the ownership of a lock to be "transfered" and allows the lock to be unlocked from elsewhere.
func WithValue(v string) MutexOption {
	option := redsync.WithValue(v)
	res := MutexOption{
		mutexOption: option,
	}
	return res
}

// ScriptLoad 返回集合內的所有的成員
func (c *Cacher) ScriptLoad(script string) (str string, err error) {
	contextDefault := context.Background()
	if c.ctx.Context != nil {
		contextDefault = c.ctx.Context
	}
	str, err = c.pool.ScriptLoad(contextDefault, script).Result()
	return str, err
}

func (c *Cacher) EvalSha(script string, keys []string, args ...interface{}) *Cmd {
	contextDefault := context.Background()
	if c.ctx.Context != nil {
		contextDefault = c.ctx.Context
	}
	cmd := &Cmd{}
	goRedisCmd := c.pool.EvalSha(contextDefault, script, keys, args...)
	cmd.cmd = goRedisCmd
	cmd.val = goRedisCmd.Val()
	cmd.Err = goRedisCmd.Err()
	if cmd.Err != nil && cmd.Err.Error() == "redis: nil" {
		// redigoNilErr := ErrNil
		cmd.Err = ErrNil
	}

	return cmd
}
