package redis

import (
	"io"
	"strconv"
	
	"github.com/mediocregopher/radix.v2/redis"
	rpool "github.com/mediocregopher/radix.v2/pool"

	"github.com/bencase/revis-service/dto"
	ki "github.com/bencase/revis-service/redis/keyiterator"
)

const defaultScanSize = 2000
// Key types:
const (
	typeString = "string"
	typeList = "list"
	typeSet = "set"
	typeZset = "zset"
	typeHash = "hash"
)

type RedisCmdRunner interface {
	io.Closer
	GetValuesOfKeys(keys []string) (map[string]string, error)
	GetKeysWithValues(pattern string, keyChan chan<- []*dto.Key,
		finalChan chan<- []*dto.Key, errorChan chan<- error)
}

type iRedisCmdRunner struct {
	pool *rpool.Pool
}

func getCmdRunner(host string, port string, password string) (RedisCmdRunner, error) {
	pool, err := rpool.NewCustom("tcp", host + ":" + port, 10, getDialFunc(password))
	if err != nil {
		return nil, err
	}
	return &iRedisCmdRunner{pool: pool}, nil
}
func getDialFunc(password string) func(network string, addr string) (*redis.Client, error) {
	return func(network string, addr string) (*redis.Client, error) {
		client, err := redis.Dial(network, addr)
		if err != nil {
			return nil, err
		}
		// If there's not a password, just return the client
		if password == "" {
			return client, nil
		}
		// If there is a password, perform auth with it
		resp := client.Cmd("AUTH", password)
		err = resp.Err
		if err != nil {
			client.Close()
			return nil, err
		}
		return client, nil
	}
}

func (this *iRedisCmdRunner) GetValuesOfKeys(keys []string) (map[string]string, error) {

	keyValsMap := make(map[string]string)
	
	conn, err := this.pool.Get()
	if err != nil { return nil, err }
	defer this.pool.Put(conn)

	for _, key := range keys {
		conn.PipeAppend("GET", key)
	}

	resps, err := getResponsesFromPipeline(conn)
	if err != nil { return nil, err }
	for i, resp := range resps {
		val, err := resp.Str()
		if err != nil { return nil, err }
		keyValsMap[keys[i]] = val
	}

	return keyValsMap, nil
}

