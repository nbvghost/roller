package middleware

import (
    "github.com/jinzhu/gorm"
    _ "github.com/jinzhu/gorm/dialects/mysql"
    "github.com/nbvghost/glog"
    "github.com/nbvghost/gweb/thread"
    "github.com/nbvghost/roller/iface"
	"github.com/nbvghost/roller/variable"
	"time"
)

type MySQLConnector struct {
    db *gorm.DB
    //config *entity.AppConfig
    //logic  entity.LogicServerConfig
    models []iface.ISQLModels
}

func (connector *MySQLConnector) SQLModels() []iface.ISQLModels {
    return connector.models
}
func (connector *MySQLConnector) RegisterSQLModels(model iface.ISQLModels) {
    if connector.models == nil {
        connector.models = make([]iface.ISQLModels, 0)
    }
    connector.models = append(connector.models, model)
}
func (connector *MySQLConnector) createMySQL(logicIndex int) *MySQLConnector {
    //connector.config = config
    //connector.logic = config.LogicServer[logicIndex]
    connector.GetMySQL()
    return connector
}
func (connector *MySQLConnector) GetMySQL() *gorm.DB {
    //只有在程序第一次运行的时候才会为nil
    if connector.db == nil {
        connector.db = connector.retryOpenDB()
        thread.NewCoroutine(func() {
            for {
                //stats, _ := json.Marshal(connector.db.DB().Stats())
                //glog.Trace("MySQL-Stats：", string(stats))
                if glog.Error(connector.db.DB().Ping()) {
                    connector.db = connector.retryOpenDB()

                }
                time.Sleep(time.Second)
            }

        }, func(v interface{}, stack []byte) {

            glog.Trace(v)
            glog.Trace(string(stack))

        })
    }

    return connector.db
}
func (connector *MySQLConnector) retryOpenDB() *gorm.DB {
    //var err error
    _database, err := gorm.Open("mysql", variable.AppConfig.MySQL.DSN()) //keys.RemoteConfig.LogicServer[keys.LogicIndex].DBUrl)
    glog.Error(err)

    if variable.AppConfig.Debug {
        _database.LogMode(true)
        _database.Debug()
    }

    //todo:
    // SetMaxIdleCons 设置连接池中的最大闲置连接数。
    //_database.DB().SetMaxIdleConns(500)

    // SetMaxOpenCons 设置数据库的最大连接数量。
    //_database.DB().SetMaxOpenConns(0)

    // SetConnMaxLifetiment 设置连接的最大可复用时间。
    _database.DB().SetConnMaxLifetime(time.Second * time.Duration(variable.AppConfig.SessionExpires))
    return _database
}
