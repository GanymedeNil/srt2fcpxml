# srt2fcpxml

srt 字幕文件转为final cut pro 字幕文件(fcpxml)

本软件使用的 final cut pro 10.4.6 版本的 fcpxml 文件作为模版开发，如果又问题请升级到对应版本


## 编译

在项目目录下执行`make`命令后在`build`目录下生成`srt2fcpxml`执行文件。

## 下载

不想编译的用户可以直接下载[执行文件](https://github.com/GanymedeNil/srt2fcpxml/releases)。

## 使用

```bash
$ ./srt2fcpxml
  -fd int
    	帧率目前只支持整数24、25、30、50、60 (default 25)
  -srt string
    	srt 字幕文件
```

执行

```bash
$ ./srt2fcpxml -srt /tmp/test.srt
```

在 srt 文件的目录中会自动生成以srt文件名命名的`fcpxml`文件。
