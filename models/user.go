package models

import (
    "time"

    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
)

// UserValidation : Entry in the Validation table
type UserValidation struct {
    Username string `bson:"username"`
    Password string `bson:"password"`
    TimeCreated time.Time `bson:"time_created"`
    TimeExpired time.Time `bson:"time_expired`
}

// User : Entry in the actual User table 
type User struct {
    Username string `bson:"username"`
    //bcrypt
    Password string `bson:"password"`
    TimeCreated time.Time `bson:"time_created"`

    // TODO: task id?
}