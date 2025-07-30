package storage

import (
	"errors"
	"time"
)

type userRole string

const (
	AdminRole userRole = "admin"
	UserRole  userRole = "user"
)

type User struct {
	Id         int      `db:"id" json:"id"`
	Email      string   `db:"email" json:"email"`
	Password   string   `db:"password" json:"-"`
	Name       *string  `db:"name" json:"name"`
	IsVerified bool     `db:"is_verified" json:"is_verified"`
	ImageUrl   *string  `db:"image_url" json:"image_url"`
	Role       userRole `db:"role" json:"role"`
	CreatedAt  string   `db:"created_at" json:"created_at"`
	UpdatedAt  *string  `db:"updated_at" json:"updated_at"`
}

type UserInvitation struct {
	Token      string `db:"token" json:"token"`
	UserId     int    `db:"user_id" json:"user_id"`
	Expiration string `db:"expiration" json:"expiration"`
}

type PasswordReset struct {
	Token        string `db:"token" json:"token"`
	UserId       int    `db:"user_id" json:"user_id"`
	ExpirationAt string `db:"expiration_at" json:"expiration_at"`
}

func (s *Storage) GetUserByEmail(email string) (*User, error) {

	var user User

	query := `SELECT id,email,password,name,is_verified,image_url,role,created_at,updated_at 
	FROM users WHERE email=$1`

	if err := s.db.Get(&user, query, email); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *Storage) GetUserById(id int) (*User, error) {

	var user User

	query := `SELECT id,email,password,name,is_verified,image_url,role,created_at,updated_at 
	FROM users WHERE id=$1`

	if err := s.db.Get(&user, query, id); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *Storage) GetVerifiedUserByEmail(email string) (*User, error) {

	var user User

	query := `SELECT id,email,password,name,is_verified,image_url,role,created_at,updated_at 
	FROM users WHERE email=$1 AND is_verified=true`

	if err := s.db.Get(&user, query, email); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *Storage) CreateUserAndInvitation(email string, password string, token string, expiration time.Time) (user User, err error) {

	tx, err := s.db.Beginx()
	if err != nil {
		return User{}, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	createUserQuery := `INSERT INTO users(email,password) VALUES($1,$2) RETURNING 
	id,email,password,name,is_verified,image_url,role,created_at,updated_at`

	row := tx.QueryRowx(createUserQuery, email, password)
	if err := row.StructScan(&user); err != nil {
		return User{}, err
	}

	createInvitationQuery := `INSERT INTO user_invitations(token,user_id,expiration) VALUES($1,$2,$3)`

	result, err := tx.Exec(createInvitationQuery, token, user.Id, expiration)
	if err != nil {
		return User{}, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return User{}, err
	}

	if rowsAffected != 1 {
		return User{}, errors.New("failed to insert user invitation")
	}

	if err := tx.Commit(); err != nil {
		return User{}, err
	}

	return user, nil
}

func (s *Storage) ActivateUserHandler(token string) (*User, error) {

	var activeUser User
	var err error

	var userInvitation UserInvitation

	query := `SELECT token,user_id,expiration FROM user_invitations WHERE token=$1 AND expiration > $2`

	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	row := tx.QueryRowx(query, token, time.Now())

	if err := row.StructScan(&userInvitation); err != nil {
		return nil, err
	}

	userId := userInvitation.UserId
	user, err := s.GetUserById(userId)
	if err != nil {
		return nil, err
	}

	// upate is_verified field of this user id and clean up other tries of this
	verifyUserQuery := `UPDATE users SET is_verified=true WHERE id=$1 
	RETURNING id,email,password,name,is_verified,image_url,role,created_at,updated_at`

	activatedUserRow := tx.QueryRowx(verifyUserQuery, user.Id)
	if err := activatedUserRow.StructScan(&activeUser); err != nil {
		return nil, err
	}

	// clean up the invitations after user is verified
	deleteUserInvitationsQuery := `DELETE FROM user_invitations WHERE user_id=$1`

	_, err = tx.Exec(deleteUserInvitationsQuery, activeUser.Id)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &activeUser, nil
}

func (s *Storage) CreateAdminUser(email string, password string) (*User, error) {
	var adminUser User

	createAdminUserQuery := `INSERT INTO users(email,password,is_verified,role) VALUES($1,$2,$3,$4) 
	RETURNING  id,email,password,name,is_verified,image_url,role,created_at,updated_at`

	row := s.db.QueryRowx(createAdminUserQuery, email, password, true, AdminRole)

	if err := row.StructScan(&adminUser); err != nil {
		return nil, err
	}

	return &adminUser, nil
}

func (s *Storage) CreatePasswordReset(token string, userId int, expiration time.Time) (*PasswordReset, error) {

	var passwordReset PasswordReset

	query := `INSERT INTO password_resets(token,user_id,expiration_at) VALUES($1,$2,$3) RETURNING token,user_id,expiration_at`

	row := s.db.QueryRowx(query, token, userId, expiration)

	if err := row.StructScan(&passwordReset); err != nil {
		return nil, err
	}

	return &passwordReset, nil
}

func (s *Storage) ResetPassword(newPassword string, token string) (userPtr *User, err error) {

	// search for entry with token
	// get user id from that query
	// update user with new password
	var passwordReset PasswordReset
	var user User

	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	query := `SELECT token,user_id,expiration_at FROM password_resets 
	WHERE token=$1 AND expiration_at > $2`

	row := tx.QueryRowx(query, token, time.Now())

	if err := row.StructScan(&passwordReset); err != nil {
		return nil, err
	}

	userId := passwordReset.UserId

	resetPasswordQuery := `UPDATE users
	SET password=$1 WHERE id=$2 RETURNING 
	id,email,password,name,is_verified,image_url,role,
	created_at,updated_at`

	updatedUserRow := tx.QueryRowx(resetPasswordQuery, newPassword, userId)

	if err := updatedUserRow.StructScan(&user); err != nil {
		return nil, err
	}

	// remove password reset entry as it has been used once now to update password

	deletePasswordResetQuery := `DELETE FROM password_resets WHERE token=$1`

	result, err := tx.Exec(deletePasswordResetQuery, token)
	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rowsAffected != 1 {
		return nil, errors.New("failed to delete password reset")
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &user, nil
}
