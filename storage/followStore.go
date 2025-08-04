package storage

import "errors"

type Follow struct {
	FollowerId  int    `db:"follower_id" json:"follower_id"`
	FollowingId int    `db:"following_id" json:"following_id"`
	FollowedAt  string `db:"followed_at" json:"followed_at"`
}

func (s *Storage) GetFollow(followerId int, followingId int) (*Follow, error) {

	var follow Follow

	query := `SELECT follower_id,following_id,followed_at FROM follows WHERE follower_id=$1 AND following_id=$2`

	if err := s.db.QueryRowx(query, followerId, followingId).StructScan(&follow); err != nil {
		return nil, err
	}

	return &follow, nil
}

func (s *Storage) CreateFollow(followerId int, followingId int) (*Follow, error) {

	var follow Follow

	query := `INSERT INTO follows(follower_id,following_id) VALUES($1,$2) 
RETURNING follower_id,following_id,followed_at`

	if err := s.db.QueryRowx(query, followerId, followingId).StructScan(&follow); err != nil {
		return nil, err
	}

	return &follow, nil
}

func (s *Storage) RemoveFollow(followerId int, followingId int) error {

	query := `DELETE FROM follows WHERE follower_id=$1 AND following_id=$2`

	result, err := s.db.Exec(query, followerId, followingId)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected != 1 {
		return errors.New("failed to remove follow")
	}

	return nil
}
