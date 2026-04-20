# miaosic CLI

`miaosic` 是一个音乐服务命令行工具，支持搜索、信息查询、URL 解析、歌词、下载、二维码登录以及音频标签读写。

## 基本用法

```bash
miaosic [全局参数] <命令> [参数] [flags]
```

## 全局参数

| 参数 | 缩写 | 说明 |
|---|---|---|
| `--session-file` | `-s` | 会话文件路径，用于保存/恢复登录状态 |

说明：`--json/-j` 不是全局参数，只在部分命令中提供。

## 命令概览

| 命令 | 说明 |
|---|---|
| `providers` | 列出所有 provider 及登录状态 |
| `search` | 按关键词搜索 |
| `info` | 查询媒体信息 |
| `url` | 获取可播放 URL |
| `quality` | 查看 provider 支持的音质 |
| `lyric` | 获取歌词并可导出到文件 |
| `download` | 下载媒体（可选写入标签） |
| `qrlogin` | 二维码登录流程 |
| `tag` | 读取/写入本地音频标签 |

## 详细命令

### providers

```bash
miaosic providers
```

### search

```bash
miaosic search <provider> <keyword> [flags]
```

flags:
- `-p, --page` 页码，默认 `1`
- `--page-size` 每页数量，默认 `10`
- `-j, --json` JSON 输出

示例：
```bash
miaosic search netease "周杰伦" -p 1 --page-size 5
miaosic search qq Jay -j
```

### info

```bash
miaosic info <provider> <uri> [flags]
```

flags:
- `-j, --json` JSON 输出

示例：
```bash
miaosic info netease 1827600686
miaosic info qq 004Z8Ihr0JIu5s -j
```

### url

```bash
miaosic url <provider> <uri> [flags]
```

flags:
- `--quality` 音质偏好，如 `128k/320k/flac/hq/sq`
- `-j, --json` JSON 输出

示例：
```bash
miaosic url netease 1827600686 --quality 320k
miaosic url qq 004Z8Ihr0JIu5s -j
```

### quality

```bash
miaosic quality <provider> [flags]
```

flags:
- `-j, --json` JSON 输出

示例：
```bash
miaosic quality netease
```

### lyric

```bash
miaosic lyric <provider> <uri> [flags]
```

flags:
- `-o, --output` 输出到指定文件
- `--save` 自动按歌曲信息命名保存
- `-j, --json` JSON 输出

示例：
```bash
miaosic lyric netease 1827600686
miaosic lyric netease 1827600686 --save
miaosic lyric qq 004Z8Ihr0JIu5s -o lyric.lrc
```

### download

```bash
miaosic download <provider> <uri> [flags]
```

flags:
- `--quality` 指定音质偏好
- `--filename` 指定输出文件名
- `--use-actual-ext` 当 `--filename` 后缀和实际下载到的音频后缀不一致时，自动替换为实际后缀
- `--metadata` 下载后写入标签（标题/艺人/专辑/歌词/封面）

示例：
```bash
miaosic download netease 1827600686
miaosic download qq 004Z8Ihr0JIu5s --quality 320k --filename song.mp3
miaosic download kugou 3e3f9e3a4b47125e4b4558ca0bb4264a --quality flac --filename song.flac --use-actual-ext
miaosic download netease 1827600686 --metadata
```

### qrlogin

获取二维码：
```bash
miaosic qrlogin getqrcode <provider>
```

验证登录：
```bash
miaosic qrlogin verify <provider> <key>
```

示例：
```bash
miaosic --session-file ~/.miaosic_session.json qrlogin getqrcode netease
miaosic --session-file ~/.miaosic_session.json qrlogin verify netease <key>
```

### tag

读取标签：
```bash
miaosic tag read <file> [--format plain|json]
```

写入标签：
```bash
miaosic tag write <file> [flags]
```

`tag write` flags:
- `--title`
- `--artist`
- `--album`
- `--lyrics`
- `--lyrics-lang` (默认 `eng`)
- `--cover` 封面图片路径
- `--cover-type` 封面类型（默认前封面）

示例：
```bash
miaosic tag read ./data/test.mp3
miaosic tag read ./data/test.m4a --format json

miaosic tag write ./data/test.flac --title "Title" --artist "Artist" --album "Album"
miaosic tag write ./data/test.mp3 --lyrics "hello world" --lyrics-lang eng
miaosic tag write ./data/test.m4a --cover ./data/cover.jpg --cover-type 3
```
