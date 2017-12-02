package main

import (
	"errors"
	r "gopkg.in/gorethink/gorethink.v3"
	"os"
	"time"
)

func InitConnectionPool() {
	var err error

	dbUrl := os.Getenv("DB")
	if dbUrl == "" {
		log.Fatal("db env variable not specified")
	}

	session, err = r.Connect(r.ConnectOpts{
		Address:    dbUrl,
		InitialCap: 10,
		MaxOpen:    10,
		Database:   "ArenaBot",
	})
	if err != nil {
		log.Fatal(err)
	}
}

func InsertUser(user User) (r.WriteResponse, error) {
	res, err := r.Table("users").Insert(user, r.InsertOpts{
		Conflict: "update",
	}).RunWrite(session)
	if err != nil {
		return r.WriteResponse{}, err
	}

	return res, nil
}

func InsertFight(id1 int, id2 int) (r.WriteResponse, error) {
	res, err := r.Table("fights").Insert(
		map[string]interface{}{
			"usersId": []int{id1, id2},
			"time":    time.Now(),
		},
	).RunWrite(session)
	if err != nil {
		return r.WriteResponse{}, err
	}

	return res, nil
}

func GetUser(id int) (User, error) {
	res, err := r.Table("users").Get(id).Run(session)
	if err != nil {
		return User{}, err
	}

	var user User
	err = res.One(&user)
	if err == r.ErrEmptyResult {
		return User{}, errors.New("db: row not found")
	}
	if err != nil {
		return User{}, err
	}

	defer res.Close()
	return user, nil
}

func GetFight(id string) (Fight, error) {
	res, err := r.Table("fights").Get(id).Run(session)
	if err != nil {
		return Fight{}, err
	}

	var fight Fight
	err = res.One(&fight)
	if err == r.ErrEmptyResult {
		return Fight{}, errors.New("db: row not found")
	}
	if err != nil {
		return Fight{}, err
	}

	defer res.Close()
	return fight, nil
}

func UpdateUser(user User) (r.WriteResponse, error) {
	res, err := r.Table("users").Get(user.ID).Update(user).RunWrite(session)
	if err != nil {
		return r.WriteResponse{}, err
	}

	return res, nil
}

func UpdateFight(fight Fight) (r.WriteResponse, error) {
	res, err := r.Table("fights").Get(fight.ID).Update(fight).RunWrite(session)
	if err != nil {
		return r.WriteResponse{}, err
	}

	return res, nil
}
