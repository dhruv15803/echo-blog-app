package storage

import "errors"

type BlogLike struct {
	LikedById   int    `db:"liked_by_id" json:"liked_by_id"`
	LikedBlogId int    `db:"liked_blog_id" json:"liked_blog_id"`
	LikedAt     string `db:"liked_at" json:"liked_at"`
}

func (s *Storage) GetBlogLikeByUser(userId int, blogId int) (*BlogLike, error) {

	var blogLike BlogLike

	query := `SELECT liked_by_id,liked_blog_id,liked_at FROM blog_likes 
	WHERE liked_by_id=$1 AND liked_blog_id=$2`

	row := s.db.QueryRowx(query, userId, blogId)

	if err := row.StructScan(&blogLike); err != nil {
		return nil, err
	}

	return &blogLike, nil
}

func (s *Storage) CreateBlogLike(likedById int, likedBlogId int) (*BlogLike, error) {

	var blogLike BlogLike

	query := `INSERT INTO blog_likes(liked_by_id,liked_blog_id) VALUES($1,$2) 
	RETURNING liked_by_id,liked_blog_id,liked_at`

	row := s.db.QueryRowx(query, likedById, likedBlogId)

	if err := row.StructScan(&blogLike); err != nil {
		return nil, err
	}

	return &blogLike, nil
}

func (s *Storage) RemoveLike(likedById int, likedBlogId int) error {

	query := `DELETE FROM blog_likes WHERE liked_by_id=$1 AND liked_blog_id=$2`

	result, err := s.db.Exec(query, likedById, likedBlogId)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected != 1 {
		return errors.New("failed to delete blog like")
	}

	return nil
}
