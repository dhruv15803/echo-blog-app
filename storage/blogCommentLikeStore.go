package storage

import "errors"

type BlogCommentLike struct {
	LikedById          int    `db:"liked_by_id" json:"liked_by_id"`
	LikedBlogCommentId int    `db:"liked_blog_comment_id" json:"liked_blog_comment_id"`
	LikedAt            string `db:"liked_at" json:"liked_at"`
}

func (s *Storage) GetBlogCommentLike(likedById int, likedBlogCommentId int) (*BlogCommentLike, error) {

	var blogCommentLike BlogCommentLike

	query := `SELECT liked_by_id, liked_blog_comment_id,liked_at 
FROM blog_comment_likes WHERE liked_by_id=$1 AND liked_blog_comment_id=$2`

	row := s.db.QueryRowx(query, likedById, likedBlogCommentId)
	if err := row.StructScan(&blogCommentLike); err != nil {
		return nil, err
	}

	return &blogCommentLike, nil
}

func (s *Storage) CreateBlogCommentLike(likedById int, likedBlogCommentId int) (*BlogCommentLike, error) {

	var blogCommentLike BlogCommentLike

	query := `INSERT INTO blog_comment_likes(liked_by_id,liked_blog_comment_id) VALUES($1,$2) 
RETURNING liked_by_id, liked_blog_comment_id,liked_at`

	row := s.db.QueryRowx(query, likedById, likedBlogCommentId)

	if err := row.StructScan(&blogCommentLike); err != nil {
		return nil, err
	}

	return &blogCommentLike, nil
}

func (s *Storage) RemoveBlogCommentLike(likedById int, likedBlogCommentId int) error {

	query := `DELETE FROM blog_comment_likes WHERE liked_by_id=$1 AND liked_blog_comment_id=$2`

	result, err := s.db.Exec(query, likedById, likedBlogCommentId)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("failed to remove blog comment")
	}

	return nil
}
