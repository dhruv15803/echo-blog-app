package storage

import (
	"errors"
	"time"
)

type Topic struct {
	Id             int     `db:"id" json:"id"`
	TopicTitle     string  `db:"topic_title" json:"topic_title"`
	TopicCreatedAt string  `db:"topic_created_at" json:"topic_created_at"`
	TopicUpdatedAt *string `db:"topic_updated_at" json:"topic_updated_at"`
}

func (s *Storage) CreateTopic(topicTitle string) (*Topic, error) {

	var topic Topic

	createTopicQuery := `INSERT INTO topics(topic_title) VALUES($1) RETURNING id,topic_title,topic_created_at,topic_updated_at`

	row := s.db.QueryRowx(createTopicQuery, topicTitle)

	if err := row.StructScan(&topic); err != nil {
		return nil, err
	}

	return &topic, nil
}

func (s *Storage) GetTopicByTopicTitle(topicTitle string) ([]Topic, error) {
	var topics []Topic

	query := `SELECT id,topic_title,topic_created_at,topic_updated_at FROM 
	topics WHERE topic_title=$1`

	rows, err := s.db.Queryx(query, topicTitle)
	if err != nil {
		return []Topic{}, err
	}

	defer rows.Close()

	for rows.Next() {

		var topic Topic

		if err := rows.StructScan(&topic); err != nil {
			return []Topic{}, err
		}

		topics = append(topics, topic)
	}

	return topics, nil
}

func (s *Storage) GetTopicById(id int) (*Topic, error) {
	var topic Topic

	query := `SELECT id,topic_title,topic_created_at,topic_updated_at
	FROM topics WHERE id=$1`

	row := s.db.QueryRowx(query, id)

	if err := row.StructScan(&topic); err != nil {
		return nil, err
	}

	return &topic, nil
}

func (s *Storage) DeleteTopicById(id int) error {

	query := `DELETE FROM topics WHERE id=$1`

	result, err := s.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected != 1 {
		return errors.New("failed to delete topic")
	}

	return nil
}

func (s *Storage) UpdateTopicById(id int, topicTitle string) (*Topic, error) {
	var newTopic Topic

	query := `UPDATE topics
	SET topic_title=$1,topic_updated_at=$2 
	WHERE id=$3 RETURNING id,topic_title,topic_created_at,topic_updated_at`

	row := s.db.QueryRowx(query, topicTitle, time.Now(), id)

	if err := row.StructScan(&newTopic); err != nil {
		return nil, err
	}

	return &newTopic, nil
}

func (s *Storage) GetTopics(skip int, limit int) ([]Topic, error) {
	var topics []Topic

	query := `SELECT id,topic_title,topic_created_at,topic_updated_at 
	FROM topics 
	ORDER BY topic_created_at DESC
	LIMIT $1 OFFSET $2`

	rows, err := s.db.Queryx(query, limit, skip)
	if err != nil {
		return []Topic{}, err
	}

	defer rows.Close()

	for rows.Next() {

		var topic Topic

		if err := rows.StructScan(&topic); err != nil {
			return []Topic{}, err
		}

		topics = append(topics, topic)
	}

	return topics, nil
}

func (s *Storage) GetTopicsCount() (int, error) {

	var totalTopicsCount int

	query := `SELECT COUNT(*) FROM topics`

	row := s.db.QueryRow(query)

	if err := row.Scan(&totalTopicsCount); err != nil {
		return -1, err
	}

	return totalTopicsCount, nil
}

func (s *Storage) GetTopicsBySearchTitleText(searchTitleText string, skip int, limit int) ([]Topic, error) {
	var topics []Topic

	query := `SELECT id,topic_title,topic_created_at,topic_updated_at
	FROM topics WHERE topic_title ILIKE $1
	ORDER BY topic_created_at DESC
	LIMIT $2 OFFSET $3`

	topicTitleSearchParam := "%" + searchTitleText + "%"

	rows, err := s.db.Queryx(query, topicTitleSearchParam, limit, skip)
	if err != nil {
		return []Topic{}, err
	}

	defer rows.Close()

	for rows.Next() {
		var topic Topic

		if err := rows.StructScan(&topic); err != nil {
			return []Topic{}, err
		}

		topics = append(topics, topic)
	}

	return topics, nil
}

func (s *Storage) GetTopicsCountBySearchTitleText(searchTitleText string) (int, error) {

	var totalTopicsCount int

	query := `SELECT COUNT(*) FROM topics WHERE topic_title ILIKE $1`

	topicTitleSearchParam := "%" + searchTitleText + "%"

	row := s.db.QueryRow(query, topicTitleSearchParam)

	if err := row.Scan(&totalTopicsCount); err != nil {
		return -1, err
	}

	return totalTopicsCount, nil
}
