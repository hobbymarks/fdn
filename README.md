# UFn

## Introduction_简介
基于[rename](http://plasmasturm.org/code/rename/)的一个小工具，用于日常统一更改资料的文件名。目前的具体格式：
1. 文件名不保留空格（首先空格会被替换为下划线，之后根据是否存在连续下划线来决定缩减）；
2. 文件名中只保留下划线字符，其余的控制类字符会被替换为下划线；
3. 多个连续的下划线字符会被缩减为一个下划线；

## Install_安装

* 内容为[bash](https://www.gnu.org/software/bash/)脚本，可以直接下载和执行;
* 将UFn.sh放置在/usr/local/bin 或者/usr/bin 或者其它[环境变量](https://en.wikipedia.org/wiki/Environment_variable)包含的目录，这样可以在任意目录执行该脚本；
* 如果需要其它目录执行可以考虑创建[软连接](https://en.wikipedia.org/wiki/Ln_(Unix))(ln -s UFn.sh TargetName)
