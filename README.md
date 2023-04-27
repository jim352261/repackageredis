# Redis
```
對於常用的func進行一對一封裝
沒有封裝到的，可以使用func Do直接下指令
目前回傳值統一為interface

如果有其他的重要功能沒包到
請再提出 & 建issue

使用的Redis套件:
github.com/go-redis/redis/v8
```


## 已封裝的func
- Do
- Get
- Set
- Del
- IncrBy
- DecrBy
- HMSet
- HSet
- HGet
- HGetAll
- BLPop
- BRPop
- LPop
- RPop
- LPush
- RPush
- LLen
- LRange
- ZAdd
- ZRem
- ZScore
- ZRank
- ZRevrank
- ZRange
- ZRevrange
- ZRangeByScore
- ZRevrangeByScore
- Publish
- Subscribe
- SetNX

## Example
```
loggerRedis := log.New(os.Stdout, "", 0)
redisClient, err := redis.New(redis.Options{
    Addr:      "0.0.0.0:6379",
    Password:  "",
    Db:        0,
    MaxActive: 100,
    MaxIdle:   10,
    Debug:     true,
    Log:       loggerRedis,
})

if err != nil {
    fmt.Println(err)
}

redisClient.Set("Hello", "world", 0)
val, err := redisClient.Get("Hello").String()
if err != nil {
    fmt.Println(err)
} else {
    fmt.Printf("Hello %s \n", val)
}

res := redisClient.Set("float64", 3.4123, 0)
if res.Err != nil {
    fmt.Println(res.Err)
}
float64Val, err := redisClient.Get("float64").Float64()
if err != nil {
    fmt.Println(err)
} else {
    fmt.Printf("float64 %f", float64Val)
}
```


## redsync 已封入的方法
- NewMutex          （產生mutex實體）
- Lock               (上鎖)
- UnLock             (解鎖)
- WithExpiry         (超時時間 預設8秒)
- WithTries          (上鎖重試次數 預設32次)
- WithRetryDelay     (重試間隔)
- WithRetryDelayFunc (重試行為)
- WithDriftFactor
- WithGenValueFunc   (自定義setNX的值)
- WithValue          (自定義可傳遞的setNX的值)


## Example

```
loggerRedis := log.New(os.Stdout, "", 0)
redisClient, err := redis.New(redis.Options{
    Addr:      "0.0.0.0:6379",
    Password:  "",
    Db:        0,
    MaxActive: 100,
    MaxIdle:   10,
    Debug:     true,
    Log:       loggerRedis,
})

if err != nil {
    fmt.Println(err)
}

mutexObject , err := redisClient.NewMutex("test", 
		                redis.WithExpiry(time.Duration(30)*time.Second),
		                redis.WithRetryDelay(time.Duration(10)*time.Millisecond)),
    )
mutexObject.Lock()
mutexObject.Unlock()
```