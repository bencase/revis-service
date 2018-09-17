package redis

import (
	"time"
	
	"github.com/bencase/revis-service/connections"
)

const poolDuration = 31 * time.Minute

type CmdRunnerRegister struct {
	cmdRunnerMap map[string]RedisCmdRunner
	timersMap map[string]*time.Timer
}

func NewRegister() *CmdRunnerRegister {
	register := &CmdRunnerRegister{}
	register.cmdRunnerMap = make(map[string]RedisCmdRunner)
	register.timersMap = make(map[string]*time.Timer)
	return register
}

func (crr *CmdRunnerRegister) GetCmdRunner(name string) (RedisCmdRunner,
		error) {
	if _, hasCmdRunner := crr.cmdRunnerMap[name]; hasCmdRunner {
		return crr.getExistingCmdRunner(name)
	} else {
		return crr.createCmdRunner(name)
	}
}

func (crr *CmdRunnerRegister) createCmdRunner(name string) (RedisCmdRunner,
		error) {

	conn, err := connections.GetConnectionWithName(name);
	if err != nil { return nil, err }
	
	cmdRunner, err := getCmdRunner(conn.Host, conn.Port, conn.Password, conn.Db)
	if err != nil { return nil, err }
	crr.cmdRunnerMap[name] = cmdRunner

	timer := time.NewTimer(poolDuration)
	crr.timersMap[name] = timer
	go func() {
		<-timer.C
		crr.CloseCmdRunner(name)
	}()

	return cmdRunner, nil
}

// This method assumes a check has already been done to confirm that
// the register contains a CmdRunner with this name.
func (crr *CmdRunnerRegister) getExistingCmdRunner(name string) (RedisCmdRunner, error) {

	timer := crr.timersMap[name]
	timer.Reset(poolDuration)

	return crr.cmdRunnerMap[name], nil
}

func (crr *CmdRunnerRegister) CloseCmdRunner(name string) error {
	cmdRunner, hasKey := crr.cmdRunnerMap[name]
	if !hasKey {
		return nil
	}
	delete(crr.cmdRunnerMap, name)

	timer, hasKey := crr.timersMap[name]
	if hasKey {
		timer.Stop()
		delete(crr.timersMap, name)
	}

	return cmdRunner.Close()
}

func (crr *CmdRunnerRegister) Close() error {
	for _, cmdRunner := range crr.cmdRunnerMap {
		cmdRunner.Close()
	}
	return nil
}