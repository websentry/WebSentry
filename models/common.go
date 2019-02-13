package models

import (
	"github.com/mongodb/mongo-go-driver/mongo"
	"reflect"
)

var mongoDB *mongo.Database

func Init(db *mongo.Database) error {
	mongoDB = db
	return nil
}

// Helper for MongoDB

// modify from mgo "func (iter *Iter) All(result interface{}) error"
func getAllFromCursor(cur *mongo.Cursor, result interface{}) error {
	resultv := reflect.ValueOf(result)
	if resultv.Kind() != reflect.Ptr || resultv.Elem().Kind() != reflect.Slice {
		panic("result argument must be a slice address")
	}
	slicev := resultv.Elem()
	slicev = slicev.Slice(0, slicev.Cap())
	elemt := slicev.Type().Elem()
	i := 0
	for cur.Next(nil) {
		if slicev.Len() == i {
			elemp := reflect.New(elemt)
			err := cur.Decode(elemp.Interface())
			if err != nil { return err }

			slicev = reflect.Append(slicev, elemp.Elem())
			slicev = slicev.Slice(0, slicev.Cap())
		} else {
			err := cur.Decode(slicev.Index(i).Addr().Interface())
			if err != nil { return err }
		}
		i++
	}
	resultv.Elem().Set(slicev.Slice(0, i))

	return cur.Close(nil)
}


//func (iter *Iter) All(result interface{}) error {
//	resultv := reflect.ValueOf(result)
//	if resultv.Kind() != reflect.Ptr || resultv.Elem().Kind() != reflect.Slice {
//		panic("result argument must be a slice address")
//	}
//	slicev := resultv.Elem()
//	slicev = slicev.Slice(0, slicev.Cap())
//	elemt := slicev.Type().Elem()
//	i := 0
//	for {
//		if slicev.Len() == i {
//			elemp := reflect.New(elemt)
//			if !iter.Next(elemp.Interface()) {
//				break
//			}
//			slicev = reflect.Append(slicev, elemp.Elem())
//			slicev = slicev.Slice(0, slicev.Cap())
//		} else {
//			if !iter.Next(slicev.Index(i).Addr().Interface()) {
//				break
//			}
//		}
//		i++
//	}
//	resultv.Elem().Set(slicev.Slice(0, i))
//	return iter.Close()
//}