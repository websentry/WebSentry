package models

import (
	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
)

// TODO: transaction support

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
