# fdn

## Introduction

A tool for uniformly change file or directory names and support rollback
these operations.

File or directory name format:

1.  no space in file name（firstly every space will be replaced by an
    underscore,then multiple consecutive underscores will be reduced to
    one);
2.  only underscore allowed in file name，all other control characters
    will be replaced by underscore;
3.  multiple consecutive underscores will be reduced to one;
4.  underscore at the beginning of file name will be deleted;
5.  underscore at the end of file name will be deleted;
6.  keep [bash special
    parameters](https://www.gnu.org/software/bash/manual/html_node/Special-Parameters.html)
    in file name;
7.  some terminology will remained(update continuously) ,such as
    `USB`,`PCIe` ...​

## Installation

To install fdn via
[cargo](https://doc.rust-lang.org/cargo/getting-started/installation.html):

```bash
$ cargo install fdn
```

## Usage/参数

```bash
Usage:
   fdn [OPTIONS] [COMMAND]
Commands:
   config Config pattern help Print this message or the help of the given subcommand(s)
Options:
   -f, --file-path <FILE_PATH> file path [default: .]
   -i, --in-place in place
   -d, --max-depth <MAX_DEPTH> max depth [default: 1]
   -t, --filetype <FILETYPE> file type,'f' for regular file and 'd' for directory [default: f]
   -I, --not-ignore-hidden not ignore hidden file
   -X, --exclude-path <EXCLUDE_PATH> exclude file or directory
   -r, --reverse reverse change
   -a, --align align origin and edited
   -V, --version print version
   -h, --help Print help
Use "fdn [command] --help" for more information about a command.
```

## **ATTENTION!**

When you run fdn ,you will see two kinds of output:

- First Kind:

  sample▯file▯name
  -->sample_file_name

The output means: a file name `sample file name` will be changed to
`sample_file_name`
`-->` means in dry run mode ,operation not take effect.The character `▯`
means space ,every space will be replaced by one `▯`.
**The character \`\`▯\`\` is only for the convenience of visual contrast
and only display in output.**
or

- Second Kind:

  sample▯file▯name
  ==>sample_file_name

The output means: a file named `sample file name` has been changed to
`sample_file_name`
`==>` means operation have taken effect.
all deleted character will be display as red color ,such as the original
file name:
**sample ▯ file ▯ name**
all added character will be diplayed as green color ,such as the changed
file name:
**sample \* file \* name**

### Options

-d option

$ fdn tgt*root -f -t dir -d 2
tgt_root/test directory/$0_T\▯Only
-->tgt_root/test directory/$0_T_Only
tgt_root/!临时文件夹
-->tgt_root/LSW 临时文件夹
tgt_root/\_is▯dir▯%
-->tgt_root/Is_dir*%
tgt*root/测试@#文件夹
-->tgt_root/CS 测试*文件夹
tgt_root/test▯directory
-->tgt_root/Test_Directory
tgt_root
-->Tgt_Root
**************\*\*\*\***************\*\*\*\***************\*\*\*\***************
In order to take effect,add option '-i' or '-c'

    $ fdn tgt_root -f -t dir -d 1
       tgt_root/!临时文件夹
    -->tgt_root/LSW临时文件夹
       tgt_root/_is▯dir▯%
    -->tgt_root/Is_dir_%
       tgt_root/测试@#文件夹
    -->tgt_root/CS测试_文件夹
       tgt_root/test▯directory
    -->tgt_root/Test_Directory
       tgt_root
    -->Tgt_Root
    ********************************************************************
    In order to take effect,add option '-i' or '-c'

-t option

    $  fdn tgt_root -f -t dir
       tgt_root/!临时文件夹
    -->tgt_root/LSW临时文件夹
       tgt_root/测试@#文件夹
    -->tgt_root/CS测试_文件夹
       tgt_root/test▯directory
    -->tgt_root/Test_Directory
       tgt_root/_is▯dir▯%
    -->tgt_root/Is_dir_%
    ********************************************************************
    In order to take effect,add option '-i' or '-c'

    $ fdn tgt_root -f
       tgt_root/thi_Is_File_%.mp4
    -->tgt_root/Thi_Is_File_%.mp4
       tgt_root/$0▯▯测试用文件.html
    -->tgt_root/$0_测试用文件.html
    ********************************************************************
    In order to take effect,add option '-i' or '-c'

-i option

    $ fdn tgt_root/\$0\ \ 测试用文件.html -io
       $0▯▯测试用文件.html
    ==>$0_测试用文件.html

-c option

    $ fdn tgt_root/\$0\ \ 测试用文件.html -c
    $0  测试用文件.html
    Please confirm(y/n/A/q) [no]:
       $0▯▯测试用文件.html
    -->$0_测试用文件.html

    $ fdn tgt_root/\$0\ \ 测试用文件.html -c
    $0  测试用文件.html
    Please confirm(y/n/A/q) [no]: y
       $0▯▯测试用文件.html
    ==>$0_测试用文件.html

-l option

This Option

-f option

    $ fdn tgt_root/\$0\ \ 测试用文件.html
       $0▯▯测试用文件.html
    -->$0_测试用文件.html
    ********************************************************************
    In order to take effect,add option '-i' or '-c'

    $ fdn tgt_root/\$0\ \ 测试用文件.html -f
       tgt_root/$0▯▯测试用文件.html
    -->tgt_root/$0_测试用文件.html
    ********************************************************************
    In order to take effect,add option '-i' or '-c'

-a option

    $ fdn
       a▯Test-file.txt
    -->A_Test_File.txt
    ********************************************************************
    In order to take effect,add option '-i' or '-c'

    $ fdn -a
       /home/hma/a▯Test-file.txt
    -->/home/hma/A_Test_File.txt
    ********************************************************************
    In order to take effect,add option '-i' or '-c'

-r option

    $ fdn tgt_root/\$0_测试用文件.html -r
       $0_测试用文件.html
    -->$0▯▯测试用文件.html
    ********************************************************************
    In order to take effect,add option '-i' or '-c'

-o option

    $ fdn tgt_root/\$0\ \ 测试用文件.html -i
    Exist:$0_测试用文件.html
    Skipped:$0  测试用文件.html
    With option '-o' to enable overwrite.

    $ fdn tgt_root/\$0\ \ 测试用文件.html -io
       $0▯▯测试用文件.html
    ==>$0_测试用文件.html

-p option

    $ fdn tgt_root
       thi_Is_File_%.mp4
    -->Thi_Is_File_%.mp4
       $0▯▯测试用文件.html
    -->$0_测试用文件.html
    ********************************************************************
    In order to take effect,add option '-i' or '-c'

    $ fdn tgt_root -p
       thi_Is_File_%.mp4
    -->Thi_Is_File_%.mp4
       $0▯▯测试用文件.html
    -->$0 _测试用文件.html
    ********************************************************************
    In order to take effect,add option '-i' or '-c'

-e option

    $ fdn tgt_root/\$0_测试用文件.html -re
       $0_测试用文件.html
    -->$0▯▯测试用文件.html
    ********************************************************************
    In order to take effect,add option '-i' or '-c'

## Example/示例

## change one file name/修改一个文件名

    $ fdn tgt_root/\$0\ 测试用文件.html
       $0▯测试用文件.html
    -->$0_测试用文件.html
    ********************************************************************
    In order to take effect,add option '-i' or '-c'

## change files in dir/修改指定目录下文件名

    $ fdn tgt_root
       $0▯测试用文件.html
    -->$0_测试用文件.html
       This▯is▯a▯Test▯file.pdf
    -->This_Is_A_Test_File.pdf
       _thi▯is▯file▯%.mp4
    -->thi_Is_File_%.mp4
       这是测试文件▯.jpg
    -->ZSC这是测试文件.jpg
    ********************************************************************
    In order to take effect,add option '-i' or '-c'

## rollback one file changed/取消一个文件名的修改

    $ fdn tgt_root/\$0_测试用文件.html -r
       $0_测试用文件.html
    -->$0▯测试用文件.html
    ********************************************************************
    In order to take effect,add option '-i' or '-c'

## rollback files changed in dir/取消目录下文件名的修改

    $ fdn tgt_root -r
       This_Is_A_Test_File.pdf
    -->This▯is▯a▯Test▯file.pdf
       ZSC这是测试文件.jpg
    -->这是测试文件▯.jpg
       thi_Is_File_%.mp4
    -->_thi▯▯is▯▯▯file▯%.mp4
       $0_测试用文件.html
    -->$0▯测试用文件.html
    ********************************************************************
    In order to take effect,add option '-i' or '-c'

## joint work with `fd`/与 `fd` 工具联合工作

[fd](https://github.com/sharkdp/fd) is a program to find entries in your
filesytem. It is a simple, fast and user-friendly alternative to find.\*

    $ fdfind -HIi html -x fdn -p {}
       $0▯▯测试用文件.html
    -->$0_测试用文件.html
    ********************************************************************
    In order to take effect,add option '-i' or '-c'

    $ fdfind -HIi html -x fdn -pf {}
       tgt_root/$0▯▯测试用文件.html
    -->tgt_root/$0 _测试用文件.html
    ********************************************************************
    In order to take effect,add option '-i' or '-c'

## 简介

一个小工具，用于日常统一更改文件（或者文件夹）名称

目前的具体格式：

1.  文件名不保留空格（首先空格会被替换为下划线，之后根据是否存在连续下划线来决定缩减）；
2.  文件名中只保留下划线字符，其余的控制类字符会被替换为下划线；
3.  多个连续的下划线字符会被缩减为一个下划线；
4.  如果文件名首字符为下划线将会被删除；
5.  除去扩展名后的文件名如果最后一个字符是下划线也会被删除；
6.  在文件名中保留 [bash special
    parameters](https://www.gnu.org/software/bash/manual/html_node/Special-Parameters.html)
    ;
7.  文件名中包含的一些术语会保留术语本身的大小写写法(持续更新中...​),例如
    `USB`,`PCIe` 等;

## 安装

建议使用[cargo](https://doc.rust-lang.org/cargo/getting-started/installation.html)进行安装:

```bash
$ cargo install fdn
```

## 参数

请前往[Usage/参数](#usage参数) 查看

## 示例

参考 [Example/示例](#example示例) 查看
