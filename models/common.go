package models

import (
	"github.com/bwmarrin/snowflake"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// TODO: check [RowsAffected] ?

var gDB *gorm.DB
var snowflakeNode *snowflake.Node

type TX struct {
	tx *gorm.DB
}

func Init(db *gorm.DB) (err error) {
	gDB = db
	snowflakeNode, err = snowflake.NewNode(1)
	if err != nil {
		return
	}
	return migrate()
}

func IsErrNoDocument(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func Transaction(fun func(tx TX) error) error {
	return gDB.Transaction(func(tx *gorm.DB) error {
		return fun(TX{tx: tx})
	})
}
