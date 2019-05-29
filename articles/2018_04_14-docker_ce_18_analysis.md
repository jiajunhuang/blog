# Docker CE 18.03源码阅读与分析

Docker更新的速度太快了。网上很多分析已经过时了。今天我翻了一下Docker的源码，发现其实Docker的源码挺
简单易懂的。当然前提是，你已经熟练掌握了Docker，经常用，知道有哪些概念，了解其原理。有了这些做基础，
就很容易读了。

建议先读一下这篇文章：[自己写一个容器](https://jiajunhuang.com/articles/2018_03_09-write_you_a_container.md.html) 和 [这个项目](https://github.com/jiajunhuang/cup) 的源代码。

## 目录结构

```bash
$ tree -d -L 1
.
├── api
├── builder
├── cli
├── client
├── cmd  # 命令行的入口，我们就要从这里跳进去看
├── container  # 容器的抽象
├── contrib  # 杂物堆，啥杂七杂八的都往这里丢
├── daemon  # 今天的主角，dockerd这个daemon
├── distribution  # 看起来发布相关的
├── dockerversion
├── docs  # 文档
├── errdefs  # 一些常见的错误
├── hack
├── image  # 镜像的抽象概念
├── integration
├── integration-cli
├── internal
├── layer  # 层。怎么说呢，就是对各种layer fs的抽象
├── libcontainerd
├── migrate
├── oci  # https://blog.docker.com/2017/07/demystifying-open-container-initiative-oci-specifications/ 容器标准库相关
├── opts  # 一些配置和配置校验相关的
├── pkg  # 类似于我们平时写的utils或者helpers等
├── plugin  # 插件相关的东西
├── profiles
├── project
├── reference
├── registry
├── reports
├── restartmanager  # 负责容器的重启，比如是否设置了"always"呀
├── runconfig
├── vendor  # go的vendor机制
└── volume  # volume相关
```

- 负责存储配置的一般都叫 `xx Store`
- Docker的设计是单机的，不是分布式的
- Docker的设计是Client-Server模式的，平时我们用的docker这个命令被分散到 https://github.com/docker/cli 这个仓库去了

## 从命令行进入

入口在 `cmd/dockerd/docker.go`:

```go
func newDaemonCommand() *cobra.Command {
        opts := newDaemonOptions(config.New())

        cmd := &cobra.Command{
                Use:           "dockerd [OPTIONS]",
                Short:         "A self-sufficient runtime for containers.",
                SilenceUsage:  true,
                SilenceErrors: true,
                Args:          cli.NoArgs,
                RunE: func(cmd *cobra.Command, args []string) error {
                        if opts.version {
                                showVersion()
                                return nil
                        }
                        opts.flags = cmd.Flags()
                        return runDaemon(opts) // 真正的入口
                },
        }
        cli.SetupRootCommand(cmd)

        flags := cmd.Flags()
        flags.BoolVarP(&opts.version, "version", "v", false, "Print version information and quit")
        flags.StringVar(&opts.configFile, "config-file", defaultDaemonConfigFile, "Daemon configuration file")
        opts.InstallFlags(flags)
        installConfigFlags(opts.daemonConfig, flags)
        installServiceFlags(flags)

        return cmd
}
```

然后跳到 `cmd/dockerd/daemon.go`：

```go
func (cli *DaemonCli) start(opts *daemonOptions) (err error) {
	stopc := make(chan bool)
	defer close(stopc)

	// warn from uuid package when running the daemon
	uuid.Loggerf = logrus.Warnf

	// 设置一些默认配置例如TLS啊blabla
	opts.SetDefaultOptions(opts.flags)

	// 加载配置
	if cli.Config, err = loadDaemonCliConfig(opts); err != nil {
		return err
	}
	cli.configFile = &opts.configFile
	cli.flags = opts.flags

	if cli.Config.Debug {
		debug.Enable()
	}

	if cli.Config.Experimental {
		logrus.Warn("Running experimental build")
	}

	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: jsonmessage.RFC3339NanoFixed,
		DisableColors:   cli.Config.RawLogs,
		FullTimestamp:   true,
	})

	// LCOW: Linux Containers On Windows. ref: https://blog.docker.com/2017/09/preview-linux-containers-on-windows/
	system.InitLCOW(cli.Config.Experimental)

	// 设置默认的umask。022。也就是说，创建的文件权限都是755
	if err := setDefaultUmask(); err != nil {
		return fmt.Errorf("Failed to set umask: %v", err)
	}

	// Create the daemon root before we create ANY other files (PID, or migrate keys)
	// to ensure the appropriate ACL is set (particularly relevant on Windows)
	// 设置daemon的执行时候的根目录
	if err := daemon.CreateDaemonRoot(cli.Config); err != nil {
		return err
	}

	if cli.Pidfile != "" {
		pf, err := pidfile.New(cli.Pidfile)
		if err != nil {
			return fmt.Errorf("Error starting daemon: %v", err)
		}
		defer func() {
			if err := pf.Remove(); err != nil {
				logrus.Error(err)
			}
		}()
	}

	// TODO: extract to newApiServerConfig()
	serverConfig := &apiserver.Config{
		Logging:     true,
		SocketGroup: cli.Config.SocketGroup,
		Version:     dockerversion.Version,
		CorsHeaders: cli.Config.CorsHeaders,
	}

	// 是否走TLS
	if cli.Config.TLS {
		tlsOptions := tlsconfig.Options{
			CAFile:             cli.Config.CommonTLSOptions.CAFile,
			CertFile:           cli.Config.CommonTLSOptions.CertFile,
			KeyFile:            cli.Config.CommonTLSOptions.KeyFile,
			ExclusiveRootPools: true,
		}

		if cli.Config.TLSVerify {
			// server requires and verifies client's certificate
			tlsOptions.ClientAuth = tls.RequireAndVerifyClientCert
		}
		tlsConfig, err := tlsconfig.Server(tlsOptions)
		if err != nil {
			return err
		}
		serverConfig.TLSConfig = tlsConfig
	}

	if len(cli.Config.Hosts) == 0 {
		cli.Config.Hosts = make([]string, 1)
	}

	cli.api = apiserver.New(serverConfig)

	var hosts []string

	// 设置API server
	for i := 0; i < len(cli.Config.Hosts); i++ {
		var err error
		if cli.Config.Hosts[i], err = dopts.ParseHost(cli.Config.TLS, cli.Config.Hosts[i]); err != nil {
			return fmt.Errorf("error parsing -H %s : %v", cli.Config.Hosts[i], err)
		}

		protoAddr := cli.Config.Hosts[i]
		protoAddrParts := strings.SplitN(protoAddr, "://", 2)
		if len(protoAddrParts) != 2 {
			return fmt.Errorf("bad format %s, expected PROTO://ADDR", protoAddr)
		}

		proto := protoAddrParts[0]
		addr := protoAddrParts[1]

		// It's a bad idea to bind to TCP without tlsverify.
		if proto == "tcp" && (serverConfig.TLSConfig == nil || serverConfig.TLSConfig.ClientAuth != tls.RequireAndVerifyClientCert) {
			logrus.Warn("[!] DON'T BIND ON ANY IP ADDRESS WITHOUT setting --tlsverify IF YOU DON'T KNOW WHAT YOU'RE DOING [!]")
		}
		ls, err := listeners.Init(proto, addr, serverConfig.SocketGroup, serverConfig.TLSConfig)
		if err != nil {
			return err
		}
		ls = wrapListeners(proto, ls)
		// If we're binding to a TCP port, make sure that a container doesn't try to use it.
		if proto == "tcp" {
			if err := allocateDaemonPort(addr); err != nil {
				return err
			}
		}
		logrus.Debugf("Listener created for HTTP on %s (%s)", proto, addr)
		hosts = append(hosts, protoAddrParts[1])
		cli.api.Accept(addr, ls...)
	}

	registryService, err := registry.NewService(cli.Config.ServiceOptions)
	if err != nil {
		return err
	}

	rOpts, err := cli.getRemoteOptions()
	if err != nil {
		return fmt.Errorf("Failed to generate containerd options: %s", err)
	}
	containerdRemote, err := libcontainerd.New(filepath.Join(cli.Config.Root, "containerd"), filepath.Join(cli.Config.ExecRoot, "containerd"), rOpts...)
	if err != nil {
		return err
	}
	signal.Trap(func() {
		cli.stop()
		<-stopc // wait for daemonCli.start() to return
	}, logrus.StandardLogger())

	// Notify that the API is active, but before daemon is set up.
	preNotifySystem()

	pluginStore := plugin.NewStore()

	// 初始化中间件
	if err := cli.initMiddlewares(cli.api, serverConfig, pluginStore); err != nil {
		logrus.Fatalf("Error creating middlewares: %v", err)
	}

	// 实例化daemon
	d, err := daemon.NewDaemon(cli.Config, registryService, containerdRemote, pluginStore)
	if err != nil {
		return fmt.Errorf("Error starting daemon: %v", err)
	}

	d.StoreHosts(hosts)

	// validate after NewDaemon has restored enabled plugins. Dont change order.
	if err := validateAuthzPlugins(cli.Config.AuthorizationPlugins, pluginStore); err != nil {
		return fmt.Errorf("Error validating authorization plugin: %v", err)
	}

	// TODO: move into startMetricsServer()
	if cli.Config.MetricsAddress != "" {
		if !d.HasExperimental() {
			return fmt.Errorf("metrics-addr is only supported when experimental is enabled")
		}
		if err := startMetricsServer(cli.Config.MetricsAddress); err != nil {
			return err
		}
	}

	// TODO: createAndStartCluster()
	name, _ := os.Hostname()

	// Use a buffered channel to pass changes from store watch API to daemon
	// A buffer allows store watch API and daemon processing to not wait for each other
	watchStream := make(chan *swarmapi.WatchMessage, 32)

	// 默认就初始化Docker Swarm？
	c, err := cluster.New(cluster.Config{
		Root:                   cli.Config.Root,
		Name:                   name,
		Backend:                d,
		ImageBackend:           d.ImageService(),
		PluginBackend:          d.PluginManager(),
		NetworkSubnetsProvider: d,
		DefaultAdvertiseAddr:   cli.Config.SwarmDefaultAdvertiseAddr,
		RaftHeartbeatTick:      cli.Config.SwarmRaftHeartbeatTick,
		RaftElectionTick:       cli.Config.SwarmRaftElectionTick,
		RuntimeRoot:            cli.getSwarmRunRoot(),
		WatchStream:            watchStream,
	})
	if err != nil {
		logrus.Fatalf("Error creating cluster component: %v", err)
	}
	d.SetCluster(c)
	err = c.Start()
	if err != nil {
		logrus.Fatalf("Error starting cluster component: %v", err)
	}

	// Restart all autostart containers which has a swarm endpoint
	// and is not yet running now that we have successfully
	// initialized the cluster.
	d.RestartSwarmContainers()

	logrus.Info("Daemon has completed initialization")

	cli.d = d

	// 注册handler到router
	routerOptions, err := newRouterOptions(cli.Config, d)
	if err != nil {
		return err
	}
	routerOptions.api = cli.api
	routerOptions.cluster = c

	// 初始化router
	initRouter(routerOptions)

	// process cluster change notifications
	watchCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go d.ProcessClusterNotifications(watchCtx, watchStream)

	cli.setupConfigReloadTrap()

	// The serve API routine never exits unless an error occurs
	// We need to start it as a goroutine and wait on it so
	// daemon doesn't exit
	serveAPIWait := make(chan error)
	// 开始执行
	go cli.api.Wait(serveAPIWait)

	// after the daemon is done setting up we can notify systemd api
	// 还要通知systemd？
	notifySystem()

	// Daemon is fully initialized and handling API traffic
	// Wait for serve API to complete
	// api server停了，daemon就跟着退出
	errAPI := <-serveAPIWait
	c.Cleanup()
	shutdownDaemon(d)
	containerdRemote.Cleanup()
	if errAPI != nil {
		return fmt.Errorf("Shutting down due to ServeAPI error: %v", errAPI)
	}

	return nil
}
```

这里做的事情可就多了，加载配置，设置相关的变量和配置，监听端口，设置信号handler，
起了API server，等待接受请求。

## 创建容器

上面我们看到了daemon的启动过程。启动完之后就等待请求了。那我们要找一个新的入口点去跟踪
代码，所以我选择 `docker run`。从 `docker/cli` 库翻了翻，发现最后是调用 `containers/create`
这样一个接口。然后就在本仓库里搜索，我们的目标是找到处理这个请求的mux所在地，然后翻出对应的
handler `api/server/router/container/container.go`，然后跳到 `api/server/router/container/container_routes.go`:

```go
ccr, err := s.backend.ContainerCreate(types.ContainerCreateCo
nfig{
        Name:             name,
        Config:           config,
        HostConfig:       hostConfig,
        NetworkingConfig: networkingConfig,
        AdjustCPUShares:  adjustCPUShares,
})
```

跳进去，发现是个接口。而 `s.backend` 其实就是daemon。具体代码在上一节就可以跟到。所以我打开fzf进行
泛搜索 `func daemon ContainerCreate`，找到了 `daemon/create.go` ：

```go
// ContainerCreate creates a regular container
// 创建一个容器
func (daemon *Daemon) ContainerCreate(params types.ContainerCreateConfig) (containertypes.ContainerCreateCreatedBody, error) {
	return daemon.containerCreate(params, false)
}

func (daemon *Daemon) containerCreate(params types.ContainerCreateConfig, managed bool) (containertypes.ContainerCreateCreatedBody, error) {
	start := time.Now()
	if params.Config == nil {
		return containertypes.ContainerCreateCreatedBody{}, errdefs.InvalidParameter(errors.New("Config cannot be empty in order to create a container"))
	}

	os := runtime.GOOS // 不知道这个是干啥的。。。Windows和Linux怎么混来混去的。。。
	if params.Config.Image != "" {
		// 拉取镜像
		img, err := daemon.imageService.GetImage(params.Config.Image)
		if err == nil {
			os = img.OS
		}
	} else {
		// This mean scratch. On Windows, we can safely assume that this is a linux
		// container. On other platforms, it's the host OS (which it already is)
		if runtime.GOOS == "windows" && system.LCOWSupported() {
			os = "linux"
		}
	}

	// 验证容器的配置，如果有问题，就不能创建
	warnings, err := daemon.verifyContainerSettings(os, params.HostConfig, params.Config, false)
	if err != nil {
		return containertypes.ContainerCreateCreatedBody{Warnings: warnings}, errdefs.InvalidParameter(err)
	}

	// 验证网络配置
	err = verifyNetworkingConfig(params.NetworkingConfig)
	if err != nil {
		return containertypes.ContainerCreateCreatedBody{Warnings: warnings}, errdefs.InvalidParameter(err)
	}

	if params.HostConfig == nil {
		params.HostConfig = &containertypes.HostConfig{}
	}
	// 调整一些配置，例如CPU如果超量了，就设置成系统允许的最大的。等。
	err = daemon.adaptContainerSettings(params.HostConfig, params.AdjustCPUShares)
	if err != nil {
		return containertypes.ContainerCreateCreatedBody{Warnings: warnings}, errdefs.InvalidParameter(err)
	}

	// 创建容器
	container, err := daemon.create(params, managed)
	if err != nil {
		return containertypes.ContainerCreateCreatedBody{Warnings: warnings}, err
	}
	containerActions.WithValues("create").UpdateSince(start)

	return containertypes.ContainerCreateCreatedBody{ID: container.ID, Warnings: warnings}, nil
}

// Create creates a new container from the given configuration with a given name.
// 创建容器
func (daemon *Daemon) create(params types.ContainerCreateConfig, managed bool) (retC *container.Container, retErr error) {
	var (
		container *container.Container
		img       *image.Image
		imgID     image.ID
		err       error
	)

	os := runtime.GOOS
	if params.Config.Image != "" {
		img, err = daemon.imageService.GetImage(params.Config.Image)
		if err != nil {
			return nil, err
		}
		if img.OS != "" {
			os = img.OS
		} else {
			// default to the host OS except on Windows with LCOW
			if runtime.GOOS == "windows" && system.LCOWSupported() {
				os = "linux"
			}
		}
		imgID = img.ID()

		if runtime.GOOS == "windows" && img.OS == "linux" && !system.LCOWSupported() {
			return nil, errors.New("operating system on which parent image was created is not Windows")
		}
	} else {
		// 没搞懂这个分支在这干啥。。。
		if runtime.GOOS == "windows" {
			os = "linux" // 'scratch' case.
		}
	}

	// 再次检查配置
	if err := daemon.mergeAndVerifyConfig(params.Config, img); err != nil {
		return nil, errdefs.InvalidParameter(err)
	}

	if err := daemon.mergeAndVerifyLogConfig(&params.HostConfig.LogConfig); err != nil {
		return nil, errdefs.InvalidParameter(err)
	}

	// 创建容器。。。。这个嵌套的有点多。。。此处返回的是内存中对容器的一个抽象 `Container`
	if container, err = daemon.newContainer(params.Name, os, params.Config, params.HostConfig, imgID, managed); err != nil {
		return nil, err
	}
	defer func() {
		if retErr != nil {
			if err := daemon.cleanupContainer(container, true, true); err != nil {
				logrus.Errorf("failed to cleanup container on create error: %v", err)
			}
		}
	}()

	if err := daemon.setSecurityOptions(container, params.HostConfig); err != nil {
		return nil, err
	}

	container.HostConfig.StorageOpt = params.HostConfig.StorageOpt

	// Fixes: https://github.com/moby/moby/issues/34074 and
	// https://github.com/docker/for-win/issues/999.
	// Merge the daemon's storage options if they aren't already present. We only
	// do this on Windows as there's no effective sandbox size limit other than
	// physical on Linux.
	if runtime.GOOS == "windows" {
		if container.HostConfig.StorageOpt == nil {
			container.HostConfig.StorageOpt = make(map[string]string)
		}
		for _, v := range daemon.configStore.GraphOptions {
			opt := strings.SplitN(v, "=", 2)
			if _, ok := container.HostConfig.StorageOpt[opt[0]]; !ok {
				container.HostConfig.StorageOpt[opt[0]] = opt[1]
			}
		}
	}

	// Set RWLayer for container after mount labels have been set
	// 要创建一个RWLayer，这样容器里面才可以读写。前面的layer都包含在image里。
	rwLayer, err := daemon.imageService.CreateLayer(container, setupInitLayer(daemon.idMappings))
	if err != nil {
		return nil, errdefs.System(err)
	}
	container.RWLayer = rwLayer

	rootIDs := daemon.idMappings.RootPair()
	if err := idtools.MkdirAndChown(container.Root, 0700, rootIDs); err != nil {
		return nil, err
	}
	if err := idtools.MkdirAndChown(container.CheckpointDir(), 0700, rootIDs); err != nil {
		return nil, err
	}

	if err := daemon.setHostConfig(container, params.HostConfig); err != nil {
		return nil, err
	}

	if err := daemon.createContainerOSSpecificSettings(container, params.Config, params.HostConfig); err != nil {
		return nil, err
	}

	var endpointsConfigs map[string]*networktypes.EndpointSettings
	if params.NetworkingConfig != nil {
		endpointsConfigs = params.NetworkingConfig.EndpointsConfig
	}
	// Make sure NetworkMode has an acceptable value. We do this to ensure
	// backwards API compatibility.
	runconfig.SetDefaultNetModeIfBlank(container.HostConfig)

	daemon.updateContainerNetworkSettings(container, endpointsConfigs)
	if err := daemon.Register(container); err != nil {
		return nil, err
	}
	stateCtr.set(container.ID, "stopped")
	daemon.LogContainerEvent(container, "create")
	return container, nil
}
```

然后跟到 `daemon.create`，仔细看，然后继续跟到了 `daemon.newContainer` ：

```go
func (daemon *Daemon) newContainer(name string, operatingSystem string, config *containertypes.Config, hostConfig *containertypes.HostConfig, imgID image.ID, managed bool) (*container.Container, error) {
	var (
		id             string
		err            error
		noExplicitName = name == ""
	)
	id, name, err = daemon.generateIDAndName(name)
	if err != nil {
		return nil, err
	}

	if hostConfig.NetworkMode.IsHost() {
		if config.Hostname == "" {
			config.Hostname, err = os.Hostname()
			if err != nil {
				return nil, errdefs.System(err)
			}
		}
	} else {
		// 原来容器的hostname默认是容器id的前12位
		daemon.generateHostname(id, config)
	}
	entrypoint, args := daemon.getEntrypointAndArgs(config.Entrypoint, config.Cmd)

	base := daemon.newBaseContainer(id) // 瞧。。。又一个创建容器
	base.Created = time.Now().UTC()
	base.Managed = managed
	base.Path = entrypoint
	base.Args = args //FIXME: de-duplicate from config
	base.Config = config
	base.HostConfig = &containertypes.HostConfig{}
	base.ImageID = imgID
	base.NetworkSettings = &network.Settings{IsAnonymousEndpoint: noExplicitName}
	base.Name = name
	base.Driver = daemon.imageService.GraphDriverForOS(operatingSystem)
	base.OS = operatingSystem
	return base, err
}
```

整个流程就跟完了。
