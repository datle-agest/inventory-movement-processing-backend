package gormc

import (
	"errors"
	"flag"
	"fmt"
	"inventory-movement-processing/pkg/components/gormc/dialets"
	"inventory-movement-processing/pkg/logger"
	sctx "inventory-movement-processing/pkg/service_context"
	"time"

	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

type GormDBType int

const (
	GormDBTypePostgres GormDBType = iota + 1
	GormDBTypeSqlite
	GormDBTypeNotSupported
)

type GormOpt struct {
	dsn                   string
	dbType                string
	maxOpenConnections    int
	maxIdleConnections    int
	maxConnectionIdleTime int
}

type gormDB struct {
	id       string
	prefix   string
	logger   logger.Logger
	logLevel string
	db       *gorm.DB
	*GormOpt
}

func NewGormDB(id, prefix string) *gormDB {
	return &gormDB{
		GormOpt: new(GormOpt),
		id:      id,
		prefix:  prefix,
	}
}

func (gdb *gormDB) ID() string {
	return gdb.id
}

func (gdb *gormDB) InitFlags() {
	prefix := gdb.prefix

	if prefix != "" {
		prefix += "-"
	}

	flag.StringVar(
		&gdb.dsn,
		fmt.Sprintf("%sdb-dsn", prefix),
		"",
		"Database dsn",
	)

	flag.StringVar(
		&gdb.dbType,
		fmt.Sprintf("%sdb-driver", prefix),
		"mysql",
		"Database driver (mysql, postgres) - Default mysql",
	)

	flag.IntVar(
		&gdb.maxOpenConnections,
		fmt.Sprintf("%sdb-max-conn", prefix),
		30,
		"maximum number of open connections to the database - Default 30",
	)

	flag.IntVar(
		&gdb.maxIdleConnections,
		fmt.Sprintf("%sdb-max-idle-conn", prefix),
		10,
		"maximum number of database connections in the idle - Default 10",
	)

	flag.IntVar(
		&gdb.maxConnectionIdleTime,
		fmt.Sprintf("%sdb-max-conn-idle-time", prefix),
		3600,
		"maximum amount of time a connection may be idle in seconds - Default 3600",
	)
}

func (gdb *gormDB) Activate(serviceCtx sctx.ServiceContext) error {
	gdb.logger = serviceCtx.Logger(gdb.id)
	gdb.logLevel = serviceCtx.LogLevel()

	gdb.logger.Infof(
		"gorm initialized (db=%s, log_level=%s)",
		gdb.dbType,
		gdb.logLevel,
	)

	dbType := getDBType(gdb.dbType)

	if dbType == GormDBTypeNotSupported {
		return errors.New("Database type not supported")
	}

	gdb.logger.Info("Connecting to database...")

	conn, err := gdb.getDBConn(dbType)

	if err != nil {
		gdb.logger.Error("cannot connect to database", err.Error())
		return err
	}

	gdb.db = conn

	return nil
}

func (gdb *gormDB) Stop() error {
	if gdb.db == nil {
		return nil
	}

	sqlDB, err := gdb.db.DB()

	if err != nil {
		return err
	}

	gdb.logger.Info("closing database connection...")
	return sqlDB.Close()
}

func (gdb *gormDB) GetDB() *gorm.DB {
	if gdb.logLevel == "debug" {
		return gdb.db.Session(&gorm.Session{NewDB: true}).Debug()
	}

	newSessionDB := gdb.db.Session(&gorm.Session{
		NewDB:  true,
		Logger: gdb.db.Logger.LogMode(gormLogger.Silent),
	})

	if sqlDB, err := newSessionDB.DB(); err == nil {
		sqlDB.SetMaxOpenConns(gdb.maxOpenConnections)
		sqlDB.SetMaxIdleConns(gdb.maxIdleConnections)
		sqlDB.SetConnMaxIdleTime(
			time.Second * time.Duration(gdb.maxConnectionIdleTime),
		)
	}

	return newSessionDB
}

func getDBType(dbType string) GormDBType {
	switch dbType {
	case "postgres":
		return GormDBTypePostgres
	case "sqlite":
		return GormDBTypeSqlite
	default:
		return GormDBTypeNotSupported
	}
}

func (gdb *gormDB) getDBConn(t GormDBType) (dbConn *gorm.DB, err error) {
	switch t {
	case GormDBTypePostgres:
		return dialets.PostgresDB(gdb.dsn)
	}
	return nil, errors.New("invalid dsn")
}
