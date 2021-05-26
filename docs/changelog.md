# localtags 更新记录


## 2021-05-14 (add: 快速创建 markdown 笔记)

有时，我想新建一个 markdown 文件，简单写几句话，然后保存到 localtags 里，那么，操作步骤如下所示。

**在添加该功能之前**：打开 waiting 文件夹 --> 按鼠标右键， 选择新建文件，点击新建 txt 文件 --> 把后缀名修改为 `.md` --> 双击打开文件，写内容，关闭文件 --> 更改一个合适的文件名。

**在添加该功能之后**：点击 Add Files 页面右上角的 new note 按钮 --> 写内容 --> 点击保存按钮即可。

可见，该功能可带来很大的方便。

### 但该功能并不完善，有一些问题需要注意

- 自动获取最开头的内容作为文件名，因此，如果 waiting 文件夹里恰好有同名文件，会被覆盖。
- 自动获取的文件名有可能包含不允许使用的特殊字符，点击 Save 按钮时会提示失败，遇到这种情况请修改笔记的第一行内容。（例如，`:`, `/`, `\`, `|`, `"`, `?`, `>`, `<`, `|`, `*` 是不能使用的，而 `#`, `-`, `!` 则可以使用）
- 暂不提供预览功能，只适合用来写简短内容。
- 没有自动保存功能，需要手动点击 Save 按钮才能保存文件。
- 在 new note 页面点击 Save 按钮即可新建一个 markdown 文件，如果想用其它文本编辑器来编辑它，请先关闭 new note 页面（注意先按 Save 按钮保存内容），避免两边同时编辑同一个文件而产生互相覆盖的情况。


## 2021-05-08 (fix: 备份仓库文件校验)

- bug: 每次备份时，都校验备份仓库里的全部文件
- 修复后: 每次备份时，只校验备份仓库里超过一定时间未校验的文件

### 修复步骤

（**注意：如果全新安装，则不需要进行修复操作**）

1. 获取最新版本，重新编译，重启程序
   ```
   $ cd /path/to/localtags
   $ git pull
   $ go build
   $ ./localtags.exe
   ```
2. 手动打开 config.json 文件，把里面的 `CheckInterval` 后面的数值修改为 `2592000` (该数值可自由设定, 单位是秒, 2592000 是 30 天, 如果设为 86400 则是一天，建议设为 2592000)