package dto

import (
	"encoding/json"
	
	conns "github.com/bencase/revis-service/connections"
)


/// Requests ///


type KeysRequest struct {
	ConnName string `json:"connName,omitempty"`
	MatchStr string `json:"matchStr,omitempty"`
	Limit int `json:"limit,omitempty"`
}


type UpsertConnectionsRequest struct {
	Connections []*conns.Connection `json:"connections"`
}
type DeleteConnectionsRequest struct {
	ConnectionNames []string `json:"connectionNames"`
}


/// Responses ///


type BaseResponse struct {
	ErrorContainer
}
func (this *BaseResponse) JsonBytes() ([]byte, error) {
	return json.Marshal(this)
}


type KeysResponse struct {
	Keys []*Key `json:"keys"`
	ErrorContainer
}
func (this *KeysResponse) JsonBytes() ([]byte, error) {
	return json.Marshal(this)
}
type Key struct {
	Key string `json:"key"`
	Val interface{} `json:"val"`
	Type string `json:"type,omitempty"`
}
type ZsetVal struct {
	Zval string `json:"zval"`
	Score float64 `json:"score"`
}
type HashVal struct {
	Hkey string `json:"hkey"`
	Hval string `json:"hval"`
}


type ConnectionsResponse struct {
	Connections []*conns.Connection `json:"connections"`
	ErrorContainer
}
func (this *ConnectionsResponse) JsonBytes() ([]byte, error) {
	return json.Marshal(this)
}


type DeleteResponse struct {
	Count int `json:"count"`
	DeletedAllKeys bool `json:"deletedAllKeys"`
	ErrorContainer
}
func (this *DeleteResponse) JsonBytes() ([]byte, error) {
	return json.Marshal(this)
}


type ErrorContainer struct {
	Error *ErrorResponse `json:"error,omitempty"`
}
type ErrorResponse struct {
	Message string `json:"message"`
}
