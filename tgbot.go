package main

import (
	"os"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

func startSharingBot() {
	b, err := tb.NewBot(tb.Settings{
		Token:  os.Getenv("SHARING_TGBOT_TOKEN"),
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		sugar.Fatalf("failed to start telegram bot: %s", err)
		return
	}

	b.Handle("/comment", func(m *tb.Message) {
		if !(m.Private() && m.Sender.Username == "jiajunhuang") {
			return
		}

		dao.CommentLatestSharing(m.Text)
		b.Send(m.Sender, "commented")
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
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		sugar.Fatalf("failed to start telegram bot: %s", err)
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
