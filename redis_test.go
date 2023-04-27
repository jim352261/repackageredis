package redis

import (
	"context"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
)

var (
	redisCacher *Cacher
	redisLogger *log.Logger
)

func init() {
	redisLogger := log.New(os.Stdout, "", 0)
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	// defer s.Close()
	conf := Options{
		Addr:     s.Addr(),
		Db:       9,
		PoolSize: 10,
		MinIdle:  3,
		Wait:     true,
		Prefix:   "RedisTest:",
		Debug:    true,
		Log:      redisLogger,
	}

	redisCacher, _ = New(conf)
}

func TestNew(t *testing.T) {
	// conf := Options{
	// 	Addr:      "127.0.0.1:6379",
	// 	Db:        2,
	// 	MaxActive: 10,
	// 	MaxIdle:   100,
	// 	Prefix:    "Evelyn_",
	// }
	// conn, err := New(conf)
	// if err != nil {
	// 	t.Errorf("New error:%s ", err)
	// }

	reply, err := redisCacher.Do("PING").Value()
	if err != nil {
		t.Errorf("PING error:%s ", err)
	}

	if reply != `PONG` {
		t.Errorf("PONG error: %s", err)
	}
}

func TestScan(t *testing.T) {
	// conf := Options{
	// 	Addr:      "127.0.0.1:6379",
	// 	Db:        2,
	// 	MaxActive: 10,
	// 	MaxIdle:   100,
	// 	Prefix:    "Evelyn_",
	// }
	// conn, err := New(conf)
	// if err != nil {
	// 	t.Errorf("New error:%s ", err)
	// }

	type User struct {
		Name string
		Age  int
	}

	user := User{
		Name: "YM",
		Age:  52,
	}
	res := redisCacher.Set("TEST_SCAN", user, 0)
	if res.Err != nil {
		t.Errorf("set error:%s ", res.Err)
	}

	var u User
	if err := redisCacher.Get("TEST_SCAN").Scan(&u); err != nil {
		t.Errorf("Scan error:%s ", err)
	}
	if u.Age != user.Age || u.Name != user.Name {
		t.Errorf("got %v, want %v ", u, user)
	}
}

func TestDo(t *testing.T) {
	// loggerRedis := log.New(os.Stdout, "", 0)
	ctx := context.WithValue(context.Background(), "trace-id", "0123456789")

	// conf := Options{
	// 	Addr:      "127.0.0.1:6379",
	// 	Db:        2,
	// 	MaxActive: 10,
	// 	MaxIdle:   100,
	// 	Prefix:    "Evelyn_",
	// 	Debug:     true,
	// 	Log:       loggerRedis,
	// }

	// conn, err := New(conf)
	redisCacher = redisCacher.WithContext(ctx, "trace-id")

	var reply interface{}
	reply, err := redisCacher.Do("GET", "TEST-BB").String()
	if err != nil {
		if !strings.Contains(err.Error(), `redigo: nil returned`) {
			t.Errorf("GET %s, error : %s ", "TEST-BB", err)
		}
	}
	if reply != `` {
		t.Errorf("GET %s, reply : %s ", "TEST-BB", reply)
	}

	res := redisCacher.Set("TEST-BB", 1234567, 100)
	if res.Err != nil {
		t.Errorf("SET error:%s ", res.Err)
	}

	val, err := redisCacher.Get("TEST-BB").String()
	if res.Err != nil {
		t.Errorf("GET error:%s ", res.Err)
	}
	if val != `1234567` {
		t.Errorf("GET %s, reply : %s ", "TEST-BB", val)
	}

	res = redisCacher.Del("TEST-BB")
	if res.Err != nil {
		t.Errorf("DEL error:%s ", res.Err)
	}
}

func TestScript(t *testing.T) {
	// loggerRedis := log.New(os.Stdout, "", 0)
	luaScript := `return redis.call('INCRBY',KEYS[1],KEYS[2])`
	ctx := context.WithValue(context.Background(), "aa-id", "98769876")

	// conf := Options{
	// 	Addr:      "127.0.0.1:6379",
	// 	Db:        2,
	// 	MaxActive: 10,
	// 	MaxIdle:   100,
	// 	Prefix:    "EvelynScript_",
	// 	Debug:     true,
	// 	Log:       loggerRedis,
	// }

	// conn, err := New(conf)
	redisCacher = redisCacher.WithContext(ctx, "aa-id")

	var reply interface{}

	script1 := NewScript(2, luaScript)

	reply, err := Int(script1.DoScript(redisCacher, `BB`, 100, "GET"))

	if err != nil {
		t.Errorf("script error:%s", err)
	}

	if reply != 100 {
		t.Errorf("INCRBY %s %v, shoud be: %v,  reply : %v", "BB", 100, 100, reply)
	}

	reply, err = redisCacher.Do("DEL", "BB").Int64()
	if err != nil {
		t.Errorf("GET error:%s ", err)
	}

	// if reply != `100` {
	// 	t.Errorf("GET %s, reply : %s ", "BB", reply)
	// }

	// err = conn.Send(`EXPIRE`, `BB`, 100)
	// if err != nil {
	// 	t.Errorf("Send error:%s ", err)
	// }

	// err = conn.Flush()
	// if err != nil {
	// 	t.Errorf("Flush error:%s ", err)
	// }
}

