package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"Skillfactory/Comments/pkg/storage/models"
	"Skillfactory/Comments/pkg/storage/postgress"

	"github.com/gorilla/mux"
)

// Объект API
type Api struct {
	newsDB     *postgress.Storage
	commentsDB *postgress.Storage
	r          *mux.Router
}

// Конструктор API
func New(newsDB, commentsDB *postgress.Storage) *Api {
	api := Api{newsDB: newsDB, commentsDB: commentsDB, r: mux.NewRouter()}
	api.endpoints()
	return &api
}

func (api *Api) Router() *mux.Router {
	return api.r
}

// Регистрация маршрутов
func (api *Api) endpoints() {
	//Маршрут предоставления списка комментариев по ID новости
	api.r.HandleFunc("/comments/{id}", api.GetCommentsHandler).Methods(http.MethodGet, http.MethodOptions)
	//Маршрут добавления комментария
	api.r.HandleFunc("/addcomment/", api.AddCommentHandler).Methods(http.MethodPost, http.MethodOptions)
}

// Хэндлер предоставления списка комментариев
func (api *Api) GetCommentsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == http.MethodOptions {
		return
	}
	//id=news_id
	s := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(s)

	comments, err := api.commentsDB.GetComments(id, api.newsDB)
	if err != nil {
		http.Error(w, "failed get comments from DB", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(comments)
	w.WriteHeader(http.StatusOK)
}

// Хэндлер добавления комментария
func (api *Api) AddCommentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", http.MethodPost)
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Methods", http.MethodPost)
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
		return
	}
	//создание объекта комментария из тела запроса
	var comment models.Comment
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body before adding comment to DB", http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(body, &comment)
	if err != nil {
		http.Error(w, "failed to decode JSON payload", http.StatusBadRequest)
		return
	}
	err = api.commentsDB.AddComment(comment, api.newsDB)
	if err != nil {
		log.Printf("Can't add comment to database: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("the comment has been added successfully!"))
}
