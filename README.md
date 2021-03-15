# srt2fcpxml
Convert srt subtitle file to final cut pro subtitle file(fcpxml)

This software uses final cut pro X 10.4.6 version fcpxml file as template development, if there is any problem, please upgrade to the corresponding version.

srt 字幕文件转为final cut pro 字幕文件(fcpxml)

本软件使用 final cut pro X 10.4.6 版本的 fcpxml 文件作为模版开发，如果有问题请升级到对应版本


## Compile (编译)
First, you need to have Go language development environment
Then execute `make` command in the project directory and generate `srt2fcpxml` executable file in `build` directory.

首先需要有 Go 语言开发环境
然后在项目目录下执行`make`命令后在`build`目录下生成`srt2fcpxml`执行文件。

## Download (下载)
Users who do not want to compile can download the [executable file](https://github.com/GanymedeNil/srt2fcpxml/releases) directly.

不想编译的用户可以直接下载[执行文件](https://github.com/GanymedeNil/srt2fcpxml/releases)。

## Use (使用)
First you need to give the program execute permission `chmod +x . /srt2fcpxml`

首先需要赋予程序执行权限 `chmod +x ./srt2fcpxml`

```bash
$ ./srt2fcpxml
  -fd int
    	Frame rate is currently supported (帧率目前支持) 23.98、24、25、29.97、30、50、59.94、60 (default 25)
  -srt string
    	srt Subtitle files (字幕文件)
```

## Execution (执行)

```bash
$ ./srt2fcpxml -srt /tmp/test.srt
```
he `fcpxml` file named with srt file name will be generated automatically in the directory of srt file.

在 srt 文件的目录中会自动生成以srt文件名命名的`fcpxml`文件。
