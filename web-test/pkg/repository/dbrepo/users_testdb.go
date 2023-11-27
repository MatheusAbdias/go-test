package dbrepo

import (
	"database/sql"
	"errors"
	"time"

	"webapp/pkg/data"
)

type TestDBRepo struct{}

func (m *TestDBRepo) Connection() *sql.DB {
	return nil
}

func (m *TestDBRepo) AllUsers() ([]*data.User, error) {
	var users []*data.User

	return users, nil
}

func (m *TestDBRepo) GetUser(id int) (*data.User, error) {
	user := data.User{ID: 1}
	return &user, nil
}

func (m *TestDBRepo) GetUserByEmail(email string) (*data.User, error) {
	if email == "admin@example.com" {
		user := data.User{
			ID:        1,
			FirstName: "admin",
			LastName:  "user",
			Email:     "admin@example.com",
			Password:  "$2a$14$ajq8Q7fbtFRQvXpdCq7Jcuy.Rx1h/L4J60Otx.gyNLbAYctGMJ9tK",
			IsAdmin:   1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		return &user, nil
	}

	return nil, errors.New("not found")
}

func (m *TestDBRepo) UpdateUser(u data.User) error {
	return nil
}

func (m *TestDBRepo) DeleteUser(id int) error {
	return nil
}

func (m *TestDBRepo) InsertUser(user data.User) (int, error) {
	return 1, nil
}

func (m *TestDBRepo) ResetPassword(id int, password string) error {
	return nil
}

func (m *TestDBRepo) InsertUserImage(i data.UserImage) (int, error) {
	return 1, nil
}
