package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/go-redis/redis/v8"
)

type Cmd struct {
	cmd *redis.Cmd
	val interface{}
	Err error
}

// NewCmd NewCmd
func NewCmd(val interface{}) *Cmd {
	return &Cmd{
		cmd: redis.NewCmd(context.Background()),
		val: val,
	}
}

// Value 取得回傳值
func (c *Cmd) Value() (interface{}, error) {
	if c.Err != nil {
		return nil, c.Err
	}

	return c.val, nil
}

// Values 取得回傳值
// go-redis 使用 result
func (c *Cmd) Values() ([]interface{}, error) {
	return Values(c.val, c.Err)
}

// String String
// go-redis 使用 string
func (c *Cmd) String() (string, error) {
	return String(c.val, c.Err)
}

// Strings Strings
// go-redis 使用
func (c *Cmd) Strings() ([]string, error) {
	return Strings(c.val, c.Err)
}

// Bytes Bytes
func (c *Cmd) Bytes() ([]byte, error) {
	return Bytes(c.val, c.Err)
}

// Int Int
func (c *Cmd) Int() (int, error) {
	return Int(c.val, c.Err)
}

// Ints Ints
func (c *Cmd) Ints() ([]int, error) {
	return Ints(c.val, c.Err)
}

// Int64 Int64
func (c *Cmd) Int64() (int64, error) {
	return Int64(c.val, c.Err)
}

// Int64 Int64
func (c *Cmd) Int64s() ([]int64, error) {
	return Int64s(c.val, c.Err)
}

// Float64 Float64
func (c *Cmd) Float64() (float64, error) {
	return Float64(c.val, c.Err)
}

// Float64s Float64s
func (c *Cmd) Float64s() ([]float64, error) {
	return Float64s(c.val, c.Err)
}

// Bool Bool
func (c *Cmd) Bool() (bool, error) {
	return Bool(c.val, c.Err)
}

// IntMap IntMap
func (c *Cmd) IntMap() (map[string]int, error) {
	return IntMap(c.val, c.Err)
}

// Int64Map Int64Map
func (c *Cmd) Int64Map() (map[string]int64, error) {
	return Int64Map(c.val, c.Err)
}

// StringMap StringMap
func (c *Cmd) StringMap() (map[string]string, error) {
	return StringMap(c.val, c.Err)
}

// Scan Scan
func (c *Cmd) Scan(obj interface{}) error {
	if c.Err != nil {
		return c.Err
	}

	// interface 轉 bytes
	val, err := c.Bytes()
	if err != nil {
		return err
	}

	// decode
	if err := json.Unmarshal(val, obj); err != nil {
		return err
	}

	return nil
}

// Int is a helper that converts a command reply to an integer
func Int(reply interface{}, err error) (int, error) {
	if err != nil {
		return 0, err
	}
	switch reply := reply.(type) {
	case int:
		return reply, nil
	case int32:
		x := int(reply)
		if int32(x) != reply {
			return 0, strconv.ErrRange
		}
		return x, nil
	case int64:
		x := int(reply)
		if int64(x) != reply {
			return 0, strconv.ErrRange
		}
		return x, nil
	case []byte:
		n, err := strconv.ParseInt(string(reply), 10, 0)
		return int(n), err
	case string:
		n, err := strconv.ParseInt(reply, 10, 0)
		return int(n), err
	case nil:
		return 0, ErrNil
	case Error:
		return 0, reply
	}
	return 0, fmt.Errorf("redigo: unexpected type for Int, got type %T", reply)
}

// Ints Ints
func Ints(reply interface{}, err error) ([]int, error) {
	var result []int
	err = sliceHelper(reply, err, "Ints", func(n int) { result = make([]int, n) }, func(i int, v interface{}) error {
		switch v := v.(type) {
		case int:
			result[i] = v
			return nil
		case int32:
			n := int(v)
			if int32(n) != v {
				return strconv.ErrRange
			}
			result[i] = n
			return nil
		case int64:
			n := int(v)
			if int64(n) != v {
				return strconv.ErrRange
			}
			result[i] = n
			return nil
		case []byte:
			n, err := strconv.Atoi(string(v))
			result[i] = n
			return err
		case string:
			n, err := strconv.Atoi(v)
			result[i] = n
			return err
		default:
			return fmt.Errorf("redigo: unexpected element type for Ints, got type %T", v)
		}
	})
	return result, err
}

