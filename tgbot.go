package main

import (
	"fmt"
	"log"
	"os"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

var (
	sharingURL = os.Getenv("SHARE_BOT_URL")
)

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
		if !(m.Private() && m.Sender.Username == "jiajunhuang") {
			return
		}

		if err := dao.CommentLatestSharing(m.Payload); err != nil {
			b.Send(m.Sender, fmt.Sprintf("failed to comment: %s", err))
			return
		}

		// 如果没有出错，就发到channel
		latestSharing, err := dao.GetLatestSharing()
		if err != nil {
			log.Printf("failed to send to channel: %s", err)
			return
		}
		msg := fmt.Sprintf("%s: %s#%d", latestSharing.Content, sharingURL, latestSharing.ID)

		channel, err := b.ChatByID("@jiajunhuangcom")
		if err != nil {
			log.Printf("failed to send to channel: %s", err)
			return
		}
		_, err = b.Send(channel, msg)
		if err != nil {
			log.Printf("failed to send to channel: %s", err)
			return
		}
	})
	b.Handle(tb.OnText, func(m *tb.Message) {
		if !(m.Private() && m.Sender.Username == "jiajunhuang") {
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
		if !(m.Private() && m.Sender.Username == "jiajunhuang") {
			return
		}

		dao.AddNote(m.Text)
		b.Send(m.Sender, "saved")
	})

	b.Start()
}
