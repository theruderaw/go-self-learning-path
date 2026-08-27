package main

import (
	"database/sql"
)

type UserRepository struct {
	db *sql.DB
}

func (r *UserRepository)createUser(user User) error {

	query := `
	INSERT INTO users(email, password_hash) 
	VALUES ($1, $2)
	`
	_, err := r.db.Exec(
		query,
		user.Email, user.PasswordHash,
	)

	return err
}


func (r *UserRepository)getUser(id int) (User,error) {

	var user User

	query := `
	SELECT id, email, password_hash, created_at
	FROM users
	WHERE id = $1
	`

	err := r.db.QueryRow(
		query,
		id,
	).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	return user,err
}

func (r *UserRepository) getUserByEmail (email string) (User,error) {

	var user User

	query := `
	SELECT id, email, password_hash,created_at
	FROM users
	WHERE email = $1 
	`

	err := r.db.QueryRow(
		query,
		email,
	).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	return user,err
}

func (r *UserRepository)getUsers() ([]User,error) {

	var users []User

	query := `
	SELECT id, email, password_hash,created_at
	FROM users
	`

	rows, err := r.db.Query(query,)
	if err != nil {
		return nil,err
	}

	defer rows.Close()

	for rows.Next(){
		var user User

		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.PasswordHash,
			&user.CreatedAt,
		)

		if err != nil {
			return nil,err
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil,err
	}

	return users,nil


}


func (r *UserRepository)updateUser(id int, user User) error {

	query := `
	UPDATE users
	SET email = COALESCE(NULLIF($1, ''), email)
		,password_hash = $2
	WHERE id = $3
	`

	_,err := r.db.Exec(
		query,
		user.Email,
		user.PasswordHash,
		id,
	)

	return err
}

func (r *UserRepository)deleteUser(id int) error {
	query := `
	DELETE FROM
	users
	WHERE id = $1
	`

	_, err := r.db.Exec(
		query,
		id,
	)

	return err
}