// Int64 is a helper that converts a command reply to 64 bit integer
func Int64(reply interface{}, err error) (int64, error) {
	if err != nil {
		return 0, err
	}
	switch reply := reply.(type) {
	case int:
		return int64(reply), nil
	case int32:
		return int64(reply), nil
	case int64:
		return reply, nil
	case []byte:
		n, err := strconv.ParseInt(string(reply), 10, 64)
		return n, err
	case string:
		n, err := strconv.ParseInt(reply, 10, 64)
		return n, err
	case nil:
		return 0, ErrNil
	case Error:
		return 0, reply
	}
	return 0, fmt.Errorf("redigo: unexpected type for Int64, got type %T", reply)
}

// Int64s Int64s
func Int64s(reply interface{}, err error) ([]int64, error) {
	var result []int64
	err = sliceHelper(reply, err, "Int64s", func(n int) { result = make([]int64, n) }, func(i int, v interface{}) error {
		switch v := v.(type) {
		case int:
			result[i] = int64(v)
			return nil
		case int32:
			result[i] = int64(v)
			return nil
		case int64:
			result[i] = v
			return nil
		case []byte:
			n, err := strconv.ParseInt(string(v), 10, 64)
			result[i] = n
			return err
		case string:
			n, err := strconv.ParseInt(v, 10, 64)
			result[i] = n
			return err
		default:
			return fmt.Errorf("redigo: unexpected element type for Int64s, got type %T", v)
		}
	})
	return result, err
}

// String is a helper that converts a command reply to a string
func String(reply interface{}, err error) (string, error) {
	if err != nil {
		return "", err
	}
	switch reply := reply.(type) {
	case []byte:
		return string(reply), nil
	case string:
		return reply, nil
	case nil:
		return "", ErrNil
	case Error:
		return "", reply
	}
	return "", fmt.Errorf("redigo: unexpected type for String, got type %T", reply)
}

// Strings Strings
func Strings(reply interface{}, err error) ([]string, error) {
	var result []string
	err = sliceHelper(reply, err, "Strings", func(n int) { result = make([]string, n) }, func(i int, v interface{}) error {
		switch v := v.(type) {
		case string:
			result[i] = v
			return nil
		case []byte:
			result[i] = string(v)
			return nil
		default:
			return fmt.Errorf("redigo: unexpected element type for Strings, got type %T", v)
		}
	})
	return result, err
}

// Bool is a helper that converts a command reply to a boolean
func Bool(reply interface{}, err error) (bool, error) {
	if err != nil {
		return false, err
	}
	switch reply := reply.(type) {
	case int64:
		return reply != 0, nil
	case []byte:
		return strconv.ParseBool(string(reply))
	case string:
		return strconv.ParseBool(reply)
	case nil:
		return false, ErrNil
	case Error:
		return false, reply
	}
	return false, fmt.Errorf("redigo: unexpected type for Bool, got type %T", reply)
}

// Float64 is a helper that converts a command reply to a float64
func Float64(reply interface{}, err error) (float64, error) {
	if err != nil {
		return 0, err
	}
	switch reply := reply.(type) {
	case []byte:
		n, err := strconv.ParseFloat(string(reply), 64)
		return n, err
	case string:
		n, err := strconv.ParseFloat(reply, 64)
		return n, err
	case nil:
		return 0, ErrNil
	case Error:
		return 0, reply
	}
	return 0, fmt.Errorf("redigo: unexpected type for Float64, got type %T", reply)
}

// Float64s Float64s
func Float64s(reply interface{}, err error) ([]float64, error) {
	var result []float64
	err = sliceHelper(reply, err, "Float64s", func(n int) { result = make([]float64, n) }, func(i int, v interface{}) error {
		p, ok := v.([]byte)
		pString, okString := v.(string)
		if !ok && !okString {
			return fmt.Errorf("redigo: unexpected element type for Floats64, got type %T", v)
		}
		if okString {
			p = []byte(pString)
		}
		f, err := strconv.ParseFloat(string(p), 64)

		result[i] = f
		return err
	})
	return result, err
}

