package storage

type BlogComment struct {
	Id               int     `db:"id" json:"id"`
	CommentContent   string  `db:"comment_content" json:"comment_content"`
	BlogId           int     `db:"blog_id" json:"blog_id"`
	CommentAuthorId  int     `db:"comment_author_id" json:"comment_author_id"`
	ParentCommentId  *int    `db:"parent_comment_id" json:"parent_comment_id"`
	CommentCreatedAt string  `db:"comment_created_at" json:"comment_created_at"`
	CommentUpdatedAt *string `db:"comment_updated_at" json:"comment_updated_at"`
}

// this creates a top level blog comment (not a nested child comment)
func (s *Storage) CreateBlogComment(commentContent string, blogId int, commentAuthorId int) (*BlogComment, error) {

	var comment BlogComment

	query := `INSERT INTO blog_comments(comment_content,blog_id,comment_author_id) VALUES($1,$2,$3)
	RETURNING id,comment_content,blog_id,comment_author_id,parent_comment_id,comment_created_at,comment_updated_at`

	row := s.db.QueryRowx(query, commentContent, blogId, commentAuthorId)

	if err := row.StructScan(&comment); err != nil {
		return nil, err
	}

	return &comment, nil
}

func (s *Storage) CreateChildBlogComment(commentContent string, blogId int, commentAuthorId int, parentCommentId int) (*BlogComment, error) {

	var childComment BlogComment

	query := `INSERT INTO blog_comments(comment_content,blog_id,comment_author_id,parent_comment_id) VALUES($1,$2,$3,$4) RETURNING 
	id,comment_content,blog_id,comment_author_id,parent_comment_id,comment_created_at,comment_updated_at`

	row := s.db.QueryRowx(query, commentContent, blogId, commentAuthorId, parentCommentId)

	if err := row.StructScan(&childComment); err != nil {
		return nil, err
	}

	return &childComment, nil
}

func (s *Storage) GetBlogCommentById(commentId int) (*BlogComment, error) {

	var blogComment BlogComment

	query := `SELECT id,comment_content,blog_id,comment_author_id,parent_comment_id,comment_created_at,comment_updated_at 
	FROM blog_comments WHERE id=$1`

	row := s.db.QueryRowx(query, commentId)

	if err := row.StructScan(&blogComment); err != nil {
		return nil, err
	}

	return &blogComment, nil
}
