package main

import (
	"database/sql"
	"time"
)

var (
	dao *Dao
)

type Dao struct{}

type Note struct {
	ID        int        `db:"id"`
	Content   string     `db:"content"`
	CreatedAt *time.Time `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

type Sharing struct {
	ID        int            `db:"id"`
	Content   string         `db:"content"`
	URL       sql.NullString `db:"url"`
	CreatedAt *time.Time     `db:"created_at"`
	UpdatedAt *time.Time     `db:"updated_at"`
	DeletedAt *time.Time     `db:"deleted_at"`
}

func (d *Dao) GetAllSharing() []Sharing {
	var sharing []Sharing
	db.Select(&sharing, "SELECT * FROM issue ORDER BY updated_at DESC")

	return sharing
}

func (d *Dao) GetSharingWithLimit(limit int) []Sharing {
	var sharing []Sharing
	db.Select(&sharing, "SELECT * FROM issue ORDER BY updated_at DESC LIMIT $1", limit)

	return sharing
}

func (d *Dao) GetAllNotes() []Note {
	var notes []Note
	db.Select(&notes, "SELECT * FROM note ORDER BY updated_at DESC")

	return notes
}

func (d *Dao) AddSharing(url string) {
	tx := db.MustBegin()
	tx.MustExec("")
	tx.Commit()
}

func (d *Dao) CommentLatestSharing(comment string) {
	tx := db.MustBegin()
	tx.MustExec("")
	tx.Commit()
}

func (d *Dao) AddNote(content string) {

}
