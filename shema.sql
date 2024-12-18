DROP TABLE IF EXISTS comments;
CREATE TABLE comments (
  id BIGSERIAL PRIMARY KEY,
  news_id INTEGER,
  comment TEXT ,
  created_at BIGINT,
  parrent_id INTEGER , --parrent comment id
);