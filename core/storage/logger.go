package storage

import (
	"context"
	"time"

	"github.com/evcc-io/evcc/util"
	"gorm.io/gorm/logger"
)

type adapter struct {
	log *util.Logger
}

func (l *adapter) LogMode(_ logger.LogLevel) logger.Interface {
	return l
}

func (l *adapter) Info(_ context.Context, format string, args ...interface{}) {
	l.log.INFO.Printf(format, args...)
}

func (l *adapter) Warn(_ context.Context, format string, args ...interface{}) {
	l.log.WARN.Printf(format, args...)
}

func (l *adapter) Error(_ context.Context, format string, args ...interface{}) {
	l.log.ERROR.Printf(format, args...)
}

func (l *adapter) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if err != nil {
		sql, _ := fc()
		l.log.ERROR.Printf("%v: %s", err, sql)
	}
}
