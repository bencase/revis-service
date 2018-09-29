package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	
	glogging "github.com/op/go-logging"
	
	rconns "github.com/bencase/revis-service/connections"
	"github.com/bencase/revis-service/dto"
	"github.com/bencase/revis-service/redis"
)

const ExposeHeadersHeader string = "Access-Control-Expose-Headers"
const ConnNameHeader string = "connname"
const PatternHeader string = "pattern"
const ScanIdHeader string = "scanid"

var logger = glogging.MustGetLogger("server")


type RedisServer struct {
	redisService *redis.RedisService
}


func NewRedisServer() (*RedisServer, error) {
	redisService := redis.NewRedisService()
	return &RedisServer{redisService: redisService}, nil
}


func (this *RedisServer) GetConnections(w http.ResponseWriter, r *http.Request) {
	defer recoverFromPanic(w, "GetConnections")
	w.Header().Add("Content-Type", "application/json")
	
	conns, err := rconns.ReadConnections()
	if err != nil {
		processError(w, "Error reading connections:", err)
		return
	}
	
	connsResp := &dto.ConnectionsResponse{Connections: conns}
	respBytes, err := connsResp.JsonBytes()
	if err != nil {
		processError(w, "Error marshalling connections list to json:", err)
		return
	}
	
	w.Write(respBytes)
}


func (this *RedisServer) UpsertConnections(w http.ResponseWriter,
		r *http.Request) {
	defer recoverFromPanic(w, "UpsertConnections")
	w.Header().Add("Content-Type", "application/json")
	
	reader := r.Body
	reqObj := new(dto.UpsertConnectionsRequest)
	err := json.NewDecoder(reader).Decode(reqObj)
	if err != nil {
		processError(w, "Error decoding json:", err)
		return
	}
	
	err = rconns.UpsertConnections(reqObj)
	if err != nil {
		processError(w, "Error upserting connections:", err)
		return
	}
	
	returnBaseResponse(w)
}


func (this *RedisServer) DeleteConnections(w http.ResponseWriter,
		r *http.Request) {
	defer recoverFromPanic(w, "DeleteConnections")
	w.Header().Add("Content-Type", "application/json")
	
	reader := r.Body
	reqObj := new(dto.DeleteConnectionsRequest)
	err := json.NewDecoder(reader).Decode(reqObj)
	if err != nil {
		processError(w, "Error decoding json:", err)
		return
	}
	
	err = rconns.DeleteConnections(reqObj.ConnectionNames)
	if err != nil {
		processError(w, "Error upserting connections:", err)
		return
	}
	
	returnBaseResponse(w)
}


func (this *RedisServer) TestConnection(w http.ResponseWriter,
		r *http.Request) {
	defer recoverFromPanic(w, "TestConnection")
	w.Header().Add("Content-Type", "application/json")
	
	reader := r.Body
	conn := new(dto.Connection)
	err := json.NewDecoder(reader).Decode(conn)
	if err != nil {
		processError(w, "Error decoding json:", err)
		return
	}

	err = redis.TestConn(conn)
	if err != nil {
		processError(w, "Connection error:", err)
		return
	}
	
	returnBaseResponse(w)
}


func (this *RedisServer) GetKeysWithValues(w http.ResponseWriter,
		r *http.Request) {
	defer recoverFromPanic(w, "GetKeysWithValues")
	// Route the request based on whether it has a scanId or not
	scanId := r.Header.Get(ScanIdHeader)
	if scanId != "" && scanId != "0" {
		this.getMoreKeys(w, r)
	} else {
		this.startGettingKeysWithValues(w, r)
	}
}


func (this *RedisServer) startGettingKeysWithValues(w http.ResponseWriter,
		r *http.Request) {
	defer recoverFromPanic(w, "startGettingKeysWithValues")
	w.Header().Add("Content-Type", "application/json")

	connName := r.Header.Get(ConnNameHeader)
	if connName == "" {
		processError(w, "Error parsing header:",
			errors.New("Header does not contain connection name"))
		return
	}
	pattern := r.Header.Get(PatternHeader)

	keys, scanId, hasMoreKeys, err := this.redisService.
		StartGettingKeysWithValues(connName, pattern)
	if err != nil {
		processError(w, "Error getting keys and values:", err)
		return
	}
	respondWithKeys(w, keys, scanId, hasMoreKeys)
}


