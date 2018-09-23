package dto

import (
	"encoding/json"
)


/// Requests ///


type KeysRequest struct {
	ConnName string `json:"connName,omitempty"`
	MatchStr string `json:"matchStr,omitempty"`
	Limit int `json:"limit,omitempty"`
}


type UpsertConnectionsRequest struct {
	Connections []*ConnUpsert `json:"connections"`
}
type ConnUpsert struct {
	OldConnName string `json:"oldConnName,omitempty"`
	NewConn *Connection `json:"newConn"`
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
	ExpAt int64 `json:"expAt,omitempty"`
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
	Connections []*Connection `json:"connections"`
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
