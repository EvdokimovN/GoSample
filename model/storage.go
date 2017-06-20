package model

/* Copyright (C) Nikita Evdokimov - All Rights Reserved
 * Unauthorized copying of this file, via any medium is strictly prohibited
 * Proprietary and confidential
 * Written by Nikita Evokimov <nevdokimovm@gmail.com>, 2017
 */
import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Storage represents abstract database
type Storage struct {
	DB          string
	session     *mgo.Session
	collections map[string]*mgo.Collection
}

// InitStorage initialises storage
func InitStorage(URL, DB string) (Storage, error) {
	c := make(map[string]*mgo.Collection)
	s, err := mgo.Dial(URL)
	return Storage{session: s, collections: c, DB: DB}, err
}

// ChangeDB changes DB field
func (s *Storage) ChangeDB(DB string) {
	s.DB = DB
}

// GetProject by its UID
func (s Storage) GetProject(pUID string) (Project, error) {
	var project Project
	err := s.GetCollection(PROJECTS).Find(bson.M{"projectuid": pUID}).One(&project)
	return project, err
}

// GetUser by its Telegram ID
func (s Storage) GetUser(id int64) (User, error) {
	var user User
	err := s.GetCollection(USERS).Find(bson.M{"userid": id}).One(&user)
	return user, err
}

//GetCollection gives access to storage collection
func (s Storage) GetCollection(cn string) *mgo.Collection {
	if s.collections == nil {
		s.collections = make(map[string]*mgo.Collection)
	}
	return s.collections[cn]
}

// Update changes fields specified by S of entities specified by F in collection cn and puts updated entry in dst
func (s Storage) Update(oper, cn string, F, S map[string]interface{}, dst interface{}) error {
	change := mgo.Change{
		Update:    bson.M{oper: S},
		ReturnNew: true,
		Upsert:    true,
	}
	_, err := s.collections[cn].Find(F).Apply(change, dst)
	return err
}

//AddCollection adds collection to current storage
func (s *Storage) AddCollection(cn string) {
	if s.collections == nil {
		s.collections = make(map[string]*mgo.Collection)
	}
	s.collections[cn] = s.session.DB(s.DB).C(cn)
}

// Destroy closes current session
func (s *Storage) Destroy() {
	s.session.Close()
}
