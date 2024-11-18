package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

var (
	sharingURL = os.Getenv("SHARE_BOT_URL")

	notifyURL   = os.Getenv("NOTIFY_URL")
	notifyToken = os.Getenv("NOTIFY_TOKEN")

	botAdminUserID = os.Getenv("BOT_ADMIN_USER_ID")
)

// sendNotifyToApp 往推送发一个通知
func sendNotifyToApp(brief string) {
	body := map[string]string{"token": notifyToken, "title": "发布了一篇新的博客", "brief": brief, "route": "https://jiajunhuang.com"}
	jsonBytes, err := json.Marshal(body)
	if err != nil {
		log.Printf("failed to marshal json %+v: %s", body, err)
		return
	}
	resp, err := http.Post(notifyURL, "application/json", bytes.NewReader(jsonBytes))
	if err != nil {
		log.Printf("failed to notify system: %s", err)
		return
	}
	defer resp.Body.Close()
	log.Printf("successfully notify the system")
}

func startSharingBot() {
	b, err := tb.NewBot(tb.Settings{
		Token:  os.Getenv("SHARE_TGBOT_TOKEN"),
		URL:    "https://api.telegram.org",
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatalf("failed to start telegram bot: %s", err)
		return
	}

	b.Handle("/comment", func(m *tb.Message) {
		if !(m.Private() && fmt.Sprintf("%d", m.Sender.ID) == botAdminUserID) {
			return
		}

		if err := dao.CommentLatestSharing(m.Payload); err != nil {
			b.Send(m.Sender, fmt.Sprintf("failed to comment: %s", err))
			return
		}

		b.Send(m.Sender, "commented")
	})
	b.Handle(tb.OnChannelPost, func(m *tb.Message) {
		log.Printf("received channel message %+v", m)
		if m.FromChannel() {
			log.Printf("gonna send request to notify system")
			sendNotifyToApp(m.Text)
			return
		}
	})

	b.Handle(tb.OnText, func(m *tb.Message) {
		log.Printf("received text message %+v", m)
		if !(m.Private() && fmt.Sprintf("%d", m.Sender.ID) == botAdminUserID) {
			return
		}

		dao.AddSharing(m.Text)
		b.Send(m.Sender, "saved")
	})

	b.Start()
}

func startNoteBot() {
	b, err := tb.NewBot(tb.Settings{
		Token:  os.Getenv("NOTE_TGBOT_TOKEN"),
		URL:    "https://api.telegram.org",
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatalf("failed to start telegram bot: %s", err)
		return
	}

	b.Handle(tb.OnText, func(m *tb.Message) {
		if !(m.Private() && fmt.Sprintf("%d", m.Sender.ID) == botAdminUserID) {
			return
		}

		dao.AddNote(m.Text)
		b.Send(m.Sender, "saved")
	})

	b.Start()
}
