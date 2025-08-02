package storage

import "errors"

type BlogBookmark struct {
	BookmarkedById   int    `db:"bookmarked_by_id" json:"bookmarked_by_id"`
	BookmarkedBlogId int    `db:"bookmarked_blog_id" json:"bookmarked_blog_id"`
	BookmarkedAt     string `db:"bookmarked_at" json:"bookmarked_at"`
}

func (s *Storage) GetBlogBookmark(bookmarkedById int, bookmarkedBlogId int) (*BlogBookmark, error) {

	var blogBookmark BlogBookmark

	query := `SELECT bookmarked_by_id,bookmarked_blog_id,bookmarked_at FROM blog_bookmarks WHERE bookmarked_by_id=$1 AND bookmarked_blog_id=$2`

	row := s.db.QueryRowx(query, bookmarkedById, bookmarkedBlogId)

	if err := row.StructScan(&blogBookmark); err != nil {
		return nil, err
	}

	return &blogBookmark, nil
}

func (s *Storage) CreateBlogBookmark(bookmarkedById int, bookmarkedBlogId int) (*BlogBookmark, error) {

	var blogBookmark BlogBookmark

	query := `INSERT INTO blog_bookmarks(bookmarked_by_id,bookmarked_blog_id) VALUES($1,$2) 
RETURNING bookmarked_by_id,bookmarked_blog_id,bookmarked_at`

	row := s.db.QueryRowx(query, bookmarkedById, bookmarkedBlogId)

	if err := row.StructScan(&blogBookmark); err != nil {
		return nil, err
	}

	return &blogBookmark, nil
}

func (s *Storage) RemoveBlogBookmark(bookmarkedById int, bookmarkedBlogId int) error {

	query := `DELETE FROM blog_bookmarks WHERE bookmarked_by_id=$1 AND bookmarked_blog_id=$2`

	result, err := s.db.Exec(query, bookmarkedById, bookmarkedBlogId)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected != 1 {
		return errors.New("failed to remove blog bookmark")
	}

	return nil
}
