package model

/* Copyright (C) Nikita Evdokimov - All Rights Reserved
 * Unauthorized copying of this file, via any medium is strictly prohibited
 * Proprietary and confidential
 * Written by Nikita Evokimov <nevdokimovm@gmail.com>, 2017
 */
import (
	"gitlab.com/evdokimovn/TaskManagerBot/bot/fsm"

	botAPI "github.com/go-telegram-bot-api/telegram-bot-api"
)

type UserFrom botAPI.User

// User represents user entity
type User struct {
	UserID      int       `json:"UserID"`
	FirstName   string    `json:"FirstName"`
	LastName    string    `json:"LastName"` // may be empty
	UserName    string    `json:"UserName"` // may be empty
	State       fsm.State `json:"State"`
	Language    string    `json:"Language"`
	Initialised bool      `json:"Initialised"`
	Projects    []P       `json:"Projects"`
	Tutorial    T         `json:"Tutorial"`
}

//OwnsProject returns true if user administrator of one of the projects
func (u User) OwnsProject() bool {
	for _, p := range u.Projects {
		if p.Admin {
			return true
		}
	}
	return false
}

//T shows which tutorial messages user has seen
type T struct {
	// Weather user has created first project and seen tutorial
	// about how projects work
	SeenGeneralTutorial bool `json:"seengeneraltutorial"`
	// Weather user has set up notification timer
	SeenTimerTutorial bool `json:"seentimertutorial"`
	//
	SeenInviteTutorial bool `json:"seeninvitetutorial"`
	//
	SeenChatsTutorial bool `json:"seenchatstutorial"`
}

// P is user's project
type P struct {
	Admin      bool
	ProjectUID string
	Name       string
}
