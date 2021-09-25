package storage

import (
	"time"

	"github.com/evcc-io/evcc/util"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Record struct {
	ID            uint64    `gorm:"id,uniqueIndex"`
	StartTime     time.Time `gorm:"start_time"`
	EndTime       time.Time `gorm:"end_time"`
	Loadpoint     int       `gorm:"loadpoint"`
	StartSoc      float64   `gorm:"start_soc"`
	EndSoc        float64   `gorm:"end_soc"`
	Vehicle       string    `gorm:"vehicle"`
	ChargedEnergy float64   `gorm:"charged_energy"`
	GridEnergy    float64   `gorm:"grid_energy"`
}

var db *gorm.DB

func Open() error {
	instance, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		return err
	}
	db = instance

	db.Logger = &adapter{log: util.NewLogger("sqlite")}

	db.AutoMigrate(&Record{})

	return nil
}

type Transaction interface {
	Start() error
	Update(update *Record) error
	Stop() error
}

type storer struct {
	Loadpoint int
	ref       interface{}
}

var _ Transaction = (*storer)(nil)

func NewTransactor(loadpoint int) Transaction {
	return &storer{
		Loadpoint: loadpoint,
	}
}

func (s *storer) Start() error {
	s.ref = &Record{
		StartTime: time.Now(),
		Loadpoint: s.Loadpoint,
	}

	tx := db.Create(s.ref)
	return tx.Error
}

func (s *storer) Update(update *Record) error {
	tx := db.Model(s.ref).Updates(update) // non-zero fields
	return tx.Error
}

func (s *storer) Stop() error {
	rec := &Record{
		EndTime: time.Now(),
	}

	tx := db.Model(s.ref).Updates(rec) // non-zero fields
	return tx.Error
}
