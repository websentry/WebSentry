package models

import (
	"github.com/bwmarrin/snowflake"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// TODO: transaction support
// TODO: check [RowsAffected] ?

var db *gorm.DB
var snowflakeNode *snowflake.Node

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
