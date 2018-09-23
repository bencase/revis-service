package connections

import (
	"errors"
	"io/ioutil"
	"os"
	
	"gopkg.in/yaml.v2"
	
	"github.com/bencase/revis-service/dto"
	"github.com/bencase/revis-service/util"
	"github.com/bencase/revis-service/connections/encrypt"
)

const filename = "conns.yml"

var ConnectionNotFoundError = errors.New("Could not find connection with that name")

func readConnectionsNoDecrypt() ([]*dto.Connection, error) {
	// If the file doesn't exist, return an empty list
	if fileDoesntExist() {
		return []*dto.Connection{}, nil
	}
	
	// Get current contents of connections file
	inBytes, err := ioutil.ReadFile(filename)
	if err != nil { return []*dto.Connection{}, err }
	conns := make([]*dto.Connection, 0)
	err = yaml.Unmarshal(inBytes, &conns)
	return conns, err
}
func ReadConnections() ([]*dto.Connection, error) {
	conns, err := readConnectionsNoDecrypt()
	if err != nil { return []*dto.Connection{}, err }

	// Decrypt passwords
	for _, conn := range conns {
		if conn.Password != "" {
			decryptedPwd, err := encrypt.DecryptFromBase64(conn.Password)
			if err != nil { return []*dto.Connection{}, err }
			conn.Password = decryptedPwd
		}
	}
	return conns, err
}

// Gets the connection that has the provided name, looking instead at
// combined host and port on connections that have no name. If no such
// connection exists, returns nil without error.
func GetConnectionWithName(name string) (*dto.Connection, error) {
	conns, err := ReadConnections()
	if err != nil {
		return nil, err
	}
	for _, conn := range conns {
		if GetEffectiveNameOfConn(conn) == name {
			return conn, nil
		}
	}
	return nil, nil
}

func UpsertConnections(reqObj *dto.UpsertConnectionsRequest) error {
	// Get current contents of connections file
	conns, err := readConnectionsNoDecrypt()
	if err != nil { return err }
	// Upsert connections
	for _, connUpsert := range reqObj.Connections {
		newConn := connUpsert.NewConn
		replace := connUpsert.OldConnName
		err = upsertConnection(newConn, replace, &conns)
		if err != nil { return err }
	}
	// Rewrite all connections
	err = writeConnections(conns)
	if err == nil {
		logger.Info("Upserted connections to file")
	} else {
		logger.Info("Error when attempting to upsert connections to file")
	}
	return err
}
func upsertConnection(newConn *dto.Connection, replace string, pConns *[]*dto.Connection) error {
	encryptedNewConn, err := getConnWithPasswordEncrypted(newConn)
	if err != nil { return err }
	conns := *pConns
	hasReplacedConn := false
	if replace != "" {
		indexOfConnToReplace := -1
		for i, originalConn := range conns {
			if replace == GetEffectiveNameOfConn(originalConn) {
				indexOfConnToReplace = i
				break
			}
		}
		if indexOfConnToReplace >= 0 {
			conns[indexOfConnToReplace] = encryptedNewConn
			hasReplacedConn = true
		}
	}
	if !hasReplacedConn {
		conns = append(conns, encryptedNewConn)
	}
	*pConns = conns
	return nil
}
func getConnsWithPasswordsEncrypted(origConns []*dto.Connection) ([]*dto.Connection, error) {
	encryptedConns := make([]*dto.Connection, 0)
	for _, conn := range origConns {
		encryptedConn, err := getConnWithPasswordEncrypted(conn)
		if err != nil { return nil, err }
		encryptedConns = append(encryptedConns, encryptedConn)
	}
	return encryptedConns, nil
}
func getConnWithPasswordEncrypted(conn *dto.Connection) (*dto.Connection, error) {
	encryptedConn := *conn
	if encryptedConn.Password != "" {
		encryptedPwd, err := encrypt.EncryptToBase64(encryptedConn.Password)
		if err != nil { return nil, err }
		encryptedConn.Password = encryptedPwd
	}
	return &encryptedConn, nil
}

func DeleteConnections(connNames []string) error {
	// If file doesn't exist, there's nothing to delete
	if fileDoesntExist() {
		return nil
	}
	
	// Get current contents of connections file
	prevConns, err := readConnectionsNoDecrypt()
	if err != nil { return err }
	
	// Create a new slice of connections, with any connections having a name
	// in the provided string slice removed
	allConns := make([]*dto.Connection, 0)
	for _, conn := range prevConns {
		if !util.ContainsString(connNames, GetEffectiveNameOfConn(conn)) {
			allConns = append(allConns, conn)
		}
	}
	
	// Re-write the file
	err = writeConnections(allConns)
	if err == nil {
		logger.Info("Deleted connections from file")
	} else {
		logger.Info("Error when attempting to delete connections from file")
	}
	return err
}

func fileDoesntExist() bool {
	_, err := os.Stat(filename)
	return os.IsNotExist(err)
}

func writeConnections(conns []*dto.Connection) error {
	// Create (or truncate if already exists) the file
	file, err := os.Create(filename)
	if err != nil { return err }
	
	// Write to the file
	outBytes, err := yaml.Marshal(conns)
	if err != nil { return err }
	_, err = file.Write(outBytes)
	return err
}