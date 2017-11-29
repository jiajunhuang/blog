# Git HTTPS 如何保存密码

读了一下 `git-credential-store` 的manual。

在项目下执行 `git config credential.helper store` 然后 `git push` 输入密码之后，家目录下就会多出一个文件：

`.git-credentials`

这个文件明文保存了https克隆所用的用户名和密码。

如果想要更加安全的话，可以改用 `git config credential.helper cache`。这个默认会在900s之后过期。所以相对更加安全。
