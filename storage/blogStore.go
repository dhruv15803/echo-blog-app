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

type BlogWithMetaData struct {
	Blog
	BlogAuthor         User    `json:"blog_author"`
	BlogTopics         []Topic `json:"blog_topics"`
	BlogLikesCount     int     `json:"blog_likes_count"`
	BlogCommentsCount  int     `json:"blog_comments_count"`
	BlogBookmarksCount int     `json:"blog_bookmarks_count"`
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

func (s *Storage) GetBlogsByTopic(topicId int, skip int, limit int, likesCountWt, bookmarksCountWt, commentsCountWt float64) ([]BlogWithMetaData, error) {

	var blogs []BlogWithMetaData

	query := `SELECT * , (($4::numeric * likes_count + $5::numeric * bookmarks_count + $6::numeric * comments_count) / ( POWER(EXTRACT (EPOCH FROM (NOW() - blog_created_at)),2))) AS activity_score FROM (
	SELECT 
	b.id,b.blog_title,b.blog_description,b.blog_content,b.blog_thumbnail,b.blog_author_id,
	b.blog_created_at,b.blog_updated_at,u.id,u.email,u.password,u.name,u.is_verified,u.image_url,
	u.role,u.created_at,u.updated_at,
	COUNT(DISTINCT bl.liked_by_id) AS likes_count,
	COUNT(DISTINCT bb.bookmarked_by_id) AS bookmarks_count,
	COUNT(DISTINCT bc.id) AS comments_count
FROM 
	blogs AS b INNER JOIN users AS u ON b.blog_author_id=u.id 
	LEFT JOIN blog_likes AS bl ON bl.liked_blog_id=b.id 
	LEFT JOIN blog_bookmarks AS bb ON bb.bookmarked_blog_id=b.id
	LEFT JOIN blog_comments AS bc ON bc.blog_id = b.id AND bc.parent_comment_id IS NULL
WHERE b.id IN (SELECT blog_id FROM blog_topics WHERE topic_id=$1)
GROUP BY 
	b.id,u.id
)  
ORDER BY activity_score DESC 
LIMIT $2 OFFSET $3`

	rows, err := s.db.Queryx(query, topicId, limit, skip, likesCountWt, bookmarksCountWt, commentsCountWt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {

		var blog BlogWithMetaData
		var activityScore float64

		if err := rows.Scan(&blog.Id, &blog.BlogTitle, &blog.BlogDescription, &blog.BlogContent, &blog.BlogThumbnail, &blog.BlogAuthorId,
			&blog.BlogCreatedAt, &blog.BlogUpdatedAt, &blog.BlogAuthor.Id, &blog.BlogAuthor.Email, &blog.BlogAuthor.Password, &blog.BlogAuthor.Name,
			&blog.BlogAuthor.IsVerified, &blog.BlogAuthor.ImageUrl, &blog.BlogAuthor.Role, &blog.BlogAuthor.CreatedAt,
			&blog.BlogAuthor.UpdatedAt, &blog.BlogLikesCount, &blog.BlogBookmarksCount, &blog.BlogCommentsCount, &activityScore); err != nil {
			return nil, err
		}

		var blogTopics []Topic
		blogTopicsQuery := `SELECT id,topic_title,topic_created_at,topic_updated_at 
FROM topics WHERE id IN (SELECT topic_id FROM blog_topics WHERE blog_id=$1)`

		topicRows, err := s.db.Queryx(blogTopicsQuery, blog.Id)
		if err != nil {
			return nil, err
		}

		for topicRows.Next() {

			var topic Topic

			if err := topicRows.StructScan(&topic); err != nil {
				return nil, err
			}

			blogTopics = append(blogTopics, topic)
		}
		blog.BlogTopics = blogTopics
		blogs = append(blogs, blog)
	}

	return blogs, nil
}

func (s *Storage) GetBlogsCountByTopic(topicId int) (int, error) {

	var totalBlogsCountByTopic int

	query := `SELECT COUNT(blog_id) FROM blog_topics WHERE topic_id=$1`

	if err := s.db.QueryRow(query, topicId).Scan(&totalBlogsCountByTopic); err != nil {
		return -1, err
	}

	return totalBlogsCountByTopic, nil
}
