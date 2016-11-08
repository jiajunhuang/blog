# coding: utf-8

import os
import hmac
import hashlib

import tornado.web
import git

from config import Config


class GithubWebHooksHandler(tornado.web.RequestHandler):
    def get(self):  # github webhooks ping
        self.finish()

    def post(self):
        if not os.path.exists(Config().github_webhook_secret_path):
            self.write({
                "error": "secret.txt not set"
            })
            raise tornado.web.HTTPError(500, "secret.txt does not exists")

        data = self.request.body
        if not self._validate_signature(data):
            raise tornado.web.HTTPError(500, "signature not valid")

        # 下面的操作是阻塞的-。- 暂且不用celery试试看
        repo = git.Repo(Config().repo_path)
        origin = repo.remotes.origin
        origin.pull()

        # 更新目录
        Config().reload_catalog()

        # 然后重启进程，因为有可能更新的是项目代码而非博客
        tornado.autoreload.start()

    def _validate_signature(self, data):
        sha_name, signature = self.request.headers.get('X-Hub-Signature').split('=')
        if sha_name != 'sha1':
            return False

        # HMAC requires its key to be bytes, but data is strings.
        mac = hmac.new(bytes(Config().github_secret_key, "utf-8"), msg=data, digestmod=hashlib.sha1)
        return hmac.compare_digest(mac.hexdigest(), signature)
