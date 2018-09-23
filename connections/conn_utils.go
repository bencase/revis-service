package connections

import (
	"strconv"

	"github.com/bencase/revis-service/dto"
)

func GetEffectiveNameOfConn(conn *dto.Connection) string {
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