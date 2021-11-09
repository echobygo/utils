package DB

import (
	"MiaGame/Library/MiaLog"
	"MiaGame/Library/MiaCrypt"
	"context"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"os"
	"strings"
	"sync"
	"time"

)

type MysqlConfig struct {
	UserName string
	Password string
	Host     string
	Port     int
	DBName   string
	Args     string
	IdleConn int
	MaxConn  int
	ShowSQL  bool
	DetectionInterval int   //mysql heart
	AutoCreateTables []interface{} //需要初始化的表格在这个地方控制
}
// 当只连接一个数据源的时候，可以直接使用GormClient
// 否则应当自己持有管理InitGormDB返回的GormDB
//var GormClient *GormDB

type GormDB struct {
	dbConfig *MysqlConfig
	LockMutex     sync.RWMutex // lock
	Client   *gorm.DB     // mysql client
	DbConnStr  string
	ExitSystem  bool
	IsCheckConnect int    //0 需要重连， 1 正在重连， 2 重连成功
}
func (appconfig  *GormDB) GetMysqlConfig() *MysqlConfig{
	return appconfig.dbConfig
}

// 本方法会给GormClient赋值，多次调用GormClient指向最后一次调用的GormDB
func InitGormDB(dbConfig *MysqlConfig,ctx context.Context, encryptkey string) *GormDB {
	MiaLog.CInfo("starting db")
	if encryptkey!="" {
		dbConfig.Password ,_= MiaCrypt.StringDecrypt(dbConfig.Password,encryptkey)
	}
	GormClient := &GormDB{
		dbConfig:&MysqlConfig{
			UserName:dbConfig.UserName,
			Password:dbConfig.Password,
			Host:dbConfig.Host,
			Port:dbConfig.Port,
			DBName:dbConfig.DBName,
			Args:dbConfig.Args,
			IdleConn:dbConfig.IdleConn,
			MaxConn:dbConfig.MaxConn,
			ShowSQL:dbConfig.ShowSQL,
			DetectionInterval:dbConfig.DetectionInterval,
		},
	}
	//MiaLog.CInfo(len(dbConfig.AutoCreateTables))
	if len(dbConfig.AutoCreateTables) >0 {
		GormClient.dbConfig.AutoCreateTables = make([]interface{},len(dbConfig.AutoCreateTables));
	//	//for key,_:=  range dbConfig.AutoCreateTables {
	//	//	myDB.dbConfig.AutoCreateTables[key]= dbConfig.AutoCreateTables[key]
	//	//}
		copy(GormClient.dbConfig.AutoCreateTables ,dbConfig.AutoCreateTables[:])//第二个冒号 设置cap的
	//	MiaLog.CInfo("初始化长度单位：", len(myDB.dbConfig.AutoCreateTables));
	}

	dbCon := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbConfig.UserName, dbConfig.Password, dbConfig.Host, dbConfig.Port,dbConfig.DBName)
	MiaLog.CDebug(dbCon);
	GormClient.DbConnStr = dbCon
	var newLogger logger.Interface =nil
	MiaLog.CInfo(dbConfig.ShowSQL)
	if dbConfig.ShowSQL {
		newLogger= logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold: time.Second,                // 慢 SQL 阈值
				LogLevel:      setLogLevel("info"), // Log level
				Colorful:      false,                      // 禁用彩色打印
			},
		)
	}
	newLogger = newLogger
	/*
	,
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true},
			Logger:                 newLogger,
			PrepareStmt:            true, // 执行任何 SQL 时都创建并缓存预编译语句，可以提高后续的调用速度
			DisableAutomaticPing:   false,
			SkipDefaultTransaction: true, // 对于写操作（创建、更新、删除），为了确保数据的完整性，GORM 会将它们封装在事务内运行。但这会降低性能，你可以在初始化时禁用这种方式
			AllowGlobalUpdate:      false,
		}
	*/
	db, err := gorm.Open(mysql.Open(dbCon),&gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true},
		Logger:                 newLogger,
		PrepareStmt:            true, // 执行任何 SQL 时都创建并缓存预编译语句，可以提高后续的调用速度
		DisableAutomaticPing:   false,
		SkipDefaultTransaction: true, // 对于写操作（创建、更新、删除），为了确保数据的完整性，GORM 会将它们封装在事务内运行。但这会降低性能，你可以在初始化时禁用这种方式
		AllowGlobalUpdate:      false,
	})
	if err == nil {
		MiaLog.CInfo("Connect Database :  数据库链接成功 connecting db success!")
		GormClient.Client = db
		GormClient.initByDBConfigs()
		GormClient.IsCheckConnect = 2;
	}else{
		GormClient.IsCheckConnect=0
	}
	//GormClient = myDB //gormClient
	//myDB.autoCreateTable()
	go GormClient.timer(ctx)

	return GormClient
}
func GetDBOpertor(ptr *GormDB)*gorm.DB{
	return ptr.Client
}

