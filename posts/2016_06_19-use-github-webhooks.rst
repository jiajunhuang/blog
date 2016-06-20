利用Github的WebHook完成自动部署
================================

之前我写了一个用rst来写博客的项目，但是每次写完都要自己push，然后到服务器pull。
再加上我采取的方式是把原来的写博客的repo作为这个项目的submodule。每次手动那就
更加麻烦了。所以通过阅读Github的文档，为这个项目添加了webhook的功能。

https://developer.github.com/webhooks/ 上有详细的文档，概括一下就是说，在某个
组织或者项目的settings里设置webhooks以后，例如设置对 ``git push`` 请求进行hook，
当你对这个项目进行 ``git push`` 以后，github就会对你在settings里填的hook地址发起
一个POST请求。

Github发起的这个请求的Header中会包含以下特殊值:

=================== ====================================================================
头部                 描述
------------------- --------------------------------------------------------------------
X-Github-Event       这次请求的发起原因，例如 create, delete。详见 github events [#]_ 。
X-Hub-Signature      在Hook里填的secret key，加密以后的值
X-Github-Delivery    这次请求的唯一ID
=================== ====================================================================

并且User-Agent中会以 ``Github-Hookshot/`` 开头。

所以在我们新建一个controller专门用来处理hook，代码如下:

.. code:: python

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
                raise tornado.web.HTTPError(500, "secret.txt does not exists")

            data = self.request.body
            if not self._validate_signature(data):
                raise tornado.web.HTTPError(500, "signature not valid")

            # 下面的操作是阻塞的-。- 暂且不用celery试试看
            repo = git.Repo(Config().repo_path)
            origin = repo.remotes.origin
            origin.pull()

            for submodule in repo.submodules:
                submodule.update(init=True)

            # 然后重启进程
            tornado.autoreload.start()

        def _validate_signature(self, data):
            sha_name, signature = self.request.headers.get('X-Hub-Signature').split('=')
            if sha_name != 'sha1':
                return False

            # HMAC requires its key to be bytes, but data is strings.
            mac = hmac.new(bytes(Config().github_secret_key, "utf-8"), msg=data, digestmod=hashlib.sha1)
            return hmac.compare_digest(mac.hexdigest(), signature)

其中的 ``def get`` 是用来响应 github webhooks的ping请求 [#]_ 的。

我们住要来看一下post的代码，首先，我们需要在项目的根目录放置我们在github填入的
secret key，保存为 ``secret.txt`` 文件，这是用来校验用的。然后当请求到来以后，
首先我们看有没有 ``secret.txt`` 如果没有，抛出异常，这样Github能直接检测到我们的
错误并且记录下来。如果有的话，接下来进行secret key的校验，如果失败了也抛出异常
让Github监测到（当然也可能是让第三方拿到）。

接下来的操作就是从github拉代码，然后重启进程。但是这个过程是阻塞的，例如git的网络操作
和磁盘操作，都会把Tornado的进程给阻塞住。解决方案有很多中，比如把git的操作用celery
来完成，但是为了保持项目精简，还是让他阻塞了。另外其实还有方法就是支一个线程出来
让线程完成git的操作。这样会稍微好一点点。不过目前就本人的测试来说，上述操作
能控制在500ms内（因为vps在国外，连github老快了）本人还是可以接受的。

.. [#] https://developer.github.com/webhooks/#events
.. [#] https://developer.github.com/webhooks/#ping-event
