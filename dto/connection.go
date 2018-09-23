package dto

type Connection struct {
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	Host string `json:"host,omitempty" yaml:"host,omitempty"`
	Port string `json:"port,omitempty" yaml:"port,omitempty"`
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
	Db int `json:"db,omitempty" yaml:"db,omitempty"`
}