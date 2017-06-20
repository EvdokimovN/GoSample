package model

/* Copyright (C) Nikita Evdokimov - All Rights Reserved
 * Unauthorized copying of this file, via any medium is strictly prohibited
 * Proprietary and confidential
 * Written by Nikita Evokimov <nevdokimovm@gmail.com>, 2017
 */
import (
	"encoding/hex"
	"fmt"

	"strings"

	guid "github.com/satori/go.uuid"

	"time"

	"bytes"

	"strconv"

	"gopkg.in/mgo.v2/bson"
)

//ProjectTime enforces that Time part of project object
//is in hh:mm offset where offset is reltive to UTC-0 format
type ProjectTime string

const (
	layout = "15:04"
	lenCap = 500
)

func NewProjectTime(t, offset string) ProjectTime {
	var buffer bytes.Buffer
	buffer.WriteString(t)
	buffer.WriteString(" ")
	buffer.WriteString(offset)
	return ProjectTime(buffer.String())
}

//InFormat return true if variable in hh:mm format
func (pt ProjectTime) InFormat() bool {
	s := string(pt)
	parts := strings.Split(s, " ")
	if len(parts) != 2 {
		return false
	}
	t := parts[0]
	offset := parts[1]
	_, err := time.Parse(layout, t)
	off, errt := strconv.Atoi(offset)
	if off < -12 || off > 14 {
		return false
	}
	return err == nil && errt == nil
}

//ToTime returns time.Time representation of ProjectTime in UTC 0 timezone.
//Note that only Hour() and Minute() will have meaningfull values.
func (pt ProjectTime) ToTime() (time.Time, error) {
	if pt.InFormat() {
		split := strings.Split(string(pt), " ")
		t, _ := time.Parse(layout, split[0])
		offset, _ := strconv.Atoi(split[1])
		return t.Add(-time.Duration(offset) * time.Hour), nil
	}
	return time.Now(), fmt.Errorf("not in format")
}

//String representation of ProjecTime. Has format hh:mm where values in UTC 0 timezone.
func (pt ProjectTime) String() string {
	split := strings.Split(string(pt), " ")
	t, _ := time.Parse(layout, split[0])
	offset, _ := strconv.Atoi(split[1])
	ts := t.Add(-time.Duration(offset) * time.Hour)
	return fmt.Sprintf("%02d:%02d", ts.Hour(), ts.Minute())
}

func (pt ProjectTime) PrettyPrint() string {
	if pt.InFormat() {
		split := strings.Split(string(pt), " ")
		off, _ := strconv.Atoi(split[1])
		return fmt.Sprintf("%s UTC %d", split[0], off)
	}
	return ""
}

func (p Project) GetCleanQuestion() string {
	// Magic number. Number of special characters appended to question
	// by CreateQuestion method
	k := 7 + len(p.Name)
	if len(p.Question) > k {
		return p.Question[k:]
	}
	return ""
}

// Create project's notification message from supplied string by prepending projects name
func (p *Project) CreateQuestion(q string) string {
	if len(q) > lenCap {
		q = q[:lenCap]
	}
	var buffer bytes.Buffer
	buffer.WriteString("❔")
	buffer.WriteString("`")
	buffer.WriteString(p.Name)
	buffer.WriteString(":")
	buffer.WriteString("`")
	buffer.WriteString(" ")
	buffer.WriteString(q)
	q = buffer.String()
	p.Question = q
	return q
}

// Project is a project in bot
type Project struct {
	// Project name set by user
	Name string `json:"Name"`
	// Iternal project number assigned by the system
	ProjectUID string `json:"ProjectUID"`
	// Admin's Telegram ID
	Admin int `json:"Admin"`
	// Array of Telegram ID of project members including Admin
	Users []int `json:"Users"`
	// Array of Telegram ID of chats added to the project
	Chats []int64 `json:"Chats"`
	// Time at wich notification must be fired. Has (hh:mm offset) format
	Time string `json:"Time"`
	// Wheather notifications are enabled. If time is empty must be set to false
	Active bool `json:"active"`
	// Time of project creation
	CreationTime time.Time `json:"creationtime"`
	// Question that sent to users with notification
	Question string `json:"question"`
}

//xor two hex string of the same size and return result in hex
func xor(a, b string) (string, error) {
	f, er1 := hex.DecodeString(a)
	s, er2 := hex.DecodeString(b)
	if er1 != nil || er2 != nil {
		return "", er1
	}
	if len(f) != len(s) {
		return "", fmt.Errorf("arguments have different lengths")
	}
	var buffer bytes.Buffer
	l := len(f)
	for i := 0; i < l; i++ {
		buffer.Write([]byte{f[i] ^ s[i]})
	}
	return fmt.Sprintf("%x", buffer.String()), nil
}

// AssignUID to project and return it
func (p *Project) AssignUID() string {
	UID := strings.Replace(guid.NewV4().String(), "-", "", -1)
	l := len(UID)
	//XOR halves of UID to shorten it. XOR preserves UUID quilites
	//though shortage inceseases collision probability
	//
	//UUID is 128 bits represented as hex string. Two of those bytes are always equall
	//which in turn reduces `original` length to 96.
	//In XOR we use whole all bits. New UUID is 64 bits long. According to Wiki
	//https://en.wikipedia.org/wiki/Birthday_problem#Probability_table
	//to get collision probability of 0.1% on average 1.9x10^8 strings must be generated.
	//
	//I realise this number may not apply to our use case, but at the time I cannot provide more
	//adequete analyssis and happy even with lower limit which is 7.5x10^5
	f := UID[:l/2]
	s := UID[l/2:]
	UID, _ = xor(f, s)
	p.ProjectUID = UID
	return UID
}

//CreateProject and save it in datastore
func (s Storage) CreateProject(n string, userID int) (Project, error) {
	p := Project{
		Name:         n,
		Admin:        userID,
		Time:         "",
		Active:       false,
		CreationTime: time.Now(),
	}
	puid := p.AssignUID()
	//TODO: (evdokimovn) get rid of hard coded falues
	cp, ok := s.collections[PROJECTS]
	if !ok {
		return p, fmt.Errorf("there is no collections named [projects]")
	}

	err := cp.Insert(&p)
	if err != nil {
		return p, err
	}
	pu := P{
		Admin:      true,
		ProjectUID: p.ProjectUID,
		Name:       p.Name,
	}

	var userT User
	err = s.Update("$push", USERS, bson.M{"userid": userID}, bson.M{"projects": pu}, &userT)
	if err != nil {
		return p, err
	}
	err = s.Update("$push", PROJECTS, bson.M{"projectuid": puid}, bson.M{"users": userT.UserID}, &p)
	if err != nil {
		return p, err
	}
	return p, nil
}
