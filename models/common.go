package models

import (
	"github.com/bwmarrin/snowflake"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// TODO: check [RowsAffected] ?

var db *gorm.DB
var snowflakeNode *snowflake.Node

type TX struct {
	tx *gorm.DB
}

func Init(_db *gorm.DB) (err error) {
	db = _db
	snowflakeNode, err = snowflake.NewNode(1)
	if err != nil {
		return
	}
	return migrate(db)
}

func IsErrNoDocument(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func Transaction(fun func(tx TX) error) error {
	return db.Transaction(func(tx *gorm.DB) error {
		return fun(TX{tx: tx})
	})
}
