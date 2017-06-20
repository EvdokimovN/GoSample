package model

/* Copyright (C) Nikita Evdokimov - All Rights Reserved
 * Unauthorized copying of this file, via any medium is strictly prohibited
 * Proprietary and confidential
 * Written by Nikita Evokimov <nevdokimovm@gmail.com>, 2017
 */
import (
	"testing"

	"gitlab.com/evdokimovn/TaskManagerBot/bot/fsm"

	"fmt"

	"os"

	"gopkg.in/mgo.v2/bson"
)

var s Storage
var urlTest = os.Getenv("DB")

const (
	database = "PROJECT"
)

func init() {
	var err error

	if urlTest == "" {
		urlTest = "mongodb://localhost:27017"
	}

	s, err = InitStorage(urlTest, database)
	if err != nil {
		panic(err)
	}

	s.AddCollection(USERS)
	s.AddCollection(PROJECTS)
}

func TestProjectQuestion(t *testing.T) {
	var params = []struct {
		input string
	}{
		{"Hoho"},
		{""},
		{" 123"},
		{" 23 "},
	}
	for _, p := range params {
		input := p.input
		var project Project
		project.CreateQuestion(input)
		if input != project.GetCleanQuestion() {
			t.Errorf("Expected <%s> but got <%s>", input, project.GetCleanQuestion())
		}
	}

}

func TestUpdate(t *testing.T) {
	var state fsm.State
	var userT User
	state = "TEST"
	s.Update("$set", USERS, bson.M{"userid": 52807762}, bson.M{"state": state}, &userT)
	if userT.State != state {
		t.Log(userT)
		t.Errorf("[WRAPPER] Expected user to have state %s, instead got %s", state, userT.State)
	}

}

func TestAssignUID(t *testing.T) {
	var project Project
	p := project.AssignUID()
	lenMustBe := 16
	fmt.Println(p)
	if len(p) != lenMustBe {
		t.Errorf("Len of %s doesn't equall %d but instead equels %d ", p, lenMustBe, len(p))
	}
}

func TestXOR(t *testing.T) {
	var params = []struct {
		f string
		s string
		e string
	}{
		{"00", "11", "11"},
		{"ff", "44", "bb"},
		{"33f385", "33f385", "000000"},
		{"9675a486b46d4e86", "b67bccb02f24cde1", "200e68369b498367"},
	}

	for _, p := range params {
		r, err := xor(p.f, p.s)
		if r != p.e {
			t.Errorf("%s expected but instead got %s. %v", p.e, r, err)
		}
	}
}

func TestProjectTime(t *testing.T) {
	var params = []struct {
		t                string
		offset           string
		shouldBeInFormat bool
	}{
		{"15:54", "+3", true},
		{"15:52", "-12", true},
		{"15", "", false},
		{"12:24", "-13", false},
		{"12:24", "14", true},
	}

	for _, p := range params {
		pt := NewProjectTime(p.t, p.offset)
		if pt.InFormat() != p.shouldBeInFormat {
			t.Errorf("Expected format %v but got %v", p.shouldBeInFormat, pt.InFormat())
		}
	}
}
