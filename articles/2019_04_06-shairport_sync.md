# 使用shairport-sync搭建airplay音频服务器

Arch下直接：

```bash
$ sudo pacmna -S shairport-sync
$ sudo systemctl enable shairport-sync
```

如果直接启动，会发现没有声音。因为默认的配置文件里，有这样一行：

```
alsa =
{
	output_device = "default"; // the name of the alsa output device. Use "alsamixer" or "aplay" to find out the names of devices, mixers, etc.
	mixer_control_name = "default"; // the name of the mixer to use to adjust output volume. If not specified, volume in adjusted in software.
	mixer_device = "default"; // the mixer_device default is whatever the output_device is. Normally you wouldn't have to use this.
    ...
}
```

修改成 `hw:0`。具体是什么要用 `aplay -l` 查看，不过一般都是这个值。

```
alsa =
{
	output_device = "hw:0"; // the name of the alsa output device. Use "alsamixer" or "aplay" to find out the names of devices, mixers, etc.
	mixer_control_name = "PCM"; // the name of the mixer to use to adjust output volume. If not specified, volume in adjusted in software.
	mixer_device = "hw:0"; // the mixer_device default is whatever the output_device is. Normally you wouldn't have to use this.
```

然后启动服务，打开iOS设备，就可以在手机上点歌，笔记本上放歌了 :)

---

不过我重启之后，使用 `hw:0` 又不行了，改回 `default` 又可以了，奇怪。