func setLogLevel(logLevel string) logger.LogLevel {
	// 设置日志级别
	level := strings.Replace(strings.ToLower(logLevel), " ", "", -1)
	switch level {
	case "silent":
		return logger.Silent
	case "info":
		return logger.Info
	case "warn":
		return logger.Warn
	case "error":
		return logger.Error
	default:
		return logger.Silent
	}
}
//重连接
func (p *GormDB) reConnect() {

	//dbCon := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
	//	dbConfig.UserName, dbConfig.Password, dbConfig.Host, dbConfig.Port,dbConfig.DBName)
	//myDB.DbConnStr = dbCon
	var newLogger logger.Interface = nil
	if p.dbConfig.ShowSQL {
		MiaLog.CInfo("open sql")
		newLogger = logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold: time.Second,         // 慢 SQL 阈值
				LogLevel:      setLogLevel("info"), // Log level
				Colorful:      false,               // 禁用彩色打印
			},
		)
		newLogger  =logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold: time.Second,   // 慢 SQL 阈值
				LogLevel:      logger.Silent, // Log level
				Colorful:      false,         // 禁用彩色打印
			},
		)
	}

	db, err := gorm.Open(mysql.Open(p.DbConnStr), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true},
		Logger:                 newLogger,
		PrepareStmt:            true, // 执行任何 SQL 时都创建并缓存预编译语句，可以提高后续的调用速度
		DisableAutomaticPing:   false,
		SkipDefaultTransaction: true, // 对于写操作（创建、更新、删除），为了确保数据的完整性，GORM 会将它们封装在事务内运行。但这会降低性能，你可以在初始化时禁用这种方式
		AllowGlobalUpdate:      false,
	})
	if err == nil {
		p.IsCheckConnect = 2
		p.Client = db
		p.initByDBConfigs()
		MiaLog.CInfo("重连数据库成功 reconnect db success!")
	} else {
		p.IsCheckConnect = 0
	}
}

// 初始化参数
func (p *GormDB) initByDBConfigs() {
	sqlDb ,dbErr:=p.Client.DB()

	if dbErr != nil {
		MiaLog.Errorf("fail to connect database: %v\n", dbErr)
		//os.Exit(-1)
		return
	}
	if sqlDb!=nil {
		sqlDb.SetMaxIdleConns(p.dbConfig.IdleConn)
		sqlDb.SetMaxOpenConns(p.dbConfig.MaxConn)
		fmt.Println(p.dbConfig.MaxConn)
		fmt.Println(p.dbConfig.IdleConn)
		sqlDb.SetConnMaxLifetime(time.Duration(time.Second * 60))
	}

}

////auto create table
func (p *GormDB) autoCreateTable() {
	if p.IsCheckConnect ==2 {
		if p.dbConfig.AutoCreateTables == nil || len(p.dbConfig.AutoCreateTables) == 0 {
			return
		}
		MiaLog.CInfo("addr>>>>>", p.DbConnStr,"begin initAutoDB")
		//err:=	p.Client.AutoMigrate(p.dbConfig.AutoCreateTables...).Error()
		for _,item := range p.dbConfig.AutoCreateTables {
			p.autoCreate(item)
		}
		//p.autoCreate()
		//MiaLog.CInfo("create database ",err)

	}else{
		MiaLog.CInfo("未连接数据库，请检查连接")
	}

}


func (p *GormDB) timer(ctx context.Context) {
	if p.dbConfig.DetectionInterval <=0 {
		p.dbConfig.DetectionInterval = 30
	}
	timer1 := time.NewTicker(time.Duration(int64(p.dbConfig.DetectionInterval) * int64(time.Second)))
LOOP:
	for {
		select {
		case <-timer1.C:
			{
				if p.IsCheckConnect ==2{
					sqldata,_:=p.Client.DB()
					err := sqldata.Ping()
					if err != nil {
						MiaLog.CError("mysql connect fail,err:", err)
						MiaLog.CInfo("reconnect beginning...")
						p.IsCheckConnect = 1
						p.reConnect()
					}
				}else if p.IsCheckConnect ==0 {
					MiaLog.CInfo("正在重连数据库")
					p.reConnect()
				}

			}
		case <-ctx.Done():{
			sqldata,_:=p.Client.DB()
			sqldata.Close()
			break LOOP
		}
		}
	}
	MiaLog.CInfo("12312312312312312312")

}


//xi详细执行内容
func (p *GormDB) autoCreate(it interface{}) {
	defer func() { // 必须要先声明defer，否则不能捕获到panic异常
		if err := recover(); err != nil {
			fmt.Println(err) // 这里的err其实就是panic传入的内容
		}
	}()
	if p.Client == nil  || it==nil {
		MiaLog.CError("==================%v",it)
		return
	}
	err := p.Client.AutoMigrate(it).Error
	if err != nil {
		MiaLog.CError("auto create ",it," error",err)
	}
}
