import os


class Config:
    def __init__(self):
        self.ALIPAY_URL = os.getenv("ALIPAY_URL", "https://jiajunhuang.com")
        self.WECHAT_PAY_URL = os.getenv("WECHAT_PAY_URL", "https://jiajunhuang.com")
        self.SQLALCHEMY_DB_URI = os.getenv("SQLALCHEMY_DB_URI", "sqlite:////data/db/blog.db")
        self.SQLALCHEMY_ECHO = os.getenv("SQLALCHEMY_ECHO") == "True"
        self.SHARE_BOT_URL = os.getenv("SHARE_BOT_URL", "http://127.0.0.1:5000/sharing")
        self.SHARE_BOT_TOKEN = os.getenv("SHARE_BOT_TOKEN", "")
        self.NOTE_BOT_TOKEN = os.getenv("NOTE_BOT_TOKEN", "")


config = Config()
