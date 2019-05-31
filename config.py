import os


class Config:
    def __init__(self):
        self.NOTES_URL = os.getenv("NOTES_URL", "http://127.0.0.1:13233/v1/notes")
        self.ALIPAY_URL = os.getenv("ALIPAY_URL", "https://jiajunhuang.com")
        self.WECHAT_PAY_URL = os.getenv("WECHAT_PAY_URL", "https://jiajunhuang.com")


config = Config()
