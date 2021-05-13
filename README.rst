====
UFDN
====

Introduction
============
A tool for uniformly changing file or directory names

.. contents::

Bash Version
~~~~~~~~~~~~

A tool base on `rename <http://plasmasturm.org/code/rename/>`__ to unify
file name.

File name format:

 1. no space in file name（firstly every space will be replaced by an underscore,then multiple consecutive underscores will be reduced to one);
 #. only underscore allowed in file name，all other control characters will be replaced by underscore;
 #. multiple consecutive underscores will be reduced to one;
 #. underscore at the beginning of file name will be deleted;
 #. underscore at the end of file name will be deleted;

Python Version
~~~~~~~~~~~~

A pure python tool to unify file or directory name.

File or directory name format:

 1. no space in file name（firstly every space will be replaced by an underscore,then multiple consecutive underscores will be reduced to one);
 #. only underscore allowed in file name，all other control characters will be replaced by underscore;
 #. multiple consecutive underscores will be reduced to one;
 #. underscore at the beginning of file name will be deleted;
 #. underscore at the end of file name will be deleted;
 #. keep `bash special parameters <https://www.gnu.org/software/bash/manual/html_node/Special-Parameters.html>`__ in file name;


Installation
============

.. _bash-version-1:

Bash Version
~~~~~~~~~~~~

-  The content is `bash <https://www.gnu.org/software/bash/>`__ script，so you can download and run directly;
-  If you wan to run UFn.sh any where,please make sure place UFn.sh in /usr/local/bin or /usr/bin or other directory include in
   `Environment_variable <https://en.wikipedia.org/wiki/Environment_variable>`__\ ；
-  If you wan to run the script in some directory , you can create `soft link <https://en.wikipedia.org/wiki/Ln_(Unix)>`__\ (ln -s UFn.sh TargetName)

.. _python-version-1:


Python Version
~~~~~~~~~~~~
To install UFDN from conda:

.. code-block:: shell

    $ conda install -c hobbymarks ufdn

Usage
====
Options::

    Usage: ufdn [OPTIONS] [PATH]...

      Files in PATH will be changed file names unified.

      You can direct set path such as ufncli.py path ...

    Options:
      -d, --max-depth INTEGER  Set travel directory tree with max depth.
                               [default: 1]
      -t, --type [file|dir]    Set types.If the value is 'file' ,only change file
                               names,If the value is 'dir',only change directory
                               names.  [default: file]
      -i, --in-place           Changes file name in place.  [default: False]
      -c, --confirm            Need confirmation before change to take effect.
                               [default: False]
      -l, --is-link            Follow the real path of a link.  [default: False]
      -f, --full-path          Show full path of file.  [default: False]
      -r, --roll-back          To roll back changed file names.  [default: False]
      -o, --overwrite          Overwrite exist files.  [default: False]
      -p, --pretty             Try to pretty output.  [default: False]
      -e, --enhanced-display   Enhanced display output.  [default: False]
      -h, --help               Show this message and exit.

Example
=======

一个小工具，用于日常统一更改文件名


Bash版
~~~~~~

基于\ `rename <http://plasmasturm.org/code/rename/>`__\ 的一个小工具，用于日常统一更改资料的文件名。

目前的具体格式：

 1. 文件名不保留空格（首先空格会被替换为下划线，之后根据是否存在连续下划线来决定缩减）；
 #. 文件名中只保留下划线字符，其余的控制类字符会被替换为下划线；
 #. 多个连续的下划线字符会被缩减为一个下划线；
 #. 如果文件名首字符为下划线将会被删除；
 #. 除去扩展名后的文件名如果最后一个字符是下划线也会被删除；

Python 版
~~~~~~~~~

纯python编写，用于日常统一更改资料的文件名。

目前的具体格式：

 1. 文件名不保留空格（首先空格会被替换为下划线，之后根据是否存在连续下划线来决定缩减）；
 #. 文件名中只保留下划线字符，其余的控制类字符会被替换为下划线；
 #. 多个连续的下划线字符会被缩减为一个下划线；
 #. 如果文件名首字符为下划线将会被删除；
 #. 除去扩展名后的文件名如果最后一个字符是下划线也会被删除；
 #. 在文件名中保留 `bash special parameters <https://www.gnu.org/software/bash/manual/html_node/Special-Parameters.html>`__;

安装
----

.. _bash版-1:

Bash版
~~~~~~

-  内容为\ `bash <https://www.gnu.org/software/bash/>`__\ 脚本，可以直接下载和执行;
-  将UFn.sh放置在/usr/local/bin 或者/usr/bin 或者其它\ `环境变量 <https://en.wikipedia.org/wiki/Environment_variable>`__\ 包含的目录，这样可以在任意目录执行该脚本；
-  如果需要其它目录执行可以考虑创建\ `软连接 <https://en.wikipedia.org/wiki/Ln_(Unix)>`__\ (ln -s UFn.sh TargetName)

.. _python-版-1:

Python 版
~~~~~~~~~

