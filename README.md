# localtags 本地文件备份工具


## 重要更新

- 2021-11-23 (improve: 为了方便与其他本地程序互动，改成了允许跨域访问)
- 2021-10-28 (improve: 改进了文件校验及修复了相关bug) [changelog.md](./docs/changelog.md)
- 2021-05-25 (add: 标签预览) [Wiki: Tag Preview](https://github.com/ahui2016/localtags/wiki/Tag-Preview-(update:-2021-05-25))
- 2021-05-18 (add: 替换文件) [Wiki: 同名文件(例子三)](https://github.com/ahui2016/localtags/wiki/Same-Name-Files)
- 2021-05-14 (add: 快速创建 markdown 笔记) [changelog.md](./docs/changelog.md)
- 2021-05-12 (add: 通过网页修改配置) 详见本页后文 "端口等的设置"
- 2021-05-10 (add: 单个文件体积上限) [details.md](./docs/details.md)
- 2021-05-09 (add: 本地 markdown 图库) [details.md](./docs/details.md)
- 2021-05-08 (fix: 备份仓库文件校验) [changelog.md](./docs/changelog.md)


## 主要功能

1. 文件备份
2. 定期查错（确保文件完整性）
3. 标签管理
4. 文件历史版本
5. 本地 markdown 图库


## 截图

本软件的界面截图请看 screenshots 文件夹。


## 安装运行

### 直接下载可执行文件

由于 Windows 编译有点麻烦 (涉及 cgo, 需要 gcc), 因此我做了 Windows 的可执行文件方便大家试用。Mac 和 Linux 编译应该很方便，就不提供二进制文件了。

> 下载地址 => https://github.com/ahui2016/localtags/releases
>
> (注意: 需要另外安装 ffmpeg 才能给视频文件生成缩略图)

### 手动编译

- 要求先正确安装 git 和 [Go 语言](https://golang.google.cn/)、ffmpeg。
- 其中，ffmpeg 不是必须的，其作用只是给视频文件生成缩略图，不安装 ffmpeg 也不妨碍软件正常运行。
- 由于采用了 go-sqlite3, 因此如果在 Windows 里编译, 需要先安装 [TDM-GCC Compiler](https://sourceforge.net/projects/tdm-gcc/)

```
$ cd ~
$ git clone https://github.com/ahui2016/localtags.git
$ cd localtags
$ go build
$ ./localtags
```

然后用浏览器访问 `http://127.0.0.1:53549` 即可。本软件采用了网页来做 GUI 界面，但它本质上是一个本地软件，不是网站。


## 端口等的设置

- 第一次运行程序后, 会自动生成 config.json 文件, 端口及其他设置可直接修改该文件, 保存后重启程序生效.
- (2021-05-12更新) 可以在程序首页点击 "Config" 进入修改配置页面，在这里修改配置比直接编辑 config.json 文件更方便（但还是需要手动重启程序）。
- 每个设置项目的具体含义详见 config.go 文件.
- 启动程序时默认在当前目录寻找 config.json, 也可以手动指定 `$ ./localtags.exe -config /path/to/config.json`


## 前端采用 MJ.js

mj.js 只是在 jquery 的基础上增加了两个函数，因此对于会用 jquery 的人来说，学习成本接近零。但与一般使用 jquery 的方式不同, mj.js 完全不写 HTML, 一切都是 js, 因此非常轻松实现组件化，并且实际效果非常好，组件可以相互交流、可以嵌套、可以复用。

关于 mj.js 的更多信息请看 https://github.com/ahui2016/mj.js


## 反直觉设计

本软件有些地方会不符合直觉，不符合主流的使用习惯。

### 看起来像个网站，但操作起来不像网站

比如，虽然用网页来做 GUI 界面，但不能通过网页上传文件，而是让用户把想要上传的文件放进一个本地文件夹中，然后手动刷新网页获取文件列表。

这样做是为了减少代码量，说白了就是偷懒，可以少写很多代码，而减少代码有利于减少 bug, 维护起来也轻松一些, 用户审查代码或个性化修改代码时也轻松一些.

### 需要使用网页控制台

有一些功能用户要按 F12 进入浏览器控制台，输入命令来操作。这样做有两个目的：1.让操作界面尽量简洁，按钮能少一个是一个，不常用的命令就让用户麻烦一点点了。2.还是为了偷懒，减少代码量。

### 独特的标签系统

另一个可能让用户不习惯的地方是**纯标签管理，完全没有文件夹**，要求用户给每个文件设置至少 2 个标签。这个设定是经过深思熟虑的，配合本软件特有的 “标签组” 功能，会让文件分类变得井井有条。


## 了解更多

更多细节及带截图的说明请看 Wiki https://github.com/ahui2016/localtags/wiki