func TestCacher_WithContext(t *testing.T) {
	type fields struct {
		pool   *redis.Client
		prefix string
		Log    *log.Logger
		ctx    ContextTraceInfo
	}
	type args struct {
		ctx   context.Context
		field string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Cacher
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cacher{
				pool:   tt.fields.pool,
				prefix: tt.fields.prefix,
				Log:    tt.fields.Log,
				ctx:    tt.fields.ctx,
			}
			if got := c.WithContext(tt.args.ctx, tt.args.field); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Cacher.WithContext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCacher_Do(t *testing.T) {
	type fields struct {
		pool   *redis.Client
		prefix string
		Log    *log.Logger
		ctx    ContextTraceInfo
	}
	type args struct {
		commandName string
		// args        []interface{}
		key string
		val string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   interface{}
	}{
		// TODO: Add test cases.
		{
			name: "testDo",
			fields: fields{
				pool:   redisCacher.pool,
				prefix: "Test-",
				Log:    redisLogger,
				ctx: ContextTraceInfo{
					Context: context.Background(),
					Field:   "test",
				},
			},
			args: args{
				commandName: "SET",
				key:         "DO-T1",
				val:         "T1",
			},
			want: nil,
		},
		{
			name: "testDo",
			fields: fields{
				pool:   redisCacher.pool,
				prefix: "Test-",
				Log:    redisLogger,
				ctx: ContextTraceInfo{
					Context: context.Background(),
					Field:   "test",
				},
			},
			args: args{
				commandName: "SETAA",
				key:         "DO-T1",
				val:         "T1",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cacher{
				pool:   tt.fields.pool,
				prefix: tt.fields.prefix,
				Log:    tt.fields.Log,
				ctx:    tt.fields.ctx,
			}
			got := c.Do(tt.args.commandName, tt.args.key, tt.args.val)
			if tt.args.commandName == "SET" && !reflect.DeepEqual(got.Err, tt.want) {
				t.Errorf("Cacher.Do() = %v, want %v", got.Err, tt.want)
			}
			if tt.args.commandName == "SETAA" && reflect.DeepEqual(got.Err, tt.want) {
				t.Errorf("Cacher.Do() = %v, want %v", got.Err, tt.want)
			}
		})
	}
}

func TestCacher_Set(t *testing.T) {
	type fields struct {
		pool   *redis.Client
		prefix string
		Log    *log.Logger
		ctx    ContextTraceInfo
	}
	type args struct {
		key    string
		val    interface{}
		expire int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   interface{}
	}{
		// TODO: Add test cases.
		{
			name: "testSet",
			fields: fields{
				pool:   redisCacher.pool,
				prefix: "Test-",
				Log:    redisLogger,
				ctx: ContextTraceInfo{
					Context: context.Background(),
					Field:   "test",
				},
			},
			args: args{
				key:    "TT-T1",
				val:    100,
				expire: 100,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cacher{
				pool:   tt.fields.pool,
				prefix: tt.fields.prefix,
				Log:    tt.fields.Log,
				ctx:    tt.fields.ctx,
			}
			if got := c.Set(tt.args.key, tt.args.val, tt.args.expire); !reflect.DeepEqual(got.Err, tt.want) {
				t.Errorf("Cacher.Set() = %v, want %v", got.Err, tt.want)
			}
		})
	}
}

func TestCacher_Get(t *testing.T) {
	type fields struct {
		pool   *redis.Client
		prefix string
		Log    *log.Logger
		ctx    ContextTraceInfo
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		// TODO: Add test cases.
		{
			name: "testGet",
			fields: fields{
				pool:   redisCacher.pool,
				prefix: "Test-",
				Log:    redisLogger,
				ctx: ContextTraceInfo{
					Context: context.Background(),
					Field:   "test",
				},
			},
			args: args{
				key: "TT-T1",
			},
			want: 100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cacher{
				pool:   tt.fields.pool,
				prefix: tt.fields.prefix,
				Log:    tt.fields.Log,
				ctx:    tt.fields.ctx,
			}
			got := c.Get(tt.args.key)
			val, _ := got.Int()
			type_int := fmt.Sprintf("%T", val)
			if !reflect.DeepEqual(val, tt.want) {
				t.Errorf("Cacher.Get() = %v, want %v", val, tt.want)
			}
			if type_int != "int" {
				t.Errorf("Type is %v", type_int)
			}
			val2, _ := got.String()
			type_string := fmt.Sprintf("%T", val2)
			if type_string != "string" {
				t.Errorf("Type is %v", type_string)
			}
			val3, _ := got.Float64()
			type_float64 := fmt.Sprintf("%T", val3)
			if type_float64 != "float64" {
				t.Errorf("Type is %v", type_float64)
			}
			val4, _ := got.Int64()
			type_int64 := fmt.Sprintf("%T", val4)
			if type_int64 != "int64" {
				t.Errorf("Type is %v", type_int64)
			}
		})
	}
}

func TestCacher_TTL(t *testing.T) {
	type fields struct {
		pool   *redis.Client
		prefix string
		Log    *log.Logger
		ctx    ContextTraceInfo
	}
	type args struct {
		key    string
		expire int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		// TODO: Add test cases.
		{
			name: "testTTL",
			fields: fields{
				pool:   redisCacher.pool,
				prefix: "Test-",
				Log:    redisLogger,
				ctx: ContextTraceInfo{
					Context: context.Background(),
					Field:   "test",
				},
			},
			args: args{
				key:    "TTL-test",
				expire: 10,
			},
			want: 10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cacher{
				pool:   tt.fields.pool,
				prefix: tt.fields.prefix,
				Log:    tt.fields.Log,
				ctx:    tt.fields.ctx,
			}
			c.Set(tt.args.key, "hello", int64(tt.args.expire))

			got, TTLErr := c.TTL(tt.args.key).Int()
			if TTLErr != nil {
				t.Errorf("Cacher.TTL() got error: %v", TTLErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Cacher.TTL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCacher_Expire(t *testing.T) {
	type fields struct {
		pool   *redis.Client
		prefix string
		Log    *log.Logger
		ctx    ContextTraceInfo
	}
	type args struct {
		key    string
		expire int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		// TODO: Add test cases.
		{
			name: "testExpire",
			fields: fields{
				pool:   redisCacher.pool,
				prefix: "Test-",
				Log:    redisLogger,
				ctx: ContextTraceInfo{
					Context: context.Background(),
					Field:   "test",
				},
			},
			args: args{
				key:    "EXPIRE-test",
				expire: 10,
			},
			want: 10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cacher{
				pool:   tt.fields.pool,
				prefix: tt.fields.prefix,
				Log:    tt.fields.Log,
				ctx:    tt.fields.ctx,
			}
			c.Set(tt.args.key, "hello", 0)
			expireErr := c.Expire(tt.args.key, int64(tt.args.expire)).Err
			if expireErr != nil {
				t.Errorf("Cacher.Expire() got error: %v", expireErr)
			}

			got, TTLErr := c.TTL(tt.args.key).Int()
			if TTLErr != nil {
				t.Errorf("Cacher.Expire() got error: %v", TTLErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Cacher.Expire() = %v, want %v", got, tt.want)
			}
		})
	}
}

// func TestCacher_Get_Race(t *testing.T) {
// 	type fields struct {
// 		pool   *redis.Client
// 		prefix string
// 		Log    *log.Logger
// 		ctx    ContextTraceInfo
// 	}
// 	type args struct {
// 		key string
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		args   args
// 		want   int
// 	}{
// 		// TODO: Add test cases.
// 		{
// 			name: "testGet",
// 			fields: fields{
// 				pool:   redisCacher.pool,
// 				prefix: "Test-",
// 				Log:    redisLogger,
// 				ctx: ContextTraceInfo{
// 					Context: context.Background(),
// 					Field:   "test",
// 				},
// 			},
// 			args: args{
// 				key: "TT-T1",
// 			},
// 			want: 100,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if setErr := redisCacher.Set(tt.args.key, 100, 100); setErr.Err != nil {
// 				t.Errorf("Set error:%s ", setErr.Err)
// 				return
// 			}
// 			var waitgroup sync.WaitGroup
// 			for c := 0; c < 1000; c++ {
// 				go func() {
// 					waitgroup.Add(1)

// 					got, zscoreErr := redisCacher.Get(tt.args.key).Int()
// 					if zscoreErr != nil {
// 						waitgroup.Done()
// 						t.Errorf("Get error:%s ", zscoreErr)
// 						return
// 					}
// 					if !reflect.DeepEqual(got, tt.want) {
// 						waitgroup.Done()
// 						t.Errorf("Cacher.Get() = %v, want %v", got, tt.want)
// 						return

// 					}
// 					waitgroup.Done()
// 				}()
// 			}
// 			waitgroup.Wait()
// 		})

// 	}
// }

func TestCacher_Helper(t *testing.T) {
	boolTest := true
	redisCacher.Set("Helper", boolTest, 0)

	v, _ := redisCacher.Get("Helper").Bool()
	if boolTest != v {
		t.Errorf("Cacher.Get() = %v, want %v", v, boolTest)
	}

	redisCacher.RPush("helper2", 100)
	redisCacher.RPush("helper2", 200)
	redisCacher.RPush("helper2", 300)

	v2, err := redisCacher.LRange("helper2", 0, -1).Int64s()
	if err != nil {
		t.Errorf("Cacher.LRange() got error: %v", err)
	}
	if !reflect.DeepEqual([]int64{100, 200, 300}, v2) {
		t.Errorf("Cacher.LRange() = %v, want %v", v2, []int64{100, 200, 300})
	}

	v3, err := redisCacher.LRange("helper2", 0, -1).Ints()
	if err != nil {
		t.Errorf("Cacher.LRange() got error: %v", err)
	}
	if !reflect.DeepEqual([]int{100, 200, 300}, v3) {
		t.Errorf("Cacher.LRange() = %v, want %v", v3, []int{100, 200, 300})
	}

	redisCacher.RPush("helper3", 1.23)
	redisCacher.RPush("helper3", 5.21)
	redisCacher.RPush("helper3", 5.33)
	v4, err := redisCacher.LRange("helper3", 0, -1).Float64s()
	if err != nil {
		t.Errorf("Cacher.LRange() got error: %v", err)
	}
	if !reflect.DeepEqual([]float64{1.23, 5.21, 5.33}, v4) {
		t.Errorf("Cacher.LRange() = %v, want %v", v4, []float64{1.23, 5.21, 5.33})
	}
}

func TestCacher_Del(t *testing.T) {
	type fields struct {
		pool   *redis.Client
		prefix string
		Log    *log.Logger
		ctx    ContextTraceInfo
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   interface{}
	}{
		// TODO: Add test cases.
		{
			name: "testDel",
			fields: fields{
				pool:   redisCacher.pool,
				prefix: "Test-",
				Log:    redisLogger,
				ctx: ContextTraceInfo{
					Context: context.Background(),
					Field:   "test",
				},
			},
			args: args{
				key: "TT-T1",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cacher{
				pool:   tt.fields.pool,
				prefix: tt.fields.prefix,
				Log:    tt.fields.Log,
			}
			if got := c.Del(tt.args.key); !reflect.DeepEqual(got.Err, tt.want) {
				t.Errorf("Cacher.Del() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCacher_IncrBy(t *testing.T) {
	type fields struct {
		pool   *redis.Client
		prefix string
		Log    *log.Logger
		ctx    ContextTraceInfo
	}
	type args struct {
		key    string
		amount int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		// TODO: Add test cases.
		{
			name: "testIncrBy",
			fields: fields{
				pool:   redisCacher.pool,
				prefix: "Test-",
				Log:    redisLogger,
				ctx: ContextTraceInfo{
					Context: context.Background(),
					Field:   "test",
				},
			},
			args: args{
				key:    "INC-T1",
				amount: 1,
			},
			want: 1,
		},
		{
			name: "testIncrBy",
			fields: fields{
				pool:   redisCacher.pool,
				prefix: "Test-",
				Log:    redisLogger,
			},
			args: args{
				key:    "INC-T1",
				amount: 1,
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cacher{
				pool:   tt.fields.pool,
				prefix: tt.fields.prefix,
				Log:    tt.fields.Log,
				ctx:    tt.fields.ctx,
			}
			if got, _ := c.IncrBy(tt.args.key, tt.args.amount).Int(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Cacher.IncrBy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCacher_DecrBy(t *testing.T) {
	type fields struct {
		pool   *redis.Client
		prefix string
		Log    *log.Logger
		ctx    ContextTraceInfo
	}
	type args struct {
		key    string
		amount int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		// TODO: Add test cases.
		{
			name: "testDecrBy",
			fields: fields{
				pool:   redisCacher.pool,
				prefix: "Test-",
				Log:    redisLogger,
				ctx: ContextTraceInfo{
					Context: context.Background(),
					Field:   "test",
				},
			},
			args: args{
				key:    "INC-T1",
				amount: 1,
			},
			want: 1,
		},
		{
			name: "testDecrBy",
			fields: fields{
				pool:   redisCacher.pool,
				prefix: "Test-",
				Log:    redisLogger,
			},
			args: args{
				key:    "DEC-T1",
				amount: 1,
			},
			want: -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cacher{
				pool:   tt.fields.pool,
				prefix: tt.fields.prefix,
				Log:    tt.fields.Log,
				ctx:    tt.fields.ctx,
			}
			if got, _ := c.DecrBy(tt.args.key, tt.args.amount).Int(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Cacher.DecrBy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCacher_HMSet(t *testing.T) {
	type fields struct {
		pool   *redis.Client
		prefix string
		Log    *log.Logger
		ctx    ContextTraceInfo
	}
	type args struct {
		key    string
		val    interface{}
		expire int
	}
	top1 := map[string]interface{}{"BookName": "Crazy golang", "Author": "Moon", "PageCount": "600", "Press": "GoodBook"}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   interface{}
	}{
		// TODO: Add test cases.
		{
			name: "testHMSet",
			fields: fields{
				pool:   redisCacher.pool,
				prefix: "Test-",
				Log:    redisLogger,
				ctx: ContextTraceInfo{
					Context: context.Background(),
					Field:   "test",
				},
			},
			args: args{
				key: "HMSet-T1",
				val: top1,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cacher{
				pool:   tt.fields.pool,
				prefix: tt.fields.prefix,
				Log:    tt.fields.Log,
				ctx:    tt.fields.ctx,
			}
			if got := c.HMSet(tt.args.key, tt.args.expire, tt.args.val); !reflect.DeepEqual(got.Err, tt.want) {
				t.Errorf("Cacher.HMSet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCacher_HSetNX(t *testing.T) {
	type fields struct {
		pool   *redis.Client
		prefix string
		Log    *log.Logger
		ctx    ContextTraceInfo
	}
	type args struct {
		key   string
		field string
		val   interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		// TODO: Add test cases.
		{
			name: "testHSetNX",
			fields: fields{
				pool:   redisCacher.pool,
				prefix: "TestHsetNX-",
				Log:    redisLogger,
				ctx: ContextTraceInfo{
					Context: context.Background(),
					Field:   "test",
				},
			},
			args: args{
				key:   "HSetNX-T1",
				field: "myhash",
				val:   1,
			},
			want: 1,
		},
		{
			name: "testHSetNX",
			fields: fields{
				pool:   redisCacher.pool,
				prefix: "TestHsetNX-",
				Log:    redisLogger,
			},
			args: args{
				key:   "HSetNX-T1",
				field: "myhash",
				val:   1,
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cacher{
				pool:   tt.fields.pool,
				prefix: tt.fields.prefix,
				Log:    tt.fields.Log,
				ctx:    tt.fields.ctx,
			}
			if got, _ := c.HSetNX(tt.args.key, tt.args.field, tt.args.val).Int(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Cacher.HSetNX() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCacher_HSet(t *testing.T) {
	type fields struct {
		pool   *redis.Client
		prefix string
		Log    *log.Logger
		ctx    ContextTraceInfo
	}
	type args struct {
		key   string
		field string
		val   interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		// TODO: Add test cases.
		{
			name: "testHSet",
			fields: fields{
				pool:   redisCacher.pool,
				prefix: "Test-",
				Log:    redisLogger,
				ctx: ContextTraceInfo{
					Context: context.Background(),
					Field:   "test",
				},
			},
			args: args{
				key:   "HSet-T1",
				field: "myhash",
				val:   1,
			},
			want: 1,
		},
		{
			name: "testHSet",
			fields: fields{
				pool:   redisCacher.pool,
				prefix: "Test-",
				Log:    redisLogger,
			},
			args: args{
				key:   "HSet-T1",
				field: "myhash",
				val:   2,
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cacher{
				pool:   tt.fields.pool,
				prefix: tt.fields.prefix,
				Log:    tt.fields.Log,
				ctx:    tt.fields.ctx,
			}
			if got, _ := c.HSet(tt.args.key, tt.args.field, tt.args.val).Int(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Cacher.HSet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCacher_HGet(t *testing.T) {
	type fields struct {
		pool   *redis.Client
		prefix string
		Log    *log.Logger
		ctx    ContextTraceInfo
	}
	type args struct {
		key   string
		field string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		// TODO: Add test cases.
		{
			name: "testHGet",
			fields: fields{
				pool:   redisCacher.pool,
				prefix: "Test-",
				Log:    redisLogger,
				ctx: ContextTraceInfo{
					Context: context.Background(),
					Field:   "test",
				},
			},
			args: args{
				key:   "HSet-T1",
				field: "myhash",
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cacher{
				pool:   tt.fields.pool,
				prefix: tt.fields.prefix,
				Log:    tt.fields.Log,
				ctx:    tt.fields.ctx,
			}
			if got, _ := c.HGet(tt.args.key, tt.args.field).Int(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Cacher.HGet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCacher_HGetAll(t *testing.T) {
	type fields struct {
		pool   *redis.Client
		prefix string
		Log    *log.Logger
		ctx    ContextTraceInfo
	}
	type args struct {
		key  string
		val  interface{}
		val2 interface{}
	}
	top1 := map[string]interface{}{"BookName": "Crazy golang", "Author": "Moon", "PageCount": "600", "Press": "GoodBook"}
	topx := map[string]interface{}{}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   interface{}
	}{
		// TODO: Add test cases.
		{
			name: "testHGetAll",
			fields: fields{
				pool:   redisCacher.pool,
				prefix: "Test-",
				Log:    redisLogger,
			},
			args: args{
				key:  "HGetAll-T1",
				val:  topx,
				val2: top1,
			},
			want: top1["BookName"],
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cacher{
				pool:   tt.fields.pool,
				prefix: tt.fields.prefix,
				Log:    tt.fields.Log,
				ctx:    tt.fields.ctx,
			}
			c.HMSet(tt.args.key, 100, tt.args.val2)
			got, _ := c.HGetAll(tt.args.key).StringMap()
			if !reflect.DeepEqual(got["BookName"], tt.want) {
				t.Errorf("Cacher.HGetAll() = %v, want %v", got["BookName"], tt.want)
			}
		})
	}
}

func TestCacher_HExists(t *testing.T) {
	type fields struct {
		pool   *redis.Client
		prefix string
		Log    *log.Logger
		ctx    ContextTraceInfo
	}
	type args struct {
		key   string
		field string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		// TODO: Add test cases.
		{
			name: "testHExists",
			fields: fields{
				pool:   redisCacher.pool,
				prefix: "Test-",
				Log:    redisLogger,
				ctx: ContextTraceInfo{
					Context: context.Background(),
					Field:   "test",
				},
			},
			args: args{
				key:   "HSet-T1",
				field: "myhash",
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cacher{
				pool:   tt.fields.pool,
				prefix: tt.fields.prefix,
				Log:    tt.fields.Log,
				ctx:    tt.fields.ctx,
			}
			c.HSet(tt.args.key, tt.args.field, "hello")
			if got, _ := c.HExists(tt.args.key, tt.args.field).Int(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Cacher.HExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCacher_HLen(t *testing.T) {
	type fields struct {
		pool   *redis.Client
		prefix string
		Log    *log.Logger
		ctx    ContextTraceInfo
	}
	type args struct {
		key   string
		field string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		// TODO: Add test cases.
		{
			name: "testHLen",
			fields: fields{
				pool:   redisCacher.pool,
				prefix: "Test-",
				Log:    redisLogger,
				ctx: ContextTraceInfo{
					Context: context.Background(),
					Field:   "test",
				},
			},
			args: args{
				key: "HLEN-T1",
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cacher{
				pool:   tt.fields.pool,
				prefix: tt.fields.prefix,
				Log:    tt.fields.Log,
				ctx:    tt.fields.ctx,
			}
			c.HSet(tt.args.key, tt.args.field, "hello")
			if got, _ := c.HLen(tt.args.key).Int(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Cacher.HLen() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCacher_HKEYS(t *testing.T) {
	type fields struct {
		pool   *redis.Client
		prefix string
		Log    *log.Logger
		ctx    ContextTraceInfo
	}
	type args struct {
		key    string
		field1 string
		field2 string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []string
	}{
		// TODO: Add test cases.
		{
			name: "testHLen",
			fields: fields{
				pool:   redisCacher.pool,
				prefix: "Test-",
				Log:    redisLogger,
				ctx: ContextTraceInfo{
					Context: context.Background(),
					Field:   "test",
				},
			},
			args: args{
				key:    "HKEYS-T1",
				field1: "field1",
				field2: "field2",
			},
			want: []string{"field1", "field2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cacher{
				pool:   tt.fields.pool,
				prefix: tt.fields.prefix,
				Log:    tt.fields.Log,
				ctx:    tt.fields.ctx,
			}
			c.HSet(tt.args.key, tt.args.field1, "hello")
			c.HSet(tt.args.key, tt.args.field2, "hello")
			if got, _ := c.HKeys(tt.args.key).Strings(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Cacher.HKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCacher_HIncrby(t *testing.T) {
	type fields struct {
		pool   *redis.Client
		prefix string
		Log    *log.Logger
		ctx    ContextTraceInfo
	}
	type args struct {
		key    string
		field  string
		number int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		// TODO: Add test cases.
		{
			name: "testHIncrby",
			fields: fields{
				pool:   redisCacher.pool,
				prefix: "Test-",
				Log:    redisLogger,
				ctx: ContextTraceInfo{
					Context: context.Background(),
					Field:   "test",
				},
			},
			args: args{
				key:    "HINCRBY-T1",
				field:  "field",
				number: 3,
			},
			want: 4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cacher{
				pool:   tt.fields.pool,
				prefix: tt.fields.prefix,
				Log:    tt.fields.Log,
				ctx:    tt.fields.ctx,
			}
			c.HSet(tt.args.key, tt.args.field, 1)
			c.HIncrby(tt.args.key, tt.args.field, tt.args.number)
			if got, _ := c.HGet(tt.args.key, tt.args.field).Int(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Cacher.HIncrby() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCacher_BLPop(t *testing.T) {
	type fields struct {
		pool   *redis.Client
		prefix string
		Log    *log.Logger
		ctx    ContextTraceInfo
	}
	type args struct {
		key     string
		timeout int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   interface{}
	}{
		// TODO: Add test cases.
		{
			name: "testBLPop",
			fields: fields{
				pool:   redisCacher.pool,
				prefix: "Test-",
				Log:    redisLogger,
			},
			args: args{
				key:     "BLPop-T1",
				timeout: 10,
			},
			want: "12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cacher{
				pool:   tt.fields.pool,
				prefix: tt.fields.prefix,
				Log:    tt.fields.Log,
				ctx:    tt.fields.ctx,
			}
			c.LPush(tt.args.key, tt.want)
			if got, _ := c.BLPop(tt.args.key, tt.args.timeout).String(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Cacher.BLPop() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCacher_BRPop(t *testing.T) {
	type fields struct {
		pool   *redis.Client
		prefix string
		Log    *log.Logger
		ctx    ContextTraceInfo
	}
	type args struct {
		key     string
		timeout int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   interface{}
	}{
		// TODO: Add test cases.
		{
			name: "testBRPop",
			fields: fields{
				pool:   redisCacher.pool,
				prefix: "Test-",
				Log:    redisLogger,
			},
			args: args{
				key:     "BRPop-T1",
				timeout: 10,
			},
			want: "12345",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cacher{
				pool:   tt.fields.pool,
				prefix: tt.fields.prefix,
				Log:    tt.fields.Log,
				ctx:    tt.fields.ctx,
			}
			c.RPush(tt.args.key, tt.want)
			if got, _ := c.BRPop(tt.args.key, tt.args.timeout).String(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Cacher.BRPop() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCacher_LPop(t *testing.T) {
	testKey := "TEST_LPop"
	redisCacher.RPush(testKey, 100)
	redisCacher.RPush(testKey, 200)

	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "LPop_1",
			args: args{
				key: testKey,
			},
			want: 100,
		},
		{
			name: "LPop_2",
			args: args{
				key: testKey,
			},
			want: 200,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := redisCacher
			cmd := c.LPop(tt.args.key)
			got, err := cmd.Int64()
			if err != nil {
				t.Errorf("Cacher.LPop() got error: %v", err)
			}

			if got != tt.want {
				t.Errorf("Cacher.LPop() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCacher_RPop(t *testing.T) {
	testKey := "TEST_RPop"
	redisCacher.RPush(testKey, "BB")
	redisCacher.RPush(testKey, "AA")

	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "RPop_1",
			args: args{
				key: testKey,
			},
			want: "AA",
		},
		{
			name: "RPop_2",
			args: args{
				key: testKey,
			},
			want: "BB",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := redisCacher
			cmd := c.RPop(tt.args.key)
			got, err := cmd.String()
			if err != nil {
				t.Errorf("Cacher.RPop() got error: %v", err)
			}

			if got != tt.want {
				t.Errorf("Cacher.RPop() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCacher_LPush(t *testing.T) {
	type args struct {
		key    string
		member interface{}
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "LPush",
			args: args{
				key:    "Test_LPush",
				member: "ABC",
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := redisCacher
			res := c.LPush(tt.args.key, tt.args.member)
			if res.Err != nil {
				t.Errorf("Cacher.LPush() Error %v", res.Err)
			}

			cmd := c.LLen(tt.args.key)
			Arrlen, err := cmd.Int()
			if err != nil {
				t.Errorf("Cacher.LLen() Error %v", err)
			}

			if Arrlen != tt.want {
				t.Errorf("LLen = %v, want %v", Arrlen, tt.want)
			}
		})
	}
}

func TestCacher_LPush2(t *testing.T) {
	type args struct {
		key    string
		member []interface{}
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "LPush2",
			args: args{
				key:    "Test_LPush2",
				member: []interface{}{1, 2, 3, 4, 5},
			},
			want: 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := redisCacher
			res := c.LPush(tt.args.key, tt.args.member...)
			if res.Err != nil {
				t.Errorf("Cacher.LPush() Error %v", res.Err)
			}

			cmd := c.LLen(tt.args.key)
			Arrlen, err := cmd.Int()
			if err != nil {
				t.Errorf("Cacher.LLen() Error %v", err)
			}

			if Arrlen != tt.want {
				t.Errorf("LLen = %v, want %v", Arrlen, tt.want)
			}
		})
	}
}

func TestCacher_RPush(t *testing.T) {
	testKey := "TEST_RPush"

	type args struct {
		key    string
		member interface{}
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "RPush",
			args: args{
				key:    testKey,
				member: "ABC",
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := redisCacher
			got := c.RPush(tt.args.key, tt.args.member)

			if got.Err != nil {
				t.Errorf("Cacher.RPush() Error :%v", got.Err)
			}

			cmd := c.LLen(tt.args.key)
			Arrlen, err := cmd.Int()
			if err != nil {
				t.Errorf("Cacher.LLen() Error %v", err)
			}

			if Arrlen != tt.want {
				t.Errorf("LLen = %v, want %v", Arrlen, tt.want)
			}
		})
	}
}

func TestCacher_RPush2(t *testing.T) {
	type args struct {
		key    string
		member []interface{}
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "RPush",
			args: args{
				key:    "Test_RPush2",
				member: []interface{}{1, 2, 3, 4, 5},
			},
			want: 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := redisCacher
			res := c.RPush(tt.args.key, tt.args.member...)
			if res.Err != nil {
				t.Errorf("Cacher.RPush() Error %v", res.Err)
			}

			cmd := c.LLen(tt.args.key)
			Arrlen, err := cmd.Int()
			if err != nil {
				t.Errorf("Cacher.LLen() Error %v", err)
			}

			if Arrlen != tt.want {
				t.Errorf("LLen = %v, want %v", Arrlen, tt.want)
			}
		})
	}
}

func TestCacher_LLen(t *testing.T) {
	testKey := "TEST_LLen"
	redisCacher.RPush(testKey, "BB")
	redisCacher.RPush(testKey, "AA")

	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "LLen",
			args: args{
				key: testKey,
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := redisCacher
			cmd := c.LLen(tt.args.key)
			got, err := cmd.Int()
			if err != nil {
				t.Errorf("Cacher.LLen() Error %v", err)
			}

			if got != tt.want {
				t.Errorf("Cacher.LLen() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCacher_LRange(t *testing.T) {
	testKey := "TEST_LRange"
	redisCacher.RPush(testKey, "BB")
	redisCacher.RPush(testKey, "AA")

	type args struct {
		key   string
		start int
		end   int
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "LRange",
			args: args{
				key:   testKey,
				start: 0,
				end:   1,
			},
			want: []string{
				"BB",
				"AA",
			},
		},
		{
			name: "LRange_2",
			args: args{
				key:   testKey,
				start: 0,
				end:   0,
			},
			want: []string{
				"BB",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := redisCacher
			cmd := c.LRange(tt.args.key, tt.args.start, tt.args.end)
			got, err := cmd.Strings()

			if err != nil {
				t.Errorf("Cacher.LRange() Error %v", err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Cacher.LRange() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCacher_LTrim(t *testing.T) {
	testKey := "TEST_LTrim"
	type args struct {
		key        string
		start      int32
		end        int32
		rangeStart int
		rangeEnd   int
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "LTRIM",
			args: args{
				key:        testKey,
				start:      1,
				end:        -1,
				rangeStart: 0,
				rangeEnd:   -1,
			},
			want: []string{
				"2",
				"3",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := redisCacher
			c.RPush(testKey, "1")
			c.RPush(testKey, "2")
			c.RPush(testKey, "3")
			trimCmd := c.LTrim(tt.args.key, int32(tt.args.start), int32(tt.args.end))
			trimRes, trimErr := trimCmd.String()
			if trimErr != nil {
				t.Errorf("Cacher.LTrim() Error %v", trimErr)
			}
			if !reflect.DeepEqual(trimRes, "OK") {
				t.Errorf("Cacher.TRim() = %v, want %v", trimRes, "OK")
			}
			cmd := c.LRange(tt.args.key, tt.args.rangeStart, tt.args.rangeEnd)
			got, err := cmd.Strings()

			if err != nil {
				t.Errorf("Cacher.LRange() Error %v", err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Cacher.LRange() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCacher_ZAdd(t *testing.T) {
	type args struct {
		key    string
		score  int64
		member string
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "testZadd",
			args: args{
				key:    "TEST-ZADD",
				score:  100,
				member: "TESTZADD",
			},
			want: 100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if zaddErr := redisCacher.ZAdd(tt.args.key, tt.args.score, tt.args.member); zaddErr.Err != nil {
				t.Errorf("Zadd error:%s ", zaddErr.Err)
			}
			got, zscoreErr := redisCacher.ZScore(tt.args.key, tt.args.member).Int64()
			if zscoreErr != nil {
				t.Errorf("ZScore error:%s ", zscoreErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Cacher.ZScore() = %v, want %v", got, tt.want)

			}
		})
	}
}

func TestCacher_ZRem(t *testing.T) {
	type args struct {
		key    string
		score  int64
		member string
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "testZadd",
			args: args{
				key:    "TEST-ZADD",
				score:  100,
				member: "TESTZADD",
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if zaddErr := redisCacher.ZAdd(tt.args.key, tt.args.score, tt.args.member); zaddErr.Err != nil {
				t.Errorf("Zadd error:%s ", zaddErr.Err)
			}
			if zremErr := redisCacher.ZRem(tt.args.key, tt.args.member); zremErr.Err != nil {
				t.Errorf("Zrem error:%s ", zremErr.Err)
			}
			got, zscoreErr := redisCacher.ZScore(tt.args.key, tt.args.member).Int64()
			if zscoreErr.Error() != "redigo: nil returned" {
				t.Errorf("ZScore error:%s ", zscoreErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Cacher.ZScore() = %v, want %v", got, tt.want)

			}
		})
	}
}

func TestCacher_ZScore(t *testing.T) {
	type args struct {
		key    string
		score  int64
		member string
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "testZadd",
			args: args{
				key:    "TEST-ZADD",
				score:  100,
				member: "TESTZADD",
			},
			want: 100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if zaddErr := redisCacher.ZAdd(tt.args.key, tt.args.score, tt.args.member); zaddErr.Err != nil {
				t.Errorf("Zadd error:%s ", zaddErr.Err)
			}
			got, zscoreErr := redisCacher.ZScore(tt.args.key, tt.args.member).Int64()
			if zscoreErr != nil {
				t.Errorf("ZScore error:%s ", zscoreErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Cacher.ZScore() = %v, want %v", got, tt.want)

			}
		})
	}
}

func TestCacher_ZRank(t *testing.T) {
	type args struct {
		key     string
		score   int64
		member  string
		score2  int64
		member2 string
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "testZadd",
			args: args{
				key:     "TEST-ZADD",
				score:   200,
				member:  "TESTZADD",
				score2:  100,
				member2: "TESTZADD2",
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if zaddErr := redisCacher.ZAdd(tt.args.key, tt.args.score, tt.args.member); zaddErr.Err != nil {
				t.Errorf("Zadd 1 error:%s ", zaddErr.Err)
			}

			if zaddErr2 := redisCacher.ZAdd(tt.args.key, tt.args.score2, tt.args.member2); zaddErr2.Err != nil {
				t.Errorf("Zadd 2 error:%s ", zaddErr2.Err)
			}
			got, zrankErr := redisCacher.ZRank(tt.args.key, tt.args.member).Int64()
			if zrankErr != nil {
				t.Errorf("ZRank error:%s ", zrankErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Cacher.ZRank() = %v, want %v", got, tt.want)

			}
		})
	}
}

func TestCacher_ZRevrank(t *testing.T) {
	type args struct {
		key     string
		score   int64
		member  string
		score2  int64
		member2 string
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "testZRevrank",
			args: args{
				key:     "TEST-ZADD",
				score:   200,
				member:  "TESTZADD",
				score2:  100,
				member2: "TESTZADD2",
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if zaddErr := redisCacher.ZAdd(tt.args.key, tt.args.score, tt.args.member); zaddErr.Err != nil {
				t.Errorf("Zadd 1 error:%s ", zaddErr.Err)
			}

			if zaddErr2 := redisCacher.ZAdd(tt.args.key, tt.args.score2, tt.args.member2); zaddErr2.Err != nil {
				t.Errorf("Zadd 2 error:%s ", zaddErr2.Err)
			}
			got, err := redisCacher.ZRevrank(tt.args.key, tt.args.member).Int64()
			if err != nil {
				t.Errorf("ZRevrank error:%s ", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Cacher.ZRevrank() = %v, want %v", got, tt.want)

			}
		})
	}
}

func TestCacher_ZRange(t *testing.T) {
	type args struct {
		key     string
		score   int64
		member  string
		score2  int64
		member2 string
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "testZadd",
			args: args{
				key:     "TEST-ZADD",
				score:   200,
				member:  "TESTZADD",
				score2:  100,
				member2: "TESTZADD2",
			},
			want: 100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if zaddErr := redisCacher.ZAdd(tt.args.key, tt.args.score, tt.args.member); zaddErr.Err != nil {
				t.Errorf("Zadd 1 error:%s ", zaddErr.Err)
			}

			if zaddErr2 := redisCacher.ZAdd(tt.args.key, tt.args.score2, tt.args.member2); zaddErr2.Err != nil {
				t.Errorf("Zadd 2 error:%s ", zaddErr2.Err)
			}

			reply, rangeErr := redisCacher.ZRangeWithScore(tt.args.key, 0, 30).Values()
			if rangeErr != nil {
				t.Errorf("ZRangeWithScore error:%s ", rangeErr)
			}
			value1, err := Int64(reply[1], rangeErr)
			if err != nil {
				t.Errorf("Int64 error:%s ", err)
			}
			if !reflect.DeepEqual(value1, tt.want) {
				t.Errorf("Cacher.ZRangeWithScore() = %v, want %v", value1, tt.want)

			}

			replymap, rangeMapErr := redisCacher.ZRangeWithScore(tt.args.key, 0, 30).Int64Map()
			if rangeMapErr != nil {
				t.Errorf("ZRangeWithScore error:%s ", rangeMapErr)
			}
			if !reflect.DeepEqual(replymap["TESTZADD2"], tt.want) {
				t.Errorf("Cacher.ZRangeWithScore() = %v, want %v", replymap["TESTZADD2"], 100)

			}
		})
	}
}

func TestCacher_ZRevrange(t *testing.T) {
	type args struct {
		key     string
		score   int64
		member  string
		score2  int64
		member2 string
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "testZadd",
			args: args{
				key:     "TEST-ZADD",
				score:   200,
				member:  "TESTZADD",
				score2:  100,
				member2: "TESTZADD2",
			},
			want: 200,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if zaddErr := redisCacher.ZAdd(tt.args.key, tt.args.score, tt.args.member); zaddErr.Err != nil {
				t.Errorf("Zadd 1 error:%s ", zaddErr.Err)
			}

			if zaddErr2 := redisCacher.ZAdd(tt.args.key, tt.args.score2, tt.args.member2); zaddErr2.Err != nil {
				t.Errorf("Zadd 2 error:%s ", zaddErr2.Err)
			}

			reply, rangeErr := redisCacher.ZRevrange(tt.args.key, 0, 30).Values()
			if rangeErr != nil {
				t.Errorf("ZRevrange error:%s ", rangeErr)
			}
			value1, err := Int64(reply[1], rangeErr)
			if err != nil {
				t.Errorf("Int64 error:%s ", err)
			}
			if !reflect.DeepEqual(value1, tt.want) {
				t.Errorf("Cacher.ZRevrange() = %v, want %v", value1, tt.want)

			}

			replymap, rangeMapErr := redisCacher.ZRangeWithScore(tt.args.key, 0, 30).IntMap()
			if rangeMapErr != nil {
				t.Errorf("ZRangeWithScore error:%s ", rangeMapErr)
			}
			if !reflect.DeepEqual(replymap["TESTZADD2"], 100) {
				t.Errorf("Cacher.ZRangeWithScore() = %v, want %v", replymap["TESTZADD2"], 100)

			}
		})
	}
}

func TestCacher_ZRangeByScore(t *testing.T) {
	type args struct {
		key     string
		score   int64
		member  string
		score2  int64
		member2 string
		score3  int64
		member3 string
		from    int64
		to      int64
		offset  int64
		count   int
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "testZRangeByScore",
			args: args{
				key:     "TEST-ZADD",
				score:   200,
				member:  "TESTZADD",
				score2:  100,
				member2: "TESTZADD2",
				score3:  300,
				member3: "TESTZADD3",
				from:    100,
				to:      200,
				offset:  0,
				count:   1,
			},
			want: 100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if zaddErr := redisCacher.ZAdd(tt.args.key, tt.args.score, tt.args.member); zaddErr.Err != nil {
				t.Errorf("Zadd 1 error:%s ", zaddErr.Err)
			}

			if zaddErr2 := redisCacher.ZAdd(tt.args.key, tt.args.score2, tt.args.member2); zaddErr2.Err != nil {
				t.Errorf("Zadd 2 error:%s ", zaddErr2.Err)
			}

			if zaddErr3 := redisCacher.ZAdd(tt.args.key, tt.args.score3, tt.args.member3); zaddErr3.Err != nil {
				t.Errorf("Zadd 3 error:%s ", zaddErr3.Err)
			}

			reply, rangeErr := redisCacher.ZRangeByScore(tt.args.key, tt.args.from, tt.args.to, tt.args.offset, tt.args.count).Values()
			if rangeErr != nil {
				t.Errorf("ZRangeByScore error:%s ", rangeErr)
			}
			value1, err := Int64(reply[1], rangeErr)
			if err != nil {
				t.Errorf("Int64 error:%s ", err)
			}
			if !reflect.DeepEqual(value1, tt.want) {
				t.Errorf("Cacher.ZRevrange() = %v, want %v", value1, tt.want)

			}
		})
	}
}

func TestCacher_ZRevrangeByScore(t *testing.T) {
	type args struct {
		key     string
		score   int64
		member  string
		score2  int64
		member2 string
		score3  int64
		member3 string
		from    int64
		to      int64
		offset  int64
		count   int
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "testZRangeByScore",
			args: args{
				key:     "TEST-ZADD",
				score:   1,
				member:  "TESTZADD",
				score2:  2,
				member2: "TESTZADD2",
				score3:  3,
				member3: "TESTZADD3",
				from:    10,
				to:      1,
				offset:  0,
				count:   1,
			},
			want: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if zaddErr := redisCacher.ZAdd(tt.args.key, tt.args.score, tt.args.member); zaddErr.Err != nil {
				t.Errorf("Zadd 1 error:%s ", zaddErr.Err)
			}

			if zaddErr2 := redisCacher.ZAdd(tt.args.key, tt.args.score2, tt.args.member2); zaddErr2.Err != nil {
				t.Errorf("Zadd 2 error:%s ", zaddErr2.Err)
			}

			if zaddErr3 := redisCacher.ZAdd(tt.args.key, tt.args.score3, tt.args.member3); zaddErr3.Err != nil {
				t.Errorf("Zadd 3 error:%s ", zaddErr3.Err)
			}

			reply, rangeErr := redisCacher.ZRevrangeByScore(tt.args.key, tt.args.from, tt.args.to, tt.args.offset, tt.args.count).Values()
			if rangeErr != nil {
				t.Errorf("ZRevrangeByScore error:%s ", rangeErr)
			}

			value1, err := Int64(reply[1], rangeErr)
			if err != nil {
				t.Errorf("Int64 error:%s ", err)
			}
			if !reflect.DeepEqual(value1, tt.want) {
				t.Errorf("Cacher.ZRevrangeByScore() = %v, want %v", value1, tt.want)

			}
		})
	}
}

func TestCacher_ZCard(t *testing.T) {
	type args struct {
		key          string
		member       string
		memberScore  int64
		member2      string
		member2Score int64
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "testScard",
			args: args{
				key:          "testnyset",
				member:       "Jim",
				memberScore:  1,
				member2:      "ken",
				member2Score: 2,
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if zaddErr := redisCacher.ZAdd(tt.args.key, tt.args.memberScore, tt.args.member); zaddErr.Err != nil {
				t.Errorf("Zadd 1 error:%s ", zaddErr.Err)
			}
			if zaddErr2 := redisCacher.ZAdd(tt.args.key, tt.args.member2Score, tt.args.member2); zaddErr2.Err != nil {
				t.Errorf("Zadd 2 error:%s ", zaddErr2.Err)
			}

			reply, zcardErr := redisCacher.ZCard(tt.args.key).Int64()
			if zcardErr != nil {
				t.Errorf("ZCard error:%s ", zcardErr)
			}

			if !reflect.DeepEqual(reply, tt.want) {
				t.Errorf("Cacher.Zcard() = %v, want %v", reply, tt.want)

			}
			//排掉
			_, remErr := redisCacher.ZRem(tt.args.key, tt.args.member).Bool()
			if remErr != nil {
				t.Errorf("ZRem error:%s ", remErr)
			}
			_, rem2Err := redisCacher.ZRem(tt.args.key, tt.args.member2).Bool()
			if rem2Err != nil {
				t.Errorf("ZRem error:%s ", rem2Err)
			}

		})
	}
}

func TestCacher_SAdd(t *testing.T) {
	type args struct {
		key    string
		member string
		count  int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "testSAdd",
			args: args{
				key:    "testnyset",
				member: "Jim",
				count:  1,
			},
			want: "Jim",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if saddErr := redisCacher.SAdd(tt.args.key, tt.args.member); saddErr.Err != nil {
				t.Errorf("Sadd 1 error:%s ", saddErr.Err)
			}

			reply, popErr := redisCacher.SPop(tt.args.key, tt.args.count).Values()
			if popErr != nil {
				t.Errorf("SPop error:%s ", popErr)
			}

			value1, err := String(reply[0], popErr)
			if err != nil {
				t.Errorf("Int64 error:%s ", err)
			}
			if !reflect.DeepEqual(value1, tt.want) {
				t.Errorf("Cacher.SAdd() = %v, want %v", value1, tt.want)

			}
		})
	}
}

func TestCacher_SRem(t *testing.T) {
	type args struct {
		key     string
		member  string
		member2 string
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "testRem",
			args: args{
				key:     "testnyset",
				member:  "Jim",
				member2: "ken",
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if saddErr := redisCacher.SAdd(tt.args.key, tt.args.member); saddErr.Err != nil {
				t.Errorf("Sadd 1 error:%s ", saddErr.Err)
			}
			if saddErr2 := redisCacher.SAdd(tt.args.key, tt.args.member2); saddErr2.Err != nil {
				t.Errorf("Sadd 2 error:%s ", saddErr2.Err)
			}
			if sremErr := redisCacher.SRem(tt.args.key, tt.args.member); sremErr.Err != nil {
				t.Errorf("SRem error:%s ", sremErr.Err)
			}

			reply, scardErr := redisCacher.SCard(tt.args.key).Int64()
			if scardErr != nil {
				t.Errorf("SCard error:%s ", scardErr)
			}

			if !reflect.DeepEqual(reply, tt.want) {
				t.Errorf("Cacher.Scard() = %v, want %v", reply, tt.want)

			}
			//排掉
			_, popErr := redisCacher.SPop(tt.args.key, 1).Values()
			if popErr != nil {
				t.Errorf("SPop error:%s ", popErr)
			}

		})
	}
}

func TestCacher_SCard(t *testing.T) {
	type args struct {
		key     string
		member  string
		member2 string
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "testScard",
			args: args{
				key:     "testnyset",
				member:  "Jim",
				member2: "ken",
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if saddErr := redisCacher.SAdd(tt.args.key, tt.args.member); saddErr.Err != nil {
				t.Errorf("Sadd 1 error:%s ", saddErr.Err)
			}
			if saddErr2 := redisCacher.SAdd(tt.args.key, tt.args.member2); saddErr2.Err != nil {
				t.Errorf("Sadd 2 error:%s ", saddErr2.Err)
			}

			reply, scardErr := redisCacher.SCard(tt.args.key).Int64()
			if scardErr != nil {
				t.Errorf("SCard error:%s ", scardErr)
			}

			if !reflect.DeepEqual(reply, tt.want) {
				t.Errorf("Cacher.Scard() = %v, want %v", reply, tt.want)

			}
			//排掉
			_, popErr := redisCacher.SPop(tt.args.key, 2).Values()
			if popErr != nil {
				t.Errorf("SPop error:%s ", popErr)
			}

		})
	}
}

func TestCacher_SPop(t *testing.T) {
	type args struct {
		key    string
		member string
		count  int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "testSAdd",
			args: args{
				key:    "testnyset",
				member: "Jim",
				count:  1,
			},
			want: "Jim",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if saddErr := redisCacher.SAdd(tt.args.key, tt.args.member); saddErr.Err != nil {
				t.Errorf("Sadd 1 error:%s ", saddErr.Err)
			}

			reply, popErr := redisCacher.SPop(tt.args.key, tt.args.count).Values()
			if popErr != nil {
				t.Errorf("SPop error:%s ", popErr)
			}

			value1, err := String(reply[0], popErr)
			if err != nil {
				t.Errorf("Int64 error:%s ", err)
			}
			if !reflect.DeepEqual(value1, tt.want) {
				t.Errorf("Cacher.SAdd() = %v, want %v", value1, tt.want)

			}
		})
	}
}

func TestCacher_SisMembers(t *testing.T) {
	type args struct {
		key         string
		member1     string
		member2     string
		checkMember string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "testSisMembers",
			args: args{
				key:         "testSismembers",
				member1:     "Jim",
				member2:     "ken",
				checkMember: "Jim",
			},
			want: true,
		}, {
			name: "testSisMembers2",
			args: args{
				key:         "testSismembers",
				member1:     "Jim",
				member2:     "ken",
				checkMember: "Wesley",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if saddErr := redisCacher.SAdd(tt.args.key, tt.args.member1); saddErr.Err != nil {
				t.Errorf("SisMembers Sadd 1 error:%s ", saddErr.Err)
			}

			if saddErr := redisCacher.SAdd(tt.args.key, tt.args.member2); saddErr.Err != nil {
				t.Errorf("SisMembers SAdd 2 error:%s ", saddErr.Err)
			}

			reply, popErr := redisCacher.SisMembers(tt.args.key, tt.args.checkMember).Bool()
			if popErr != nil {
				t.Errorf("SisMembers error:%s ", popErr)
			}

			if !reflect.DeepEqual(reply, tt.want) {
				t.Errorf("Cacher.SisMembers() = %v, want %v", reply, tt.want)

			}
		})
	}
}

func TestCacher_SMembers(t *testing.T) {
	type args struct {
		key     string
		member1 string
		member2 string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "testSMembers",
			args: args{
				key:     "testSmembers",
				member1: "Jim",
				member2: "ken",
			},
			want: []string{"Jim", "ken"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if saddErr := redisCacher.SAdd(tt.args.key, tt.args.member1); saddErr.Err != nil {
				t.Errorf("SMembers Sadd 1 error:%s ", saddErr.Err)
			}

			if saddErr := redisCacher.SAdd(tt.args.key, tt.args.member2); saddErr.Err != nil {
				t.Errorf("SMembers SAdd 2 error:%s ", saddErr.Err)
			}

			reply, popErr := redisCacher.SMembers(tt.args.key).Strings()
			if popErr != nil {
				t.Errorf("SMembers error:%s ", popErr)
			}

			if !reflect.DeepEqual(reply, tt.want) {
				t.Errorf("Cacher.SMembers() = %v, want %v", reply, tt.want)

			}
		})
	}
}

// 用miniredis跑Publish會有問題
func TestCacher_Publish(t *testing.T) {
}

func TestCacher_Subscribe(t *testing.T) {
	callBack := func(channel string, data []byte) error {
		fmt.Println(channel)
		fmt.Println(data)

		return nil
	}

	type args struct {
		onMessage func(channel string, data []byte) error
		channels  []string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Test_Subscribe_1",
			args: args{
				onMessage: callBack,
				channels:  []string{"orders"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := redisCacher
			if err := c.Subscribe(tt.args.onMessage, tt.args.channels...); err != nil {
				t.Errorf("Cacher.Subscribe() error = %v", err)
			}
		})
	}
}

func TestCacher_KEYS(t *testing.T) {
	// 先set key
	redisCacher.Set("aa:a", "test", 0)
	redisCacher.Set("aa:b", "test", 0)
	redisCacher.Set("aa:c", "test", 0)

	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "testKEYS",
			args: args{
				key: "aa:*",
			},
			want: []string{"a", "b", "c"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := redisCacher.Keys(tt.args.key)
			if cmd.Err != nil {
				t.Errorf("got error = %v", cmd.Err)
			}

			val, err := cmd.Strings()
			if err != nil {
				t.Errorf("got error = %v", cmd.Err)
			}

			if len(tt.want) != len(val) {
				t.Errorf("len error got = %v, want = %v", len(val), len(tt.want))
			}
		})
	}
}

func TestCacher_SetNX(t *testing.T) {
	type fields struct {
		pool   *redis.Client
		prefix string
		Log    *log.Logger
		ctx    ContextTraceInfo
	}
	type args struct {
		key    string
		val    int64
		expire int64
	}
	tfields := fields{
		pool:   redisCacher.pool,
		prefix: "Test-",
		Log:    redisLogger,
		ctx: ContextTraceInfo{
			Context: context.Background(),
			Field:   "test",
		},
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
		{
			name:   "testSetNX1",
			fields: tfields,
			args: args{
				key:    "SetNX-T1",
				val:    1,
				expire: 1,
			},
			want: true,
		},
		{
			name:   "testSetNX2",
			fields: tfields,
			args: args{
				key:    "SetNX-T1",
				val:    1,
				expire: 1,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cacher{
				pool:   tt.fields.pool,
				prefix: tt.fields.prefix,
				Log:    tt.fields.Log,
				ctx:    tt.fields.ctx,
			}
			exits, _ := c.SetNX(tt.args.key, tt.args.val, tt.args.expire).Bool()
			if !reflect.DeepEqual(exits, tt.want) {
				t.Errorf("Cacher.Set() = %v, want %v", exits, tt.want)
			}
		})
	}
}
