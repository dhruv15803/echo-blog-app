package storage

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
