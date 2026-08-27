package main

import (
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo *UserRepository
	jwtSecret string
}

func NewAuthService(repo *UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		repo: repo,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthService) Register(email string, password string) error {
	passwordHash,err :=bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return err
	}

	user := User{
		Email: email,
		PasswordHash: string(passwordHash),
	}

	return s.repo.createUser(user)

}

func (s *AuthService) Login (email string, password string) (string,error) {
	user, err := s.repo.getUserByEmail(email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows){
			return "",errors.New("invalid credentials")
		}
		
		return "", err
	}
	
	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password),
	); err != nil {
		return "",errors.New("invalid credentials")
	}

	return createAccessToken(user.ID, s.jwtSecret)
}

func (s *AuthService) UpdateUser(id int, email string,password string, oldPassword string) error {
	user, err := s.repo.getUser(id)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(oldPassword),
	); err != nil {
		return errors.New("invalid credentials")
	}

	updated := User{
		Email:       email,
		PasswordHash: user.PasswordHash,
	}

	if password != "" {
		passwordHash, err := bcrypt.GenerateFromPassword(
			[]byte(password),
			bcrypt.DefaultCost,
		)
		if err != nil {
			return err
		}

		updated.PasswordHash = string(passwordHash)
	}

	if email == "" {
		updated.Email = user.Email
	}

	err = s.repo.updateUser(
		id,
		updated,
	)
	if err != nil {
		return err
	}

	return nil	
}
