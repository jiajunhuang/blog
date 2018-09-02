# 写了一个Telegram Bot：自动化分享高质量内容

平时我们读到优秀的资源例如文章，视频或者电子书等等，总是会忍不住收藏起来，但是如果我们能分享出来给所有人看到，那会不会
更好呢？

所以我做了一个Telegram Bot，平时我只要把阅读到的不错的链接分享给我的bot，然后选择性的添加一些我自己的评语（或者叫推荐
理由）。然后会有另外一个页面来动态的渲染所有我分享的链接和评语。

欢迎直接访问：https://share.jiajunhuang.com

## Bot代码

首先你要去Telegram上和BotFather聊天，其实就是输入一些命令，来创建一个你自己的Bot，它会下发给你一个Token。然后参考
https://core.telegram.org/bots/api 开始开发。当然，为了方便省事儿，我是直接用的一个Python封装好的包来做的。Bot的核心代码
如下：

```python
import gevent.monkey
gevent.monkey.patch_all()  # noqa

import logging
logging.basicConfig(level=logging.INFO)  # noqa

from telegram.ext import Updater, CommandHandler, MessageHandler, Filters
from telegram import MessageEntity

from models import get_session, URLShare
from config import config


AUTHORS_FILTER = Filters.user(username="@jiajunhuang")


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
        url_share = URLShare(url=url)
        s.add(url_share)
        s.flush()
        return url_share.id


def save_comment(comment):
    with get_session() as s:
        share = s.query(URLShare).order_by(URLShare.id.desc()).first()
        if share:
            share.comment = comment
            s.add(share)
            return "mapped with url: " + share.url

        return "not found"


def update_comment(share_id, comment):
    with get_session() as s:
        share = s.query(URLShare).filter(URLShare.id == share_id).first()
        if share:
            share.comment = comment
            s.add(share)
            return "mapped with url: " + share.url

        return "not found"


@report_error
def comment_handler(bot, update, args):
    if len(args) == 0:
        text = "Usage: /comment <your comments>"
    else:
        text = save_comment("".join(args))

    bot.send_message(chat_id=update.message.chat_id, text=text)


@report_error
def update_comment_handler(bot, update, args):
    if len(args) == 0:
        text = "Usage: /update <id> <new comments>"
    else:
        text = update_comment(int(args[0]), "".join(args[1:]))

    bot.send_message(chat_id=update.message.chat_id, text=text)


@report_error
def url_share_handler(bot, update):
    bot.send_message(chat_id=update.message.chat_id, text="save with id: {}".format(save_url(update.message.text)))


if __name__ == "__main__":
    updater = Updater(token=config.TGBOTTOKEN)
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
```

models定义如下：

```python
import datetime
import contextlib

from sqlalchemy import create_engine, Column, DateTime, Integer, String
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker

from config import config

engine = create_engine(config.SQLALCHEMY_DB_URI, echo=config.SQLALCHEMY_ECHO)
Session = sessionmaker(bind=engine)

Base = declarative_base()


class BaseMixin:
    id = Column(Integer, primary_key=True, autoincrement=True)

    created_at = Column(DateTime, nullable=False, default=datetime.datetime.now)
    updated_at = Column(DateTime, nullable=False, default=datetime.datetime.now, onupdate=datetime.datetime.now)
    deleted_at = Column(DateTime, nullable=True, index=True)


@contextlib.contextmanager
def get_session():
    s = Session()
    try:
        yield s
        s.commit()
    except Exception:
        s.rollback()
        raise
    finally:
        s.close()


class URLShare(Base, BaseMixin):
    __tablename__ = "url_share"

    url = Column(String(1024), nullable=False)
    comment = Column(String(512))
```

当然上面的代码里 `config.py` 里的内容我就不贴出来了，毕竟为了简单方便，我直接吧token和数据库URL写到了代码里，在实际
工作上，这是 **不好** 的习惯，请不要学，谢谢。

数据库用的是SQLite。为啥不用MySQL或者PG？答：为啥要用大炮打蚊子？而且，SQLite没有你想象中的那么弱。

## 当然要支持RSS

虽然是一个简单到不行的网页，但是为了自动化方便，那也是要支持RSS的！

访问 https://share.jiajunhuang.com/rss 获取feed。有了RSS，你可以选择在IFTTT上增加有新的订阅时就给你发送一条消息，这样
当我有新的分享时，你就可以自动收到推送啦 :)


-----------

- https://core.telegram.org/bots
- https://core.telegram.org/bots/api
