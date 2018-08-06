package connections

type Connection struct {
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	Host string `json:"host,omitempty" yaml:"host,omitempty"`
	Port string `json:"port,omitempty" yaml:"port,omitempty"`
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
}
func (conn *Connection) GetEffectiveName() string {
	if (conn.Name != "") {
		return conn.Name
	} else {
		return conn.Host + ":" + conn.Port
	}
}