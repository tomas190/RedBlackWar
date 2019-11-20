package conf

import (
	"encoding/json"
	"github.com/name5566/leaf/log"
	"io/ioutil"
)

var Server struct {
	LogLevel    string
	LogPath     string
	WSAddr      string
	Port        string
	CertFile    string
	KeyFile     string
	TCPAddr     string
	MaxConnNum  int
	ConsolePort int
	ProfilePath string

	MongoDBAddr string
	MongoDBAuth string
	MongoDBUser string
	MongoDBPwd  string

	TokenServer      string
	CenterServer     string
	CenterServerPort string
	DevKey           string
	DevName          string
	GameID           string
	CenterUrl        string

	LogAddr string
	RandNum string
}

func init() {
	data, err := ioutil.ReadFile("conf/server.json")
	if err != nil {
		log.Fatal("%v", err)
	}
	err = json.Unmarshal(data, &Server)
	if err != nil {
		log.Fatal("%v", err)
	}
}