func (this *iRedisCmdRunner) GetKeysWithValues(pattern string, keyChan chan<- []*dto.Key,
		finalChan chan<- []*dto.Key, errorChan chan<- error) {

	keyIterator, err := ki.NewKeyIterator(this.pool, pattern)
	if err != nil {
		pushErrorToErrorChan(err, keyChan, finalChan, errorChan)
		return
	}
	defer keyIterator.Close()
	
	var keyChunk []*dto.Key
	keysScanned := 0
	for keyIterator.HasNext() && keysScanned < maxTotalKeysPerScan {
		key, err := keyIterator.Next()
		if err != nil {
			pushErrorToErrorChan(err, keyChan, finalChan, errorChan)
			return
		}
		keyChunk = append(keyChunk, key)
		if len(keyChunk) >= defaultLimit {
			err = this.getMetadataAndValuesForKeys(keyChunk)
			if err != nil {
				pushErrorToErrorChan(err, keyChan, finalChan, errorChan)
				return
			}
			if keyIterator.HasNext() {
				keyChan <- keyChunk
				keyChunk = make([]*dto.Key, 0)
			} else {
				close(keyChan)
				close(errorChan)
				finalChan <- keyChunk
				close(finalChan)
				return
			}
		}
		keysScanned++
	}
	if len(keyChunk) > 0 {
		err = this.getMetadataAndValuesForKeys(keyChunk)
		if err != nil {
			pushErrorToErrorChan(err, keyChan, finalChan, errorChan)
			return
		}
		close(keyChan)
		close(errorChan)
		finalChan <- keyChunk
		close(finalChan)
		return
	}
}
func (this *iRedisCmdRunner) getMetadataAndValuesForKeys(keys []*dto.Key) error {
	conn, err := this.pool.Get()
	if err != nil { return err }
	defer this.pool.Put(conn)
	err = this.addTypesForKeys(conn, keys)
	if err != nil { return err }
	err = this.addValuesForKeys(conn, keys)
	return err
}
func (this *iRedisCmdRunner) addTypesForKeys(conn *redis.Client, keys []*dto.Key) error {
	// Add commands to pipeline
	for _, key := range keys {
		conn.PipeAppend("TYPE", key.Key)
	}
	// Get responses off pipeline
	resps, err := getResponsesFromPipeline(conn)
	if err != nil { return err }
	for i, resp := range resps {
		typ, err := resp.Str()
		if err != nil { return err }
		// If the type is "string", then don't add the type, since this will be the default
		if typ != typeString {
			keys[i].Type = typ
		}
	}
	return nil
}
func (this *iRedisCmdRunner) addValuesForKeys(conn *redis.Client, keys []*dto.Key) error {
	// Add commands to pipeline
	for _, key := range keys {
		// Add the appropriate command for the key's type
		switch key.Type {
		case "" : conn.PipeAppend("GET", key.Key)
		case typeList : conn.PipeAppend("LRANGE", key.Key, 0, -1)
		case typeSet : conn.PipeAppend("SMEMBERS", key.Key)
		case typeZset : conn.PipeAppend("ZRANGEBYSCORE", key.Key, "-inf", "+inf", "WITHSCORES")
		case typeHash : conn.PipeAppend("HGETALL", key.Key)
		default : conn.PipeAppend("GET", key.Key)
		}
	}
	// Get responses off pipeline
	resps, err := getResponsesFromPipeline(conn)
	if err != nil { return err }
	for i, resp := range resps {
		key := keys[i]
		switch key.Type {
		case "" : err = this.getValForStringKey(key, resp)
		case typeList : err = this.getValForListOrSetKey(key, resp)
		case typeSet : err = this.getValForListOrSetKey(key, resp)
		case typeZset : err = this.getValForZsetKey(key, resp)
		case typeHash : err = this.getValForHashKey(key, resp)
		default : err = this.getValForStringKey(key, resp)
		}
		if err != nil { return err }
	}
	return nil
}
func (this *iRedisCmdRunner) getValForStringKey(key *dto.Key, resp *redis.Resp) error {
	val, err := resp.Str()
	if err != nil { return err }
	key.Val = val
	return nil
}
func (this *iRedisCmdRunner) getValForListOrSetKey(key *dto.Key, resp *redis.Resp) error {
	vals, err := resp.List()
	if err != nil { return err }
	key.Val = vals
	return nil
}
func (this *iRedisCmdRunner) getValForZsetKey(key *dto.Key, resp *redis.Resp) error {
	vals, err := resp.List()
	if err != nil { return err }
	var zvals []*dto.ZsetVal
	// The vals list will be in pairs. The first value will be the member, the second the score.
	for i := 0; i < len(vals); i = i + 2 {
		score, err := strconv.ParseFloat(vals[i + 1], 64)
		if err != nil { return err }
		zval := &dto.ZsetVal{Score: score, Zval: vals[i]}
		zvals = append(zvals, zval)
	}
	key.Val = zvals
	return nil
}
func (this *iRedisCmdRunner) getValForHashKey(key *dto.Key, resp *redis.Resp) error {
	vals, err := resp.List()
	if err != nil { return err }
	var hvals []*dto.HashVal
	// The vals list will be in pairs. The first will be the key, the second the value.
	for i := 0; i < len(vals); i = i + 2 {
		key := vals[i]
		val := vals[i + 1]
		hval := &dto.HashVal{Hkey: key, Hval: val}
		hvals = append(hvals, hval)
	}
	key.Val = hvals
	return nil
}
func pushErrorToErrorChan(err error, keyChan chan<- []*dto.Key, finalChan chan<- []*dto.Key,
		errorChan chan<- error) {
	close(keyChan)
	close(finalChan)
	errorChan <- err
	close(errorChan)
	return
}

func (this *iRedisCmdRunner) Close() error {
	this.pool.Empty()
	return nil
}


func getResponsesFromPipeline(conn *redis.Client) ([]*redis.Resp, error) {
	resps := []*redis.Resp{}
	resp := conn.PipeResp()
	for resp.Err == nil {
		resps = append(resps, resp)
		resp = conn.PipeResp()
	}
	if resp.Err != redis.ErrPipelineEmpty {
		return []*redis.Resp{}, resp.Err
	}
	return resps, nil
}


type keySet map[string]bool
func (this *keySet) add(str string) {
	mp := *this
	mp[str] = true
}
func (this *keySet) remove(str string) {
	this.remove(str)
}
func (this *keySet) contains(str string) bool {
	mp := *this
	return mp[str]
}
func (this *keySet) toArray() []string {
	mp := *this
	keys := make([]string, len(mp))
	i := 0
	for key := range mp {
		keys[i] = key
		i++
	}
	return keys
}