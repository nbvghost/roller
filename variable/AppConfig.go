package variable

import (
    "fmt"
    "strings"
)

/*type LogicServerConfig struct {
    IP      string
    TcpPort string
    DBUrl   string
    //RpcPort string
    //Offline bool
}*/
type RedisConfig struct {
    Host     string
    Password string
    DB       int
}

type ETCDConfig struct {
    Endpoints []string
    Username  string
    Password  string
}

type MySQLConfig struct {
    Host     string
    User     string
    Password string
    DBName   string
    Params   []string
}

func (mysql *MySQLConfig) DSN() string {

    return fmt.Sprintf("%v:%v@tcp(%v)/%v?%v", mysql.User, mysql.Password, mysql.Host, mysql.DBName, strings.Join(mysql.Params, "&"))
}

var AppConfig = struct {
    //LogicServer     []LogicServerConfig
    Debug           bool
    LogServer       string
    LogDir          string
    ETCD            ETCDConfig
    Redis           RedisConfig
    MySQL           MySQLConfig
    EnableHttps     bool
    KeyFile         string
    CertFile        string
    SessionExpires  int64
    PersistenceTime int64
    ServerPort      int
}{
    LogServer: "192.168.2.10:9090",
    LogDir:    "log",
    Debug:     true,
    ETCD: ETCDConfig{
        Endpoints: []string{"127.0.0.1:12379", "127.0.0.1:22379", "127.0.0.1:32379"},
        Username:  "",
        Password:  "",
    },
    Redis: RedisConfig{
        Host:     "127.0.0.1:6379",
        Password: "",
        DB:       0,
    },
    MySQL: MySQLConfig{
        Host:     "127.0.0.1:3306",
        User:     "root",
        Password: "274455411",
        DBName:   "test",
        Params: []string{
            "charset=utf8mb4",
            "collation=utf8mb4_unicode_ci",
            "parseTime=True",
            "loc=Local",
        },
    },

    SessionExpires:  1800,
    PersistenceTime: 180,
    EnableHttps:     false,
    CertFile:        "ssl/cert.pem",
    KeyFile:         "ssl/key.key",
}
