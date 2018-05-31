package utils

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"sync"
)

var (
	once     sync.Once
	instance *mgo.Database
)

func GetMongoDB() *mgo.Database {
	once.Do(func() {
		if dialInfo, err := mgo.ParseURL("mongodb://localhost:27017/yuanbenlian"); err != nil {
			fmt.Printf("parse mongodb connection url error :%s \n",err.Error())
		} else {
			if session, err := mgo.DialWithInfo(dialInfo); err != nil {
				fmt.Printf("open mongodb session error :%s \n",err.Error())
			} else {
				instance = session.DB(dialInfo.Database)
			}
		}
	})
	return instance
}

func Insert(insert bson.M, col string) error {
	if insert == nil {
		return errors.New("document is nil")
	}
	coll := getCollection(col)
	return coll.Insert(insert)
}

//insert if there is no-exist
func FindAndInsert(query bson.M, insert interface{}, col string) error {
	if insert == nil {
		return errors.New("document is nil")
	}
	coll := getCollection(col)
	c, err := coll.Find(query).Count()
	if err != nil {
		return err
	}
	if c != 0 {
		fmt.Errorf("record is exists")
		return nil
	}
	return coll.Insert(insert)
}

func DeleteAndInsert(query bson.M, insert interface{}, col string) error {
	if insert == nil {
		return errors.New("document is nil")
	}
	coll := getCollection(col)
	coll.Remove(query)
	return coll.Insert(insert)
}

func FindOne(query bson.M, col string, result *map[string]interface{}) error {
	coll := getCollection(col)

	err := coll.Find(query).One(&result)
	return err
}

func FindOneBySelect(query bson.M, sel bson.M, col string, result *map[string]interface{}) error {
	coll := getCollection(col)

	err := coll.Find(query).Select(sel).One(&result)
	return err
}

func Exist(query bson.M, col string) bool {
	coll := getCollection(col)
	c, err := coll.Find(query).Count()
	if err != nil {
		return false
	} else {
		return c != 0
	}

}

func UpdateOne(query bson.M, update bson.M, col string) error {
	coll := getCollection(col)
	return coll.Update(query, update)
}

func DeleteOne(query bson.M, col string) error {
	coll := getCollection(col)
	coll.EnsureIndexKey()
	return coll.Remove(query)
}

func Count(col string) int {
	coll := getCollection(col)
	r, _ := coll.Count()
	return r
}

func getCollection(col string) *mgo.Collection {
	if instance == nil {
		GetMongoDB()
	}
	coll := instance.C(col)
	return coll
}
