package redis

import (
	"errors"

	"revis-service/dto"
)

type RedisService struct {
	cmdRunnerRegister *CmdRunnerRegister
	scanIdChanMap map[int]*chanContainer
}

const defaultLimit = 200
const maxTotalKeysPerScan = 2100

// It starts at 1 instead of 0 since a 0 may omit the value from the json
var scanId = 1

func NewRedisService() *RedisService {
	cmdRunnerRegister := NewRegister()
	scanIdChanMap := make(map[int]*chanContainer)
	redisService := &RedisService{cmdRunnerRegister: cmdRunnerRegister,
		scanIdChanMap: scanIdChanMap}
	return redisService
}

// The bool returned by this function will be true if there are more keys yet to come,
// or false if there will be no more keys.
func (this *RedisService) StartGettingKeysWithValues(connName string, pattern string) ([]*dto.Key,
		int, bool, error) {

	id := scanId
	scanId++
	
	cmdRunner, err := this.cmdRunnerRegister.GetCmdRunner(connName)
	if err != nil { return []*dto.Key{}, id, true, err }
	
	keyChan := make(chan []*dto.Key, maxTotalKeysPerScan / defaultLimit)
	finalChan := make(chan []*dto.Key)
	errChan := make(chan error)

	go cmdRunner.GetKeysWithValues(pattern, keyChan, finalChan, errChan)

	chans := &chanContainer{keyChan: keyChan,
		finalChan: finalChan,
		errChan: errChan}
	keys, hasMoreKeys, err := chans.getNextKeys()
	if hasMoreKeys {
		this.scanIdChanMap[id] = chans
	}
	return keys, id, hasMoreKeys, err
}
// The bool returned by this function will be true if there are more keys yet to come,
// or false if there will be no more keys.
func (this *RedisService) GetNextKeys(id int) ([]*dto.Key, int, bool, error) {
	chans := this.scanIdChanMap[id]
	keys, hasMoreKeys, err := chans.getNextKeys()
	if !hasMoreKeys {
		delete(this.scanIdChanMap, id)
	}
	return keys, id, hasMoreKeys, err
}

func (this *RedisService) Close() error {
	return this.cmdRunnerRegister.Close()
}

type chanContainer struct {
	keyChan <-chan []*dto.Key
	finalChan <-chan []*dto.Key
	errChan <-chan error
}
// The bool returned by this function will be true if there are more keys yet to come,
// or false if there will be no more keys.
func (this *chanContainer) getNextKeys() ([]*dto.Key, bool, error) {

	// It first tries to read from the key chan
	if keys, ok := <-this.keyChan; ok {
		return keys, true, nil
	}

	// If that channel is closed, it then will wait until it gets either a value
	// or a closed indicator from both the finalChan and the errChan
	hasReceivedFromFinalChan := false
	hasReceivedFromErrChan := false
	for !(hasReceivedFromFinalChan && hasReceivedFromErrChan) {
		select {
		case keys, ok := <-this.finalChan:
			if ok {
				return keys, false, nil
			} else {
				hasReceivedFromFinalChan = true
				this.finalChan = nil
			}
		case chanErr, ok := <-this.errChan:
			if ok {
				return []*dto.Key{}, false, chanErr
			} else {
				hasReceivedFromErrChan = true
				this.errChan = nil
			}
		}
	}
	
	return []*dto.Key{}, false, errors.New("All scan channels are closed")
}