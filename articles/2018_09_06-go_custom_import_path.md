# Go的custom import path

什么是 custom import path？就是 `package` 后面的注释，注释的内容是 `// import xxxx`。这就是传说中的 custom import path。

例如 `Docker` 中的 [代码](https://github.com/moby/moby/blob/master/daemon/configs.go)：

```go
package daemon // import "github.com/docker/docker/daemon"

import (
	swarmtypes "github.com/docker/docker/api/types/swarm"
	"github.com/sirupsen/logrus"
)

// SetContainerConfigReferences sets the container config references needed
func (daemon *Daemon) SetContainerConfigReferences(name string, refs []*swarmtypes.ConfigReference) error {
	if !configsSupported() && len(refs) > 0 {
		logrus.Warn("configs are not supported on this platform")
		return nil
	}

	c, err := daemon.GetContainer(name)
	if err != nil {
		return err
	}
	c.ConfigReferences = append(c.ConfigReferences, refs...)
	return nil
}
```

第一行，就是 custom import path。因为目前 `github.com/docker/docker` 已经重命名为了 `github.com/moby/moby`。所以如果
想愉快的补全代码或者是用IDE分析代码，正确的做法是，拷贝代码到本地，然后把 `moby/moby` 的路径改成 `docker/docker`。

custom import path的出现就是为了防止仓库名发生改变之后，无法导入，然后制定的一种方案。好了，到了吐槽时间，这确实是属于
Go的设计不合理的地方之一。

----------

- https://golang.org/pkg/cmd/go/#hdr-Import_path_checking
- https://docs.google.com/document/d/1jVFkZTcYbNLaTxXD9OcGfn7vYv5hWtPx9--lTx1gPMs/edit
