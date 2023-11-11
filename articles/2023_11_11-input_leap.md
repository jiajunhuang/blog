# macOS/Linux 编译 InputLeap

上一篇说到，我安装了 synergy-core，但是是下载别人编译好的二进制。我就想着，反正我也能编译，干脆改改源码，把Synergy里
烦人的激活提示框干掉得了？但是又想到，Barrier本来就是改好了的，只是不维护了，InputLeap 又是 Barrier 的继任，那我就
直接自己编译 InputLeap 试试看了。

其实参照WIKI就可以，但是WIKI有点点坑没有指出来： https://github.com/input-leap/input-leap/wiki/Building-on-macOS

## 安装依赖

首先需要安装依赖，WIKI没有列全：

1. 安装 Homebrew

```bash
$ /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

2. 安装 xcode

去 AppStore 安装 Xcode，然后打开，同意协议。

3. 安装依赖

```bash
$ brew install qt5 openssl pkg-config cmake
```

4. Clone 代码

```bash
$ git clone https://github.com/input-leap/input-leap
```

5. 编译

```bash
$ cd input-leap
$ export B_BUILD_TYPE=Release
$ ./clean_build.sh
```

6. 签名

```bash
$ codesign --force --deep --sign - ./build/bundle/InputLeap.app
```

注意:

- 其中第5步可能会失败，报错 `ld ... System` 之类的，就是说链接系统库失败，这个时候上面一般会有一个 Warning，
说是系统库没有找到，修改 `clean_build.sh` 中

`B_CMAKE_FLAGS="${B_CMAKE_FLAGS} -DCMAKE_OSX_SYSROOT=$(xcode-select --print-path)/Platforms/MacOSX.platform/Developer/SDKs/MacOSX.sdk -DCMAKE_OSX_DEPLOYMENT_TARGET=10.9"
`

这一行，把 `$(xcode-select --print-path)/Platforms/MacOSX.platform/Developer/SDKs/MacOSX.sdk` 替换成实际路径，一般来说，都是

`/Applications/Xcode.app/Contents/Developer/Platforms/MacOSX.platform/Developer/SDKs/MacOSX14.sdk/`

具体取决于你的系统版本。然后顺手把 `-DCMAKE_OSX_DEPLOYMENT_TARGET=10.9` 改成你的系统版本，例如 `-DCMAKE_OSX_DEPLOYMENT_TARGET=14`，
再执行编译。

- 第6步，我本来想一台机器编译，直接拷贝到另一个机器使用，但是发现不行，因为整个App不是做静态链接，所以还是需要哪个机器使用，哪个机器就要安装对应依赖，按道理应该可以做成静态链接的。

剩下的就是常规的操作了，拷贝 `build/bundle/` 下 `InputLeap.app` 到 `Applications` 中，打开，给权限，配置。

## Linux 编译

Linux 编译就要简单很多，照着WIKI来就可以：

```bash
sudo apt update && sudo apt upgrade
sudo apt install git cmake make xorg-dev g++ libcurl4-openssl-dev \
                 libavahi-compat-libdnssd-dev libssl-dev libx11-dev \
                 qttools5-dev qtbase5-dev

git clone https://github.com/input-leap/input-leap.git
# This builds from master, but you can also checkout a tag or branch.
cd input-leap
git submodule update --init --recursive
export B_BUILD_TYPE=Release
./clean_build.sh
sudo cmake --install build # install to /usr/local/
```

然后就是创建system service文件：

```systemd
[Unit]
Description=Soft KVM Client
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=<改成你的名字>
ExecStart=/usr/bin/unbuffer /usr/local/bin/input-leapc -f --display :0 <macOS IP地址>:24800
Environment=XAUTHORITY=/var/run/lightdm/root/:0
Restart=always
RestartSec=3

[Install]
WantedBy=graphical.target
```

启动。

---

refs:

- https://github.com/input-leap/input-leap/wiki/Building-on-macOS
- https://github.com/input-leap/input-leap/wiki/Building-on-Linux
