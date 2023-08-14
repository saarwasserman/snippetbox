package models

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type user struct {
	ID             int
	Name           string
	Email          string
	HashedPassword []byte
	Created        time.Time
}

type UserModel struct {
	DB *sql.DB
}

func (m *UserModel) Insert(name, email, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	stmt := `INSERT INTO users (name, email, hashed_password, created) VALUES ($1, $2, $3, NOW())`
	_, err = m.DB.Exec(stmt, name, email, string(hashedPassword))
	if err != nil {
		var postgresError *pq.Error
		fmt.Println("got an error while inserting")
		if errors.As(err, &postgresError) {
			fmt.Printf("%v\n", postgresError)
			fmt.Printf("%v\n", postgresError.Code)
			if postgresError.Code == "23505" && strings.Contains(postgresError.Message, "users_uc_email") {
				return ErrDuplicateEmail
			}
		}
	}

	return nil
}

func (m *UserModel) Authenticate(email, password string) (int, error) {
	var id int
	var hashed_password []byte

	stmt := `SELECT id, hashed_password FROM users WHERE email = $1`
	err := m.DB.QueryRow(stmt, email).Scan(&id, &hashed_password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	fmt.Println(email, password)

	err = bcrypt.CompareHashAndPassword(hashed_password, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	return id, nil
}

func (m *UserModel) Exists(id int) (bool, error) {
	var exists bool

	stmt := "SELECT EXISTS(SELECT true FROM users WHERE id = $1)"

	err := m.DB.QueryRow(stmt, id).Scan(&exists)
	return exists, err
}
