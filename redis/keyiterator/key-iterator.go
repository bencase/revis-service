package keyiterator

import (
	"errors"
	"io"

	rpool "github.com/mediocregopher/radix.v2/pool"
	"github.com/mediocregopher/radix.v2/redis"

	"revis-service/dto"
)

const defaultScanSize = 200

var NoMoreElements = errors.New("There are no further elements in this iterator")

type KeyIterator interface {
	HasNext() bool
	Next() (*dto.Key, error)
	io.Closer
}

type iKeyIterator struct {
	pool *rpool.Pool
	conn *redis.Client
	pattern string
	scanCursor int
	keysList []*dto.Key
	index int
	err error
}

func NewKeyIterator(pool *rpool.Pool, pattern string) (KeyIterator, error) {
	conn, err := pool.Get()
	if err != nil { return nil, err }
	initialCursorVal, keysList, err := getKeysList(conn, 0, pattern)
	if err != nil { return nil, err }
	keyIterator := &iKeyIterator{pool: pool,
		conn: conn,
		pattern: pattern,
		scanCursor: initialCursorVal,
		keysList: keysList,
		index: 0}
	return keyIterator, nil
}

func (this *iKeyIterator) HasNext() bool {
	return (this.index < len(this.keysList) || this.scanCursor != 0) && this.err == nil
}

func (this *iKeyIterator) Next() (*dto.Key, error) {

	if this.err != nil {
		return nil, errors.New("Iterator failed due to prior encountered error: " + this.err.Error())
	}

	var key *dto.Key
	if this.index >= len(this.keysList) {
		if this.scanCursor == 0 {
			return nil, NoMoreElements
		} else {
			err := this.refillKeyStrList()
			if err != nil {
				this.err = err
				return nil, err
			}
			key = this.keysList[0]
			this.index = 1
		}
	} else {
		key = this.keysList[this.index]
		this.index++
	}

	return key, nil
}

func (this *iKeyIterator) refillKeyStrList() error {
	newCursorVal, newKeys, err := getKeysList(this.conn, this.scanCursor, this.pattern)
	if err != nil { return err }
	this.scanCursor = newCursorVal
	this.keysList = newKeys
	return nil
}

func getKeysList(conn *redis.Client, scanCursor int, pattern string) (int, []*dto.Key, error) {

	var cursorVal int
	var keys []*dto.Key

	resp := conn.Cmd("scan",
		scanCursor,
		"match",
		pattern,
		"count",
		defaultScanSize)
	respSlice, err := resp.Array()
	if resp.Err != nil { return 0, keys, resp.Err }

	var keyStrs []string
	for i, resp := range respSlice {
		if i == 0 {
			// The first value of the resulting array will be the new cursorVal
			cursorVal, err = resp.Int()
			if err != nil { return cursorVal, keys, err }
		} else {
			newKeyStrs, err := getKeysFromArrayResp(resp)
			if err != nil { return cursorVal, keys, err }
			keyStrs = append(keyStrs, newKeyStrs...)
		}
	}

	for _, keyStr := range keyStrs {
		key := &dto.Key{}
		key.Key = keyStr
		keys = append(keys, key)
	}
	
	return cursorVal, keys, nil
}
func getKeysFromArrayResp(outerResp *redis.Resp) ([]string, error) {
	var keys []string
	resps, err := outerResp.Array()
	if err != nil { return keys, err }
	for _, resp := range resps {
		key, err := resp.Str()
		if err != nil { return keys, err }
		keys = append(keys, key)
	}
	return keys, nil
}

func (this *iKeyIterator) Close() error {
	this.pool.Put(this.conn)
	return nil
}