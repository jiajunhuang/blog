import os


class Config:
    def __init__(self):
        self.NOTES_URL = os.getenv("NOTES_URL", "http://127.0.0.1:13233/v1/notes")


config = Config()
