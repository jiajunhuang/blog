import os


class Config:
    def __init__(self):
        self.ALIPAY_URL = os.getenv("ALIPAY_URL", "https://jiajunhuang.com")
        self.WECHAT_PAY_URL = os.getenv("WECHAT_PAY_URL", "https://jiajunhuang.com")


config = Config()
