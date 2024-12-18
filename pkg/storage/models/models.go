package models

type Comment struct {
	ID         int    `db:"id" json:"id"`
	News_id    int    `db:"news_id" json:"newsID"`
	Message    string `db:"comment" json:"comment"`
	Created_at int64  `db:"created_at" json:"createdAt"`
	Parrent_id int    `db:"parrent_id" json:"parentID"`
	Censore    bool   `db:"censore" json:"isCensored"`
}

// Модель идентичная модели сервиса gonews.Нужна для проверки начличия id новости
// в рамках работы функции NewsIdCheck
type NewsFullDetailed struct {
	ID        int    `db:"id"`
	Title     string `db:"title"`
	Content   string `db:"description"`
	Preview   string `db:"preview"`
	Published int64  `db:"published"`
	Link      string `db:"link"`
}
