# srt2fcpxml

Convert subtitle files (SRT, WebVTT, SSA/ASS, TTML, STL) into a Final Cut Pro project (`.fcpxml`) where every subtitle is a **Basic Title** clip — and export the titles on a Final Cut Pro timeline back to SRT with `fcpxml2srt`.

把字幕文件（SRT、WebVTT、SSA/ASS、TTML、STL）转换为 Final Cut Pro 工程（`.fcpxml`），每条字幕都是一个**基本字幕（Basic Title）**片段；也可以用 `fcpxml2srt` 把 Final Cut Pro 时间线上的标题导出为 SRT。

> Final Cut Pro can import SRT files itself, but only as *captions*, whose styling options are limited, and it cannot export titles as SRT. srt2fcpxml makes *titles*, so you can change their font, color, outline, position and effects like any other title.
>
> Final Cut Pro 自带的 SRT 导入生成的是**隐藏式字幕（caption）**，样式选项很少，也不能把标题导出为 SRT。srt2fcpxml 生成的是**字幕标题（title）**，可以像普通标题一样修改字体、颜色、描边、位置和效果。

**Web version (网页版)**: <https://ganymedenil.github.io/srt2fcpxml/> — runs in your browser; files are never uploaded. 在浏览器中运行，文件不会上传。

## Install (安装)

