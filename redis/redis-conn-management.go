package redis

import (
	"errors"
	"time"

	"github.com/mediocregopher/radix.v2/redis"

	conns "github.com/bencase/revis-service/connections"
)


const defaultTimeout = time.Second * 4;


func getConn(host string, port string, password string, db int,
		timeout time.Duration) (*redis.Client, error) {
	dialFunc := getDialFunc(password, db, timeout)
	return dialFunc("tcp", host + ":" + port)
}
func getDialFunc(password string, db int, timeout time.Duration) func(network string,
		addr string) (*redis.Client, error) {
	return func(network string, addr string) (*redis.Client, error) {
		var client *redis.Client
		var err error
		if timeout >= 0 {
			client, err = redis.DialTimeout(network, addr, timeout)
		} else {
			client, err = redis.Dial(network, addr)
		}
		if err != nil {
			return nil, err
		}
		// If there's not a password or, just return the client
		if password == "" && db <= 0 {
			return client, nil
		}
		// If there is a password, perform auth with it
		if password != "" {
			resp := client.Cmd("AUTH", password)
			err = resp.Err
			if err != nil {
				client.Close()
				return nil, err
			}
		}
		// If it has a non-zero database, select it
		if db >= 1 {
			resp := client.Cmd("SELECT", db)
			err = resp.Err
			if err != nil {
				client.Close()
				return nil, err
			}
		}
		return client, nil
	}
}


func TestConn(conn *conns.Connection) error {
	client, err := getConn(conn.Host, conn.Port, conn.Password, conn.Db, defaultTimeout)
	if err != nil {
		return err
	}
	defer client.Close()
	resp := client.Cmd("PING")
	str, err := resp.Str()
	if err != nil {
		return err
	}
	if str != "PONG" {
		logger.Error("Unexpected connection test output:", str)
		return errors.New("Connection test gave unexpected result")
	}
	return nil
}