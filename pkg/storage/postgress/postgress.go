package postgress

import (

	//"Skillfactory/Comments/pkg/storage/postgress"
	"Skillfactory/Comments/pkg/storage/models"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	//"strings"
	"time"

	//"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
)

type Storage struct {
	DB *pgxpool.Pool
}

// Функция инициализации подключения к новостной БД
func NewDB(DBname string) (*Storage, error) {
	err := godotenv.Load()
	if err != nil {
		log.Println("cant loading .env file")
		return nil, err
	}
	pwd := os.Getenv("DBPASSWORD")
	connString := "postgres://postgres:" + pwd + "@localhost:5432/" + DBname

	pool, err := pgxpool.Connect(context.Background(), connString)
	if err != nil {
		log.Printf("cant create new instance of DB: %v\n", err)
		return nil, err
	}
	s := Storage{
		DB: pool,
	}
	return &s, nil
}

// Функция проверки наличия новости по ID
func NewsIdCheck(id int, newsDB *Storage) bool {
	var exists bool = true
	row := newsDB.DB.QueryRow(context.Background(), `SELECT EXISTS(SELECT 1 FROM news WHERE id = $1)`, id)

	err := row.Scan(&exists)
	if err != nil {
		log.Printf("cant scan row: %v\n", err)
		return false
	}
	if !exists {
		log.Printf("news with ID %d does not exist.\n", id)
		return false
	}
	return true
}

// Функция проверки наличия комментария по ID
func CommentIdCheck(parrentID int, commentsDB *Storage) bool {
	var exists bool = true
	row := commentsDB.DB.QueryRow(context.Background(), `SELECT EXISTS(SELECT 1 FROM comments WHERE id = $1)`, parrentID)
	err := row.Scan(&exists)
	if parrentID == 0 {
		return true
	}
	if err != nil {
		log.Printf("cant scan row: %v\n", err)
		return false
	}
	if !exists {
		log.Printf("comment with ID %d does not exist.\n", parrentID)
		return false
	}
	return true
}

// Старая версия метода добавления комментария в БД
func (s *Storage) AddCommentDepricated(newsID int, parrentID int, comment string, newsStore *Storage) (err error) {
	if !NewsIdCheck(newsID, newsStore) {
		e := fmt.Errorf("the news with this ID does not exist - %v\n", newsID)
		log.Println(e)
		err = errors.New("the news with this ID does not exist")
		return err
	}
	if !CommentIdCheck(parrentID, s) {
		e := fmt.Errorf("the comment with this ID does not exist - %v\n", parrentID)
		log.Println(e)
		err = errors.New("the comment with this ID does not exist")
		return err
	}
	unixTime := time.Now().Unix()

	_, err = s.DB.Exec(context.Background(), `INSERT INTO comments (news_id,comment,created_at,parrent_id) 
	VALUES($1,$2,$3,$4)`, newsID, comment, unixTime, parrentID)
	if err != nil {
		log.Printf("cant add comment in database! %v\n", err)
		return err
	}
	return nil
}

// Метод добавления комментария в БД
func (s *Storage) AddComment(comment models.Comment, newsStore *Storage) (err error) { //заменить на интерфейс
	if !NewsIdCheck(comment.News_id, newsStore) {
		e := fmt.Errorf("the news with this ID does not exist - %v\n", comment.News_id)
		log.Println(e)
		err = errors.New("the news with this ID does not exist")
		return err
	}
	if !CommentIdCheck(comment.Parrent_id, s) {
		e := fmt.Errorf("the comment with this ID does not exist - %v\n", comment.Parrent_id)
		log.Println(e)
		err = errors.New("the comment with this ID does not exist")
		return err
	}
	unixTime := time.Now().Unix()

	_, err = s.DB.Exec(context.Background(), `INSERT INTO comments (news_id,comment,created_at,parrent_id) 
		VALUES($1,$2,$3,$4)`, comment.News_id, comment.Message, unixTime, comment.Parrent_id)
	if err != nil {
		log.Printf("cant add comment in database! %v\n", err)
		return err
	}
	return nil
}

// Метод получения списка комментариев по ID новости
func (s *Storage) GetComments(newsID int, newsStore *Storage) (comments []models.Comment, err error) {
	if !NewsIdCheck(newsID, newsStore) {
		return []models.Comment{}, fmt.Errorf("the news with this ID does not exist- %v\n", err)
	}
	if newsID < 1 {
		err := fmt.Errorf("bad news ID:  %v", newsID)
		log.Println(err)
		e := errors.New("bad news ID")
		return []models.Comment{}, e
	}
	rows, err := s.DB.Query(context.Background(), `SELECT * FROM comments WHERE news_id=$1
	ORDER BY created_at`, newsID)
	if err != nil {
		log.Printf("cant get comments from database! %v\n", err)
		return []models.Comment{}, err
	}
	defer rows.Close()

	for rows.Next() {
		comment := models.Comment{}
		err = rows.Scan(
			&comment.ID,
			&comment.News_id,
			&comment.Message,
			&comment.Created_at,
			&comment.Parrent_id,
		)
		if err != nil {
			return nil, fmt.Errorf("unable scan row: %w", err)
		}
		comments = append(comments, comment)
	}
	return comments, nil
}
