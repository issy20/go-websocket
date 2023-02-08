package repository

import (
	"database/sql"
	"log"

	"github.com/issy20/go-websocket/models"
)

type User struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (user *User) GetId() string {
	return user.Id
}

func (user *User) GetName() string {
	return user.Name
}

type UserRepository struct {
	Db *sql.DB
}

func (ur *UserRepository) AddUser(id string, name string, username string, password string) error {
	stmt, err := ur.Db.Prepare("INSERT INTO users(id, name, username, password) values(?, ?, ?, ?)")
	checkErr(err)
	if _, err := stmt.Exec(id, name, username, password); err != nil {
		return err
	}
	return nil
}

func (ur *UserRepository) RemoveUser(user models.IUser) {
	stmt, err := ur.Db.Prepare("DELETE FROM users WHERE id = ?")
	checkErr(err)
	_, err = stmt.Exec(user.GetId())
	checkErr(err)
}

func (ur *UserRepository) FindUserById(ID string) models.IUser {
	row := ur.Db.QueryRow("SELECT id, name FROM users WHERE id = ? LIMIT 1", ID)
	var user User
	if err := row.Scan(&user.Id, &user.Name); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		panic(err)
	}
	return &user
}

func (ur *UserRepository) GetAllUsers() []models.IUser {
	rows, err := ur.Db.Query("SELECT id, name FROM users")
	if err != nil {
		log.Fatal(err)
	}
	var users []models.IUser
	defer rows.Close()
	for rows.Next() {
		var user User
		rows.Scan(&user.Id, &user.Name)
		users = append(users, &user)
	}
	return users
}

func (ur *UserRepository) FindUserByUsername(username string) *User {
	row := ur.Db.QueryRow("SELECT id, name, username, password FROM users WHERE username = ? LIMIT 1", username)
	var user User
	if err := row.Scan(&user.Id, &user.Name, &user.Username, &user.Password); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		panic(err)
	}

	return &user
}

// insert into users(id, name, username, password) values("user1", "Taro", "Taro","password");
// insert into users(id, name, username, password) values("user2", "Jiro", "Jiro","password");
