package main

import (
	"log"
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
	if err := db.Select(&sharing, "SELECT * FROM issue ORDER BY updated_at DESC"); err != nil {
		log.Printf("failed to get all issue: %s", err)
		return nil
	}

	return sharing
}

// GetSharing 获取分享
func (d *Dao) GetSharing(limit, offset int) []Sharing {
	var sharing []Sharing
	if err := db.Select(&sharing, "SELECT * FROM issue WHERE content != '' ORDER BY updated_at DESC LIMIT ? OFFSET ?", limit, offset); err != nil {
		log.Printf("failed to get latest %d sharing: %s", limit, err)
		return nil
	}

	return sharing
}

// GetSharingWithLimit 获取分享
func (d *Dao) GetSharingWithLimit(limit int) []Sharing {
	var sharing []Sharing
	if err := db.Select(&sharing, "SELECT * FROM issue WHERE content != '' ORDER BY updated_at DESC LIMIT ?", limit); err != nil {
		log.Printf("failed to get latest %d sharing: %s", limit, err)
		return nil
	}

	return sharing
}

// GetAllNotes 获取所有随想
func (d *Dao) GetAllNotes() []Note {
	var notes []Note
	if err := db.Select(&notes, "SELECT * FROM note ORDER BY updated_at DESC"); err != nil {
		log.Printf("failed to get all notes: %s", err)
		return nil
	}

	return notes
}

// GetNotes 获取分享
func (d *Dao) GetNotes(limit, offset int) []Note {
	var notes []Note
	if err := db.Select(&notes, "SELECT * FROM note ORDER BY updated_at DESC LIMIT ? OFFSET ?", limit, offset); err != nil {
		log.Printf("failed to get all notes: %s", err)
		return nil
	}

	return notes
}

// GetLatestSharing 获取最后一条分享
func (d *Dao) GetLatestSharing() (Sharing, error) {
	var sharing Sharing
	err := db.Get(&sharing, "SELECT * FROM issue ORDER BY updated_at DESC LIMIT 1")
	if err != nil {
		log.Printf("failed to get latest sharing: %s", err)
	}

	return sharing, err
}

// AddSharing 增加一条分享
func (d *Dao) AddSharing(url string) error {
	tx := db.MustBegin()
	now := time.Now()
	tx.MustExec("INSERT INTO issue(url, content, created_at, updated_at) VALUES (?, '', ?, ?)", url, now, now)
	return tx.Commit()
}

// CommentLatestSharing 评论最近一条分享
func (d *Dao) CommentLatestSharing(comment string) error {
	var sharing Sharing

	tx := db.MustBegin()

	if err := tx.Get(&sharing, "SELECT * FROM issue ORDER BY updated_at DESC LIMIT 1"); err != nil {
		log.Printf("failed to get latest sharing: %s", err)
		return err
	}
	tx.MustExec("UPDATE issue SET content=?, updated_at=? WHERE id = ?", comment, time.Now(), sharing.ID)
	return tx.Commit()
}

// AddNote 增加一条随想
func (d *Dao) AddNote(content string) error {
	now := time.Now()

	tx := db.MustBegin()
	tx.MustExec("INSERT INTO note(content, created_at, updated_at) VALUES(?, ?, ?)", content, now, now)

	return tx.Commit()
}
