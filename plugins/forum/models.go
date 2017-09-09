package forum

import (
	"time"

	"github.com/kapmahc/og/plugins/nut"
)

// Article article
type Article struct {
	ID        uint      `gorm:"primary_key" json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Type      string    `json:"type"`
	UpdatedAt time.Time `json:"updatedAt"`
	CreatedAt time.Time `json:"createdAt"`

	UserID   uint      `json:"userId"`
	User     nut.User  `json:"user"`
	Tags     []Tag     `json:"tags" gorm:"many2many:forum_articles_tags;"`
	Comments []Comment `json:"comments"`
}

// TableName table name
func (Article) TableName() string {
	return "forum_articles"
}

// Tag tag
type Tag struct {
	ID        uint      `gorm:"primary_key" json:"id"`
	Name      string    `json:"name"`
	UpdatedAt time.Time `json:"updatedAt"`
	CreatedAt time.Time `json:"createdAt"`

	Articles []Article `json:"articles" gorm:"many2many:forum_articles_tags;"`
}

// TableName table name
func (Tag) TableName() string {
	return "forum_tags"
}

// Comment comment
type Comment struct {
	ID        uint      `gorm:"primary_key" json:"id"`
	Body      string    `json:"body"`
	Type      string    `json:"type"`
	UpdatedAt time.Time `json:"updatedAt"`
	CreatedAt time.Time `json:"createdAt"`

	UserID    uint     `json:"userId"`
	User      nut.User `json:"user"`
	ArticleID uint     `json:"articleId"`
	Article   Article  `json:"article"`
}

// TableName table name
func (Comment) TableName() string {
	return "forum_comments"
}
