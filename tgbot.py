import gevent.monkey
gevent.monkey.patch_all()  # noqa

import logging
logging.basicConfig(level=logging.INFO)  # noqa

import threading

from telegram.ext import Updater, CommandHandler, MessageHandler, Filters
from telegram import MessageEntity

from models import (
    get_session,
    Issue,
    Note,
)
from config import config


AUTHORS_FILTER = Filters.user(username="@jiajunhuang")
TGBOT_USER_ID = 4
TGBOT_TOPIC_ID = 7  # 好物分享


def report_error(func):
    def wrapper(bot, update, *args, **kwargs):
        try:
            return func(bot, update, *args, **kwargs)
        except Exception as e:
            logging.exception("failed to handle message from telegram")
            bot.send_message(chat_id=update.message.chat_id, text="出错啦：" + str(e))
    return wrapper


def save_url(url):
    with get_session() as s:
        issue = Issue(url=url)
        s.add(issue)
        s.flush()
        return issue.id


def save_note(content):
    with get_session() as s:
        note = Note(content=content)
        s.add(note)
        s.flush()
        return note.id


def save_comment(comment):
    with get_session() as s:
        issue = Issue.get_latest_one(s)
        if issue:
            issue.content = comment
            s.add(issue)
            return "{}: {}#{}".format(comment, config.SHARE_BOT_URL, issue.id)

        return "not found"


def update_comment(issue_id, comment):
    with get_session() as s:
        issue = Issue.get_by_id(issue_id)
        if issue:
            issue.content = comment
            s.add(issue)
            return "mapped with url: " + issue.id

        return "not found"


@report_error
def comment_handler(bot, update, args):
    if len(args) == 0:
        text = "Usage: /comment <your comments>"
        bot.send_message(chat_id=update.message.chat_id, text=text)
    else:
        text = save_comment(" ".join(args))
        bot.send_message(chat_id="@jiajunhuangcom", text=text)  # send to channel


@report_error
def update_comment_handler(bot, update, args):
    if len(args) == 0:
        text = "Usage: /update <id> <new comments>"
    else:
        text = update_comment(int(args[0]), "".join(args[1:]))

    bot.send_message(chat_id=update.message.chat_id, text=text)


@report_error
def url_share_handler(bot, update):
    bot.send_message(chat_id=update.message.chat_id, text="saved with id: {}".format(save_url(update.message.text)))


@report_error
def note_share_handler(bot, update):
    bot.send_message(chat_id=update.message.chat_id, text="saved with id: {}".format(save_note(update.message.text)))


def share_bot():
    updater = Updater(token=config.SHARE_BOT_TOKEN)
    dispatcher = updater.dispatcher
    dispatcher.add_handler(
        CommandHandler(
            'comment', comment_handler, pass_args=True, filters=AUTHORS_FILTER,
        )
    )
    dispatcher.add_handler(
        CommandHandler(
            'update', update_comment_handler, pass_args=True, filters=AUTHORS_FILTER,
        ),
    )
    dispatcher.add_handler(MessageHandler(
        Filters.text & (
            Filters.entity(MessageEntity.URL) | Filters.entity(MessageEntity.TEXT_LINK)
        ) & AUTHORS_FILTER,
        url_share_handler,
    ))
    updater.start_polling()


def note_bot():
    updater = Updater(token=config.NOTE_BOT_TOKEN)
    dispatcher = updater.dispatcher
    dispatcher.add_handler(MessageHandler(
        Filters.text & AUTHORS_FILTER,
        note_share_handler,
    ))
    updater.start_polling()


if __name__ == "__main__":
    share_thread = threading.Thread(target=share_bot)
    note_thread = threading.Thread(target=note_bot)

    share_thread.start()
    logging.info("share_thread already start!")
    note_thread.start()
    logging.info("note_thread already start!")

    share_thread.join()
    note_thread.join()
