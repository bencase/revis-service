package connections

import (
	"strconv"
)

type Connection struct {
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	Host string `json:"host,omitempty" yaml:"host,omitempty"`
	Port string `json:"port,omitempty" yaml:"port,omitempty"`
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
	Db int `json:"db,omitempty" yaml:"db,omitempty"`
}
func (conn *Connection) GetEffectiveName() string {
	if conn.Name != "" {
		return conn.Name
	} else {
		name := conn.Host + ":" + conn.Port
		if conn.Db >= 1 {
			name = name + "[" + strconv.Itoa(conn.Db) + "]"
		}
		return name
	}
}