package main

import (
	"time"
)

var (
	dao *Dao
)

// Dao 数据库转化层
type Dao struct{}

// Note 随想
type Note struct {
	ID        int        `db:"id"`
	Content   string     `db:"content"`
	CreatedAt *time.Time `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

// Sharing 分享
type Sharing struct {
	ID        int        `db:"id"`
	Content   string     `db:"content"`
	URL       string     `db:"url"`
	CreatedAt *time.Time `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

// GetAllSharing 获取所有分享
func (d *Dao) GetAllSharing() []Sharing {
	var sharing []Sharing
	db.Select(&sharing, "SELECT * FROM issue ORDER BY updated_at DESC")

	return sharing
}

// GetSharingWithLimit 获取分享
func (d *Dao) GetSharingWithLimit(limit int) []Sharing {
	var sharing []Sharing
	db.Select(&sharing, "SELECT * FROM issue ORDER BY updated_at DESC LIMIT $1", limit)

	return sharing
}

// GetAllNotes 获取所有随想
func (d *Dao) GetAllNotes() []Note {
	var notes []Note
	db.Select(&notes, "SELECT * FROM note ORDER BY updated_at DESC")

	return notes
}

// GetLatestSharing 获取最后一条分享
func (d *Dao) GetLatestSharing() (Sharing, error) {
	var sharing Sharing
	err := db.Select(&sharing, "SELECT * FROM issue ORDER BY updated_at DESC LIMIT 1")

	return sharing, err
}

// AddSharing 增加一条分享
func (d *Dao) AddSharing(url string) error {
	tx := db.MustBegin()
	now := time.Now()
	tx.MustExec("INSERT INTO issue(url, content, created_at, updated_at) VALUES ($1, '', $2, $3)", url, now, now)
	return tx.Commit()
}

// CommentLatestSharing 评论最近一条分享
func (d *Dao) CommentLatestSharing(comment string) error {
	var sharing Sharing

	tx := db.MustBegin()

	tx.Select(&sharing, "SELECT * FROM issue ORDER BY updated_at DESC LIMIT 1")
	tx.MustExec("UPDATE issue SET content=$1, updated_at=$2 WHERE id = $3", comment, time.Now(), sharing.ID)
	return tx.Commit()
}

// AddNote 增加一条随想
func (d *Dao) AddNote(content string) error {
	now := time.Now()

	tx := db.MustBegin()
	tx.MustExec("INSERT INTO note(content, created_at, updated_at) VALUES($1, $2, $3)", content, now, now)

	return tx.Commit()
}
