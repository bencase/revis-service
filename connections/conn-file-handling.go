package connections

import (
	"errors"
	"io/ioutil"
	"os"
	
	"gopkg.in/yaml.v2"
	
	"github.com/bencase/revis-service/util"
)

const filename = "conns.yml"

var ConnectionNotFoundError = errors.New("Could not find connection with that name")

func ReadConnections() ([]*Connection, error) {
	// If the file doesn't exist, return an empty list
	if fileDoesntExist() {
		return []*Connection{}, nil
	}
	
	// Get current contents of connections file
	inBytes, err := ioutil.ReadFile(filename)
	if err != nil { return []*Connection{}, err }
	conns := make([]*Connection, 0)
	err = yaml.Unmarshal(inBytes, &conns)
	return conns, err
}

// Gets the connection that has the provided name, looking instead at
// combined host and port on connections that have no name. If no such
// connection exists, returns nil without error.
func GetConnectionWithName(name string) (*Connection, error) {
	conns, err := ReadConnections()
	if err != nil {
		return nil, err
	}
	for _, conn := range conns {
		if conn.GetEffectiveName() == name {
			return conn, nil
		}
	}
	return nil, nil
}

func UpsertConnections(newConns []*Connection) error {
	// Get current contents of connections file
	prevConns, err := ReadConnections()
	if err != nil { return err }
	
	// Create a list of connections with all previous connections and the new ones
	allConns := make([]*Connection, 0)
	// Go through each previous connection. If there isn't a new connection with
	// the same name, add it.
	for _, conn := range prevConns {
		// Iterate through the new connections
		for _, newConn := range newConns {
			if conn.GetEffectiveName() != newConn.GetEffectiveName() {
				allConns = append(allConns, conn)
			}
		}
	}
	// Add all the new conns
	for _, newConn := range newConns {
		allConns = append(allConns, newConn)
	}
	
	// Rewrite all connections
	err = writeConnections(allConns)
	if err == nil {
		logger.Info("Upserted connections to file")
	} else {
		logger.Info("Error when attempting to upsert connections to file")
	}
	return err
}

func DeleteConnections(connNames []string) error {
	// If file doesn't exist, there's nothing to delete
	if fileDoesntExist() {
		return nil
	}
	
	// Get current contents of connections file
	prevConns, err := ReadConnections()
	if err != nil { return err }
	
	// Create a new slice of connections, with any connections having a name
	// in the provided string slice removed
	allConns := make([]*Connection, 0)
	for _, conn := range prevConns {
		if !util.ContainsString(connNames, conn.GetEffectiveName()) {
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

func writeConnections(conns []*Connection) error {
	// Create (or truncate if already exists) the file
	file, err := os.Create(filename)
	if err != nil { return err }
	
	// Write to the file
	outBytes, err := yaml.Marshal(conns)
	if err != nil { return err }
	_, err = file.Write(outBytes)
	return err
}