// StringMap StringMap
func StringMap(reply interface{}, err error) (map[string]string, error) {
	values, err := Values(reply, err)
	if err != nil {
		return nil, err
	}
	if len(values)%2 != 0 {
		return nil, errors.New("redigo: StringMap expects even number of values result")
	}
	m := make(map[string]string, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, okKey := values[i].(string)
		value, okValue := values[i+1].(string)
		if !okKey || !okValue {
			return nil, errors.New("redigo: StringMap key not a bulk string value")
		}
		m[key] = value
	}
	return m, nil
}

// IntMap IntMap
func IntMap(reply interface{}, err error) (map[string]int, error) {
	values, err := Values(reply, err)
	if err != nil {
		return nil, err
	}
	if len(values)%2 != 0 {
		return nil, errors.New("redigo: IntMap expects even number of values result")
	}
	m := make(map[string]int, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, errors.New("redigo: IntMap key not a bulk string value")
		}
		value, err := Int(values[i+1], nil)
		if err != nil {
			return nil, err
		}
		m[string(key)] = value
	}
	return m, nil
}

// Int64Map Int64Map
func Int64Map(reply interface{}, err error) (map[string]int64, error) {
	values, err := Values(reply, err)
	if err != nil {
		return nil, err
	}
	if len(values)%2 != 0 {
		return nil, errors.New("redigo: Int64Map expects even number of values result")
	}
	m := make(map[string]int64, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, errors.New("redigo: Int64Map key not a bulk string value")
		}
		value, err := Int64(values[i+1], nil)
		if err != nil {
			return nil, err
		}
		m[string(key)] = value
	}
	return m, nil
}

// Bytes Bytes
func Bytes(reply interface{}, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}
	switch reply := reply.(type) {
	case []byte:
		return reply, nil
	case string:
		return []byte(reply), nil
	case nil:
		return nil, ErrNil
	case Error:
		return nil, reply
	}
	return nil, fmt.Errorf("redigo: unexpected type for Bytes, got type %T", reply)
}

// Error represents an error returned in a command reply.
type Error string

func (err Error) Error() string { return string(err) }

// ErrNil redis nil return
var ErrNil = errors.New("redigo: nil returned")

func sliceHelper(reply interface{}, err error, name string, makeSlice func(int), assign func(int, interface{}) error) error {
	if err != nil {
		return err
	}
	switch reply := reply.(type) {
	case []interface{}:
		makeSlice(len(reply))
		for i := range reply {
			if reply[i] == nil {
				continue
			}
			if err := assign(i, reply[i]); err != nil {
				return err
			}
		}
		return nil
	case nil:
		return ErrNil
	case Error:
		return reply
	}
	return fmt.Errorf("redigo: unexpected type for %s, got type %T", name, reply)
}

// Values is a helper that converts an array command reply to a []interface{}.
// If err is not equal to nil, then Values returns nil, err. Otherwise, Values
// converts the reply as follows:
//
//	Reply type      Result
//	array           reply, nil
//	nil             nil, ErrNil
//	other           nil, error
func Values(reply interface{}, err error) ([]interface{}, error) {
	if err != nil {
		return nil, err
	}
	switch reply := reply.(type) {
	case []interface{}:
		return reply, nil
	case nil:
		return nil, ErrNil
	case Error:
		return nil, reply
	}
	return nil, fmt.Errorf("redigo: unexpected type for Values, got type %T", reply)
}

func appendArgs(dst, src []interface{}) []interface{} {
	if len(src) == 1 {
		return appendArg(dst, src[0])
	}

	dst = append(dst, src...)
	return dst
}

func appendArg(dst []interface{}, arg interface{}) []interface{} {
	switch arg := arg.(type) {
	case []string:
		for _, s := range arg {
			dst = append(dst, s)
		}
		return dst
	case []interface{}:
		dst = append(dst, arg...)
		return dst
	case map[string]interface{}:
		for k, v := range arg {
			dst = append(dst, k, v)
		}
		return dst
	default:
		return append(dst, arg)
	}
}
