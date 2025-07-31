package storage

import "errors"

type Blog struct {
	Id              int     `db:"id" json:"id"`
	BlogTitle       string  `db:"blog_title" json:"blog_title"`
	BlogDescription *string `db:"blog_description" json:"blog_description"`
	BlogContent     string  `db:"blog_content" json:"blog_content"`
	BlogThumbnail   *string `db:"blog_thumbnail" json:"blog_thumbnail"`
	BlogAuthorId    int     `db:"blog_author_id" json:"blog_author_id"`
	BlogCreatedAt   string  `db:"blog_created_at" json:"blog_created_at"`
	BlogUpdatedAt   *string `db:"blog_updated_at" json:"blog_updated_at"`
}

type BlogTopic struct {
	BlogId  int `db:"blog_id" json:"blog_id"`
	TopicId int `db:"topic_id" json:"topic_id"`
}

type BlogWithTopics struct {
	Blog
	BlogTopics []Topic `json:"blog_topics"`
}

func (s *Storage) CreateBlog(blogTitle string, blogDescription string, blogContent string, blogThumbnail string, blogAuthorId int, blogTopicIds []int) (newBlog *BlogWithTopics, err error) {

	var blogWithTopics BlogWithTopics
	var blog Blog
	var blogTopics []BlogTopic
	var topics []Topic

	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	createBlogQuery := `INSERT INTO blogs(blog_title,blog_description,blog_content,blog_thumbnail,blog_author_id) VALUES($1,$2,$3,$4,$5) RETURNING
	id,blog_title,blog_description,blog_content,blog_thumbnail,blog_author_id,blog_created_at,blog_updated_at`

	row := tx.QueryRowx(createBlogQuery, blogTitle, blogDescription, blogContent, blogThumbnail, blogAuthorId)

	if err := row.StructScan(&blog); err != nil {
		return nil, err
	}

	// create blog topic entries
	createBlogTopicQuery := `INSERT INTO blog_topics(blog_id,topic_id) VALUES($1,$2) RETURNING 
	blog_id,topic_id`

	for _, topicId := range blogTopicIds {

		var blogTopic BlogTopic

		row := tx.QueryRowx(createBlogTopicQuery, blog.Id, topicId)

		if err := row.StructScan(&blogTopic); err != nil {
			return nil, err
		}

		topic, err := s.GetTopicById(topicId)
		if err != nil {
			return nil, err
		}

		blogTopics = append(blogTopics, blogTopic)
		topics = append(topics, *topic)
	}

	blogWithTopics.Blog = blog
	blogWithTopics.BlogTopics = topics

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &blogWithTopics, nil
}

func (s *Storage) GetBlogById(blogId int) (*Blog, error) {

	var blog Blog

	query := `SELECT id,blog_title,blog_description,blog_content,blog_thumbnail,blog_author_id,blog_created_at,blog_updated_at
	FROM blogs WHERE id=$1`

	row := s.db.QueryRowx(query, blogId)

	if err := row.StructScan(&blog); err != nil {
		return nil, err
	}

	return &blog, nil
}

func (s *Storage) DeleteBlogById(blogId int) error {

	query := `DELETE FROM blogs WHERE id=$1`

	result, err := s.db.Exec(query, blogId)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected != 1 {
		return errors.New("failed to delete blog")
	}

	return nil
}
