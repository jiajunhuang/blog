package main

import (
	"fmt"
	"os"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

var (
	sharingURL = os.Getenv("SHARE_BOT_URL")
)

type Channel struct{}

func (c *Channel) Recipient() string {
	return "jiajunhuangcom"
}

func startSharingBot() {
	b, err := tb.NewBot(tb.Settings{
		Token:  os.Getenv("SHARE_TGBOT_TOKEN"),
		URL:    "https://api.telegram.org",
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

		err := dao.CommentLatestSharing(m.Text)
		b.Send(m.Sender, fmt.Sprintf("commented with error: %s", err))

		// 如果没有出错，就发到channel
		if err == nil {
			latestSharing, err := dao.GetLatestSharing()
			if err != nil {
				b.Send(m.Sender, fmt.Sprintf("failed to send to channel: %s", err))
				return
			}
			msg := fmt.Sprintf("%s: %s#%d", latestSharing.Content, sharingURL, latestSharing.ID)

			b.Send(&Channel{}, msg)
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
