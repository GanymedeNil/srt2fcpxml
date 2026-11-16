# srt2fcpxml

Convert subtitle files (SRT, WebVTT, SSA/ASS, TTML, STL) into a Final Cut Pro project (`.fcpxml`) where every subtitle is a **Basic Title** clip.

把字幕文件（SRT、WebVTT、SSA/ASS、TTML、STL）转换为 Final Cut Pro 工程（`.fcpxml`），每条字幕都是一个**基本字幕（Basic Title）**片段。

> Final Cut Pro can import SRT files itself, but only as *captions*, whose styling options are limited. srt2fcpxml makes *titles*, so you can change their font, color, outline, position and effects like any other title.
>
> Final Cut Pro 自带的 SRT 导入生成的是**隐藏式字幕（caption）**，样式选项很少。srt2fcpxml 生成的是**字幕标题（title）**，可以像普通标题一样修改字体、颜色、描边、位置和效果。

## Install (安装)

**Download (下载)**: get the archive for your system from [Releases](https://github.com/GanymedeNil/srt2fcpxml/releases). The macOS build runs on both Apple silicon and Intel.

从 [Releases](https://github.com/GanymedeNil/srt2fcpxml/releases) 下载对应系统的压缩包。macOS 版本同时支持 Apple 芯片和 Intel。

macOS marks downloaded files as quarantined. If it refuses to run the binary, remove the flag once:

macOS 会隔离下载的文件。如果提示无法打开，执行一次：

```bash
xattr -d com.apple.quarantine ./srt2fcpxml
```

**With Go (使用 Go 安装)**:

```bash
go install github.com/GanymedeNil/srt2fcpxml/v3/cmd/srt2fcpxml@latest
```

**From source (从源码编译)**: run `make` in the project directory; the binary is written to `build/srt2fcpxml`.

在项目目录执行 `make`，生成 `build/srt2fcpxml`。

## Use (使用)

```bash
srt2fcpxml -fd 23.976 talk.srt
```

This writes `talk.fcpxml` next to `talk.srt`. In Final Cut Pro choose **File → Import → XML…** and pick it.

会在 `talk.srt` 旁边生成 `talk.fcpxml`，在 Final Cut Pro 中选择**文件 → 导入 → XML…** 导入即可。

**Set the frame rate to match your video (帧率必须与视频一致).** Subtitles are snapped to whole frames of this rate.

```text
Usage: srt2fcpxml [flags] <subtitle file>

  -fd string            frame rate: 23.976 (23.98), 24, 25, 29.97, 30, 50, 59.94, 60 (default "25")
  -width int            video width in pixels (default 1920)
  -height int           video height in pixels (default 1080)
  -o string             output file, or - for stdout (default: next to the subtitle file)
  -style string         JSON style file; the -font*, -stroke-* and -position-y flags override it
  -font string          font family (default "PingFang SC")
  -font-face string     font face, e.g. Regular, Bold (default "Semibold")
  -font-size float      font size (default 52)
  -font-color string    text color: "#RRGGBB", "#RRGGBBAA" or "r g b a" (0-1) (default "#FFFFFF")
  -stroke-color string  outline color (needs -stroke-width)
  -stroke-width float   outline width
  -position-y float     vertical position from the center of the frame; negative is lower (default -450)
  -srt string           subtitle file (can also be given as the argument)
  -version              print the version and exit
```

Examples (示例):

```bash
# 4K vertical video at 29.97 fps, written to a chosen file
srt2fcpxml -fd 29.97 -width 2160 -height 3840 -o ~/Desktop/subs.fcpxml talk.srt

# Yellow text with a black outline, a little higher than default
srt2fcpxml -font-color "#FFD700" -stroke-color "#000000" -stroke-width 3 -position-y -380 talk.srt
```

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
  "positionY": -450
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

Sizes and positions use the coordinates of the Basic Title template (1920×1080); Final Cut Pro scales the template to your project's resolution.

字号与位置使用基本字幕模板的坐标（1920×1080），Final Cut Pro 会按工程分辨率缩放模板。

## Compatibility (兼容性)

- Writes FCPXML 1.7, which Final Cut Pro 10.4.6 and later can import. Output is checked against the FCPXML 1.7 and 1.14 DTDs shipped with Final Cut Pro.
- 输出 FCPXML 1.7，Final Cut Pro 10.4.6 及以上版本可导入；输出通过了 Final Cut Pro 自带的 FCPXML 1.7 与 1.14 DTD 校验。
- Inline tags such as `<i>` and `<b>` are removed from the text; the style applies to the whole subtitle.
- `<i>`、`<b>` 等行内标签会被去除，样式作用于整条字幕。

### Changes in v3 (v3 的变化)

- Timing is exact at every supported frame rate (v2 drifted about 0.6 s per hour at 23.98 fps).
- Invalid input (unknown frame rate, unreadable or empty file, bad resolution) is reported with exit code 1. Running without an input file prints the usage and exits with code 2 (v2 used 20).
- New: positional input file, `-o`, `-style` and the style flags, `-version`, and WebVTT/ASS/TTML/STL input.
- 所有支持的帧率下时间都精确（v2 在 23.98 帧下每小时偏移约 0.6 秒）；非法输入以退出码 1 报错；未指定输入文件时打印用法并以退出码 2 退出（v2 为 20）；新增直接传文件名、`-o`、`-style` 和样式参数、`-version`，以及 WebVTT/ASS/TTML/STL 输入。

## Development (开发)

```bash
make test
```

Golden files in `internal/convert/testdata` hold the expected output. After an intended change to the output, regenerate them and review the diff:

`internal/convert/testdata` 中的 golden 文件是期望输出；有意修改输出后重新生成并检查差异：

```bash
go test ./internal/convert -update
```

On a Mac with Final Cut Pro and `xmllint` installed, the tests also validate the golden files against Final Cut Pro's DTDs.

在装有 Final Cut Pro 和 `xmllint` 的 Mac 上，测试还会用 Final Cut Pro 的 DTD 校验 golden 文件。

Releases are built by GoReleaser when a `v*` tag is pushed.

推送 `v*` 标签时由 GoReleaser 自动构建发布。
