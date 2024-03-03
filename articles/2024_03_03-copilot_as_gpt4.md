# 有GitHub Copilot？那就可以搭建你的ChatGPT4服务

我有Github Copilot，也订阅了GPT Plus，GPT Plus 20刀每月，我看了一下其实我用的不是特别多。本着开猿节流，降本增笑的精神，
我停止续订了GPT Plus，并且着手于找到 GPT Plus 的替代方案。

## 方案1：直接使用API Key搭建

如果你有 API Key，那么可以直接使用 API Key来搭建，这个很简单，方案也很多，比如川虎，lobe-chat, ChatGPT-Next-Web，然后
输入你的 API Key就可以。

> 我有 API Key，但是折腾是一种乐趣，所以我选择了方案2。

## 方案2：把 Github Copilot 转化成 API

之前在Hacker News看到过，Github Copilot底层其实就是用的OpenAI，因此有人开发了将 Github 接口转化成 OpenAI API 的服务。
因此我们想要搭建一套自己私有的 ChatGPT Web的话，需要3步：

1. 获取 Github Copilot Token
2. 搭建 Github Copilot 转化为 OpenAI API 的客户端并且配置域名
3. 搭建并配置 ChatGPT-Next-Web 或其他Web客户端

第一步，其实比较难，现在网上很多方案都失效了，但是记住一点，既然 Copilot 要在本地使用，而Copilot服务端是要鉴权的，
本地就必然保存了Token在磁盘上，要不然它下次咋鉴权呢？我用的是 vim-copilot，所以我直接在家目录找：

```bash
$ find .config -name '*copilot*'
.config/github-copilot
.config/gh-copilot
.config/nvim/plugged/copilot.vim
.config/nvim/plugged/copilot.vim/autoload/copilot
.config/nvim/plugged/copilot.vim/autoload/copilot.vim
.config/nvim/plugged/copilot.vim/lua/_copilot.lua
.config/nvim/plugged/copilot.vim/doc/copilot.txt
.config/nvim/plugged/copilot.vim/plugin/copilot.vim
.config/nvim/plugged/copilot.vim/syntax/copilot.vim
```

然后一个一个翻，果然，配置文件就藏在 `.config/github-copilot/hosts.json` 里，找出来，是一个 `ghu_` 开头的token，保存。

第二步，搭建Copilot 转化为 API 的服务，我直接用Docker：

```bash
docker run -d \
    -e COPILOT_TOKEN=<刚才找到的Copilot里的token，ghu_ 开头的那个> \
    -e SUPER_TOKEN=<自定义的token，等会儿给 ChatGPT-Next-Web使用> \
    -e ENABLE_SUPER_TOKEN=true \
    --name copilot-gpt4-service \
    --restart always \
    -p 8080:8080 \
    aaamoon/copilot-gpt4-service:latest
```

为了方便，我配置了一个域名，假设为 `https://openai.example.com`。

第三步，搭建 `ChatGPT-Next-Web`，同样，我直接使用 Docker:

```bash
docker run -d -p 3000:3000 \
    -e BASE_URL=<你配置的域名> \
    -e OPENAI_API_KEY=<刚才设置的 SUPER_TOKEN，也就是自定义的token> \
    -e CODE=<等于一个登录密码，防止 ChatGPT-Next-Web 被他人滥用> \
    yidadaa/chatgpt-next-web
```

同样为了方便，为 ChatGPT-Next-Web 也配置一个域名，然后就可以访问了。

访问以后，在设置里，设置登录密码，将模型改为 `GPT4` 或者 `GPT4-Turbo`，将界面语言改为中文，如果想要同步聊天记录的话，
可以配置上自己的 webdav。

至此，大功告成！注意一点，就是搭建出来的服务，仅限于自己使用，不要到处发，否则使用频率太高，被Github检测出来，那有可能
就被拉黑了以后就无法使用Copilot了💔

---

Refs:

- https://github.com/aaamoon/copilot-gpt4-service
- https://github.com/GaiZhenbiao/ChuanhuChatGPT
- https://github.com/lobehub/lobe-chat
- https://github.com/ChatGPTNextWeb/ChatGPT-Next-Web