**Download (下载)**: get the archive for your system from [Releases](https://github.com/GanymedeNil/srt2fcpxml/releases). It contains both `srt2fcpxml` and `fcpxml2srt`; the macOS build runs on both Apple silicon and Intel.

从 [Releases](https://github.com/GanymedeNil/srt2fcpxml/releases) 下载对应系统的压缩包，里面包含 `srt2fcpxml` 和 `fcpxml2srt`。macOS 版本同时支持 Apple 芯片和 Intel。

macOS marks downloaded files as quarantined. If it refuses to run them, remove the flag once:

macOS 会隔离下载的文件。如果提示无法打开，执行一次：

```bash
xattr -d com.apple.quarantine ./srt2fcpxml ./fcpxml2srt
```

**With Go (使用 Go 安装)**:

```bash
go install github.com/GanymedeNil/srt2fcpxml/v3/cmd/srt2fcpxml@latest
```

```bash
go install github.com/GanymedeNil/srt2fcpxml/v3/cmd/fcpxml2srt@latest
```

**From source (从源码编译)**: run `make` in the project directory; the binaries are written to `build/`.

在项目目录执行 `make`，程序生成在 `build/` 中。

## Subtitles → Final Cut Pro (字幕 → Final Cut Pro)

```bash
srt2fcpxml -fd 23.976 talk.srt
```

This writes `talk.fcpxml` next to `talk.srt`. In Final Cut Pro choose **File → Import → XML…** and pick it.

会在 `talk.srt` 旁边生成 `talk.fcpxml`，在 Final Cut Pro 中选择**文件 → 导入 → XML…** 导入即可。

**Set the frame rate to match your video (帧率必须与视频一致)**, or let `-video` read it (below). Subtitles are snapped to whole frames of this rate.

```text
Usage: srt2fcpxml [flags] <subtitle file>

  -fd string                    frame rate: 23.976 (23.98), 24, 25, 29.97, 30, 50, 59.94, 60 (default "25")
  -width int                    video width in pixels (default 1920)
  -height int                   video height in pixels (default 1080)
  -video string                 video file to put under the subtitles; reads -fd, -width and -height from it (needs ffprobe)
  -o string                     output file, or - for stdout (default: next to the subtitle file)
  -secondary string             second-language subtitle file, shown under the matching subtitle
  -split-lines                  bilingual file: first line of each subtitle is the main language, the rest the second language
  -style string                 JSON style file; the -font*, -secondary-font*, -stroke-* and -position-y flags override it
  -font string                  font family (default "PingFang SC")
  -font-face string             font face, e.g. Regular, Bold (default "Semibold")
  -font-size float              font size (default 52)
  -font-color string            text color: "#RRGGBB", "#RRGGBBAA" or "r g b a" (0-1) (default "#FFFFFF")
  -stroke-color string          outline color (needs -stroke-width)
  -stroke-width float           outline width
  -position-y float             vertical position from the center of the frame; negative is lower (default -450)
  -secondary-font string        font family of the second language (default "PingFang SC")
  -secondary-font-size float    font size of the second language (default 36)
  -secondary-font-color string  text color of the second language (default "#FFFFFF")
  -srt string                   subtitle file (can also be given as the argument)
  -version                      print the version and exit
```

Examples (示例):

```bash
# 4K vertical video at 29.97 fps, written to a chosen file
srt2fcpxml -fd 29.97 -width 2160 -height 3840 -o ~/Desktop/subs.fcpxml talk.srt

# Yellow text with a black outline, a little higher than default
srt2fcpxml -font-color "#FFD700" -stroke-color "#000000" -stroke-width 3 -position-y -380 talk.srt
```

### With the video (带视频生成时间线)

```bash
srt2fcpxml -video talk.mov talk.srt
```

The project puts `talk.mov` in the primary storyline with the subtitles connected to it, so they line up with the picture as soon as you open the project. The frame rate, size (portrait phone videos included) and start timecode come from the video, read with `ffprobe` from [FFmpeg](https://ffmpeg.org) (`brew install ffmpeg`). `-fd`, `-width` and `-height` still override them.

工程的主故事情节是 `talk.mov`，字幕连接在它上面，打开工程即与画面对齐。帧率、分辨率（包括手机竖屏视频）和起始时间码都从视频读取，需要 [FFmpeg](https://ffmpeg.org) 的 `ffprobe`（`brew install ffmpeg`）；`-fd`、`-width`、`-height` 仍可覆盖。

### Bilingual subtitles (双语字幕)

```bash
# Two files: each English subtitle goes under the Chinese subtitle it overlaps most
srt2fcpxml -secondary talk.en.srt talk.zh.srt

# One file: the first line of each subtitle is the main language, the other lines the second language
srt2fcpxml -split-lines talk.bilingual.srt
```

The second language uses its own style (default: same font, Regular, size 36). A second-language subtitle that overlaps no main subtitle becomes a subtitle of its own.

第二语言使用单独的样式（默认同字体、Regular、字号 36）。与任何主字幕都不重叠的第二语言字幕会单独成为一条。

Inline bold, italic and SRT `<font color>` are kept as Final Cut Pro text styles. 行内的粗体、斜体和 SRT 的 `<font color>` 会保留为 Final Cut Pro 文字样式。

### Style file (样式文件)

Save a style you like as JSON and reuse it with `-style mystyle.json`. Fields you leave out keep their default; unknown fields are reported as errors.

把常用样式保存为 JSON 文件，用 `-style mystyle.json` 复用。未写的字段使用默认值，拼错的字段会报错。

```json
{
  "font": "PingFang SC",
  "fontFace": "Semibold",
  "fontSize": 52,
  "fontColor": "#FFFFFF",
  "bold": true,
  "italic": false,
  "strokeColor": "#000000CC",
  "strokeWidth": 3,
  "shadowColor": "#000000BF",
  "shadowDistance": 5,
  "shadowAngle": 315,
  "shadowBlurRadius": 0,
  "lineSpacing": 0,
  "alignment": "center",
  "positionX": 0,
  "positionY": -450,
  "secondary": {
    "font": "PingFang SC",
    "fontFace": "Regular",
    "fontSize": 36,
    "fontColor": "#FFD700"
  }
}
```

| Field | Meaning (含义) |
|---|---|
| `font`, `fontFace`, `fontSize` | Font family, face and size (字体、字重、字号) |
| `fontColor`, `strokeColor`, `shadowColor` | `"#RRGGBB"`, `"#RRGGBBAA"` or `"r g b a"` (0–1). `""` leaves it to the template (留空则使用模板默认值) |
| `bold`, `italic` | `true` / `false` |
| `strokeWidth` | Outline width; required when `strokeColor` is set (描边宽度) |
| `shadowDistance`, `shadowAngle`, `shadowBlurRadius` | Drop shadow; angle in degrees (阴影距离、角度、模糊) |
| `lineSpacing` | Extra space between lines (行距) |
| `alignment` | `left`, `center` or `right` (对齐) |
| `positionX`, `positionY` | Position relative to the frame center (相对画面中心的位置) |
| `secondary` | The second language's font, face, size, colors, bold, italic, stroke and shadow, with the same field names (第二语言的文字样式，字段名相同) |

Sizes and positions use the coordinates of the Basic Title template (1920×1080); Final Cut Pro scales the template to your project's resolution.

字号与位置使用基本字幕模板的坐标（1920×1080），Final Cut Pro 会按工程分辨率缩放模板。

## Final Cut Pro → subtitles (Final Cut Pro → 字幕)

In Final Cut Pro choose **File → Export XML…**, then:

在 Final Cut Pro 中选择**文件 → 导出 XML…**，然后：

```bash
fcpxml2srt talk.fcpxmld
```

This writes `talk.srt` with the titles and captions on the project's timeline, in time order. Titles inside compound clips, secondary storylines and connected clips are included; disabled clips are skipped.

会生成 `talk.srt`，包含工程时间线上的标题和字幕，按时间排序。复合片段、次级故事情节和连接片段中的标题都会导出，停用的片段会跳过。

```text
Usage: fcpxml2srt [flags] <project.fcpxml | project.fcpxmld>

  -o string        output file: .srt or .vtt, or - for SRT on stdout (default: next to the input, with an .srt extension)
  -project string  project to export, when the file holds more than one
  -only string     export only titles or only captions: title, caption
  -version         print the version and exit
```

Titles on retimed (speed-changed) clips are reported with a warning, since their times may be off. 变速片段上的标题会给出警告，其时间可能不准。

## Compatibility (兼容性)

- Writes FCPXML 1.7, which Final Cut Pro 10.4.6 and later can import. Output is checked against the FCPXML 1.7 DTD shipped with Final Cut Pro, and against 1.14 too except for projects with a video (FCPXML 1.9 changed how media files are referenced).
- 输出 FCPXML 1.7，Final Cut Pro 10.4.6 及以上版本可导入；输出通过了 Final Cut Pro 自带的 FCPXML 1.7 DTD 校验，不带视频的工程也通过 1.14 校验（FCPXML 1.9 改变了媒体文件的引用方式）。
- `fcpxml2srt` reads every FCPXML 1.x version and `.fcpxmld` bundles. `fcpxml2srt` 可读取所有 FCPXML 1.x 版本和 `.fcpxmld` 包。

### Changes in v3 (v3 的变化)

- v3.1: bilingual subtitles, inline styles, `-video`, `fcpxml2srt` and the web version. Spaces between inline tags are kept (v3.0 turned `<i>a</i> <b>b</b>` into "ab").
- v3.0: timing is exact at every supported frame rate (v2 drifted about 0.6 s per hour at 23.98 fps). Invalid input is reported with exit code 1; running without an input file prints the usage and exits with code 2 (v2 used 20). New `-o`, `-style`, style flags, `-version`, and WebVTT/ASS/TTML/STL input.
- v3.1：双语字幕、行内样式、`-video`、`fcpxml2srt` 和网页版；保留行内标签之间的空格（v3.0 会把 `<i>a</i> <b>b</b>` 变成 "ab"）。
- v3.0：所有支持的帧率下时间都精确（v2 在 23.98 帧下每小时偏移约 0.6 秒）；非法输入以退出码 1 报错；未指定输入文件时打印用法并以退出码 2 退出（v2 为 20）；新增 `-o`、`-style` 和样式参数、`-version`，以及 WebVTT/ASS/TTML/STL 输入。

## Development (开发)

```bash
make test
```

Golden files in `internal/convert/testdata` hold the expected output. After an intended change to the output, regenerate them and review the diff:

`internal/convert/testdata` 中的 golden 文件是期望输出；有意修改输出后重新生成并检查差异：

```bash
go test ./internal/convert -update
```

On a Mac with Final Cut Pro and `xmllint` installed, the tests also validate the golden files against Final Cut Pro's DTDs. Tests that need FFmpeg are skipped without it.

在装有 Final Cut Pro 和 `xmllint` 的 Mac 上，测试还会用 Final Cut Pro 的 DTD 校验 golden 文件；没有 FFmpeg 时跳过需要它的测试。

The web version is built with `make web` into `web/dist` (serve that folder to try it). It is published to GitHub Pages from `master`.

网页版用 `make web` 构建到 `web/dist`（用任意静态服务器打开该目录即可试用），并从 `master` 发布到 GitHub Pages。

Releases are built by GoReleaser when a `v*` tag is pushed.

推送 `v*` 标签时由 GoReleaser 自动构建发布。