func (this *RedisServer) getMoreKeys(w http.ResponseWriter, r *http.Request) {
	defer recoverFromPanic(w, "getMoreKeys")
	w.Header().Add("Content-Type", "application/json")

	scanIdStr := r.Header.Get(ScanIdHeader)
	scanId, err := strconv.Atoi(scanIdStr)
	if err != nil {
		processError(w, "Error parsing scan ID from header:", err)
		return
	}

	keys, scanId, hasMoreKeys, err := this.redisService.
		GetNextKeys(scanId)
	if err != nil {
		processError(w, "Error getting keys and values:", err)
		return
	}
	respondWithKeys(w, keys, scanId, hasMoreKeys)
}


func respondWithKeys(w http.ResponseWriter, keys []*dto.Key, scanId int, hasMoreKeys bool) {
	if hasMoreKeys {
		w.Header().Add(ExposeHeadersHeader, ScanIdHeader)
		w.Header().Set(ScanIdHeader, strconv.Itoa(scanId))
		w.WriteHeader(202)
	} else {
		w.WriteHeader(200)
	}
	keysResp := &dto.KeysResponse{Keys: keys}
	respBytes, err := json.Marshal(keysResp)
	if err != nil {
		processError(w, "Error marshalling keys and values to json:", err)
		return
	}
	w.Write(respBytes)
}


func (this *RedisServer) DeleteKeysMatchingPattern(w http.ResponseWriter,
		r *http.Request) {
	defer recoverFromPanic(w, "DeleteKeysMatchingPattern")
	w.Header().Add("Content-Type", "application/json")

	connName := r.Header.Get(ConnNameHeader)
	if connName == "" {
		processError(w, "Error parsing header:",
			errors.New("Header does not contain connection name"))
		return
	}
	pattern := r.Header.Get(PatternHeader)

	deletedAllKeys, count, err := this.redisService.DeleteKeysMatchingPattern(connName, pattern)
	if err != nil {
		processError(w, fmt.Sprintf("Error deleting keys matching pattern %[1]v: ",
			pattern), err)
		return
	}
	
	delResp := &dto.DeleteResponse{DeletedAllKeys: deletedAllKeys, Count: count}
	respBytes, err := delResp.JsonBytes()
	if err != nil {
		processError(w, "Error marshalling delete response to json:", err)
		return
	}

	w.Write(respBytes)
}


func (this *RedisServer) Close() error {
	return this.redisService.Close()
}


func processError(w http.ResponseWriter, logMessagePrefix string, err error) {
	logger.Error(logMessagePrefix, err)
	w.WriteHeader(500)
	//message := "There was an error processing the request"
	message := err.Error()
	errResp := &dto.ErrorResponse{Message: message}

	respObj := &dto.BaseResponse{}
	respObj.Error = errResp
	
	jsonBytes, err := respObj.JsonBytes()
	if err != nil { logger.Error("Error getting bytes from json:", err) }
	w.Write(jsonBytes)
}


func returnBaseResponse(w http.ResponseWriter) {
	respObj := &dto.BaseResponse{}
	respBytes, err := respObj.JsonBytes()
	if err != nil {
		processError(w, "Error marshalling response object to json:", err)
		return
	}
	
	w.Write(respBytes)
}

func recoverFromPanic(w http.ResponseWriter, fname string) {
	if r := recover(); r != nil {
		var err error
		switch rtyp := r.(type) {
		case string : err = errors.New(rtyp)
		case error : err = rtyp
		default : err = errors.New("Panic is unknown type")
		}
		processError(w, fmt.Sprintf("Panic in %[1]v:", fname), err)
	}
}


func (this *RedisServer) SayHello(w http.ResponseWriter, r *http.Request) {
	defer recoverFromPanic(w, "SayHello")

	message := "Hello, World!"

	w.Write([]byte(message))
}