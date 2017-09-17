package main

import (
	"errors"
	"fmt"
	"log"
	"time"
)

var (
	errUserExist    = errors.New("user exist")
	errUserNotFound = errors.New("user not found")
)

var (
	users  = make([]*User, 0)
	nextId = 0
)

type User struct {
	Id   int
	Name string
}

func (u *User) Login() error {
	for _, user := range users {
		if u.Name == user.Name {
			return errUserExist
		}
	}
	u.Id = nextId
	nextId++
	users = append(users, u)
	return nil
}

func (u *User) Logout() error {
	for i, user := range users {
		if u.Id == user.Id {
			users = append(users[:i], users[i+1:]...)
		}
	}
	return errUserNotFound
}

func watchUsers() {
	for {
		str := "["
		for _, user := range users {
			str += fmt.Sprintf("%v", user)
		}
		str += "]"
		log.Println("users", str)
		time.Sleep(5 * time.Second)
	}
}
