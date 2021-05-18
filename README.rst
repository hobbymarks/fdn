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

A \ `Python <https://www.python.org/>`_ \ tool to unify file or directory name.

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

    Options:
      -d, --max-depth INTEGER     Set travel directory tree with max depth.
                                  [default: 1]
      -t, --file-type [file|dir]  Set type.If 'file',operations are only valid for
                                  file,If 'dir',operations are only valid for
                                  directory.  [default: file]
      -i, --in-place              Changes file name in place.  [default: False]
      -c, --confirm               Need confirmation before change to take effect.
                                  [default: False]
      -l, --is-link               Follow the real path of a link.  [default:
                                  False]
      -f, --full-path             Show full path of file.Relative to the input
                                  path.  [default: False]
      -a, --absolute-path         Show absolute path of file.  [default: False]
      -r, --roll-back             To roll back changed file names.  [default:
                                  False]
      -o, --overwrite             Overwrite exist files.  [default: False]
      -p, --pretty                Try to pretty output.  [default: False]
      -e, --enhanced-display      Enhanced display output.  [default: False]
      --version                   Show the version and exit.
      -h, --help                  Show this message and exit.


-d option
    .. code-block:: shell

        $  ufdn tgt_root -f -t dir -d 2
           tgt_root/test directory/$0_T\▯Only
        -->tgt_root/test directory/$0_T_Only
           tgt_root/!临时文件夹
        -->tgt_root/LSW临时文件夹
           tgt_root/测试@#文件夹
        -->tgt_root/CS测试_文件夹
           tgt_root/test▯directory
        -->tgt_root/Test_Directory
           tgt_root/_is▯dir▯%
        -->tgt_root/Is_dir_%
        ***************************************************
        In order to take effect,add option '-i' or '-c'

-t option
    .. code-block:: shell

        $  ufdn tgt_root -f -t dir
           tgt_root/!临时文件夹
        -->tgt_root/LSW临时文件夹
           tgt_root/测试@#文件夹
        -->tgt_root/CS测试_文件夹
           tgt_root/test▯directory
        -->tgt_root/Test_Directory
           tgt_root/_is▯dir▯%
        -->tgt_root/Is_dir_%
        ***************************************************
        In order to take effect,add option '-i' or '-c'

-i option
    .. code-block:: shell

        $ ufdn tgt_root/\$0\ 测试用文件.html -i
           $0▯测试用文件.html
        ==>$0_测试用文件.html

-c option
    .. code-block:: shell

        $ ufdn tgt_root/\$0_测试用文件.html -rc
        $0_测试用文件.html
        Please confirm(y/n/A/q) [no]: y
           $0_测试用文件.html
        ==>$0▯测试用文件.html

-l option
    This Option

-f option
    .. code-block:: shell

        $ ufdn tgt_root/\$0\ 测试用文件.html
           $0▯测试用文件.html
        -->$0_测试用文件.html
        ***************************************************
        In order to take effect,add option '-i' or '-c'

        $ ufdn tgt_root/\$0\ 测试用文件.html -f
           tgt_root/$0▯测试用文件.html
        -->tgt_root/$0_测试用文件.html
        ****************************************************
        In order to take effect,add option '-i' or '-c'

-r option
    .. code-block:: shell

        $ ufdn tgt_root/\$0_测试用文件.html -r
           $0_测试用文件.html
        -->$0▯测试用文件.html
        ***************************************************
        In order to take effect,add option '-i' or '-c'

-o option
    This Option

-p option
    .. code-block:: shell

        $ ufdn tgt_root
           $0▯测试用文件.html
        -->$0_测试用文件.html
           This▯is▯a▯Test▯file.pdf
        -->This_Is_A_Test_File.pdf
           这是测试文件▯.jpg
        -->ZSC这是测试文件.jpg
           _thi▯▯is▯▯▯file▯%.mp4
        -->thi_Is_File_%.mp4
        ***************************************************
        In order to take effect,add option '-i' or '-c'

        $ ufdn tgt_root -p
           $0▯测试用文件.html
        -->$0_测试用文件.html
           This▯is▯a▯Test▯file.pdf
        -->This_Is_A_Test_File.pdf
              这是测试文件▯.jpg
        -->ZSC这是测试文件 .jpg
           _thi▯▯is▯▯▯file▯%.mp4
        --> thi _Is  _File_%.mp4
        ***************************************************
        In order to take effect,add option '-i' or '-c'

-e option
    This Option


.. _reference_example:
Example
=======

change one file name
~~~~~~~~~~~~~~~~~~~~

.. code-block:: shell

    $ ufdn tgt_root/\$0\ 测试用文件.html
       $0▯测试用文件.html
    -->$0_测试用文件.html
    ********************************************************************************
    In order to take effect,add option '-i' or '-c'
.. image:: ./docs/images/UFDN_one_file_NoOp.png
  :width: 680px
  :align: center

change files in dir
~~~~~~~~~~~~~~~~~~~~

.. code-block:: shell

    $ ufdn tgt_root
       $0▯测试用文件.html
    -->$0_测试用文件.html
       This▯is▯a▯Test▯file.pdf
    -->This_Is_A_Test_File.pdf
       _thi▯is▯file▯%.mp4
    -->thi_Is_File_%.mp4
       这是测试文件▯.jpg
    -->ZSC这是测试文件.jpg
    ********************************************************************************
    In order to take effect,add option '-i' or '-c'


.. image:: ./docs/images/UFDN_one_dir_NoOp.png
  :width: 680px
  :align: center

rollback one file changed
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: shell

    $ ufdn tgt_root/\$0_测试用文件.html -r
       $0_测试用文件.html
    -->$0▯测试用文件.html
    *******************************************************************************
    In order to take effect,add option '-i' or '-c'
.. image:: ./docs/images/UFDN_one_file_rOp.png
  :width: 680px
  :align: center

rollback files changed in dir
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: shell

    $ ufdn tgt_root -r
       This_Is_A_Test_File.pdf
    -->This▯is▯a▯Test▯file.pdf
       ZSC这是测试文件.jpg
    -->这是测试文件▯.jpg
       thi_Is_File_%.mp4
    -->_thi▯▯is▯▯▯file▯%.mp4
       $0_测试用文件.html
    -->$0▯测试用文件.html
    *******************************************************************************
    In order to take effect,add option '-i' or '-c'
.. image:: ./docs/images/UFDN_one_dir_rOp.png
  :width: 680px
  :align: center

简介
====
一个小工具，用于日常统一更改文件（或者文件夹）名称

Bash版
~~~~~~

基于\ `rename <http://plasmasturm.org/code/rename/>`__\ 的一个小工具，用 `bash <https://www.gnu.org/software/bash/>`__ 编写,用于日常统一更改资料的文件名。

目前的具体格式：

 1. 文件名中不保留空格（空格首先会被替换为下划线，之后根据是否存在连续下划线来决定缩减）；
 #. 文件名中只保留下划线字符，其余的控制类字符会被替换为下划线；
 #. 多个连续的下划线字符会被缩减为一个下划线；
 #. 如果文件名首字符为下划线将会被删除；
 #. 除去扩展名后的文件名如果最后一个字符是下划线也会被删除；

Python 版
~~~~~~~~~

用\ `Python <https://www.python.org/>`_ \编写，用于日常统一更改资料的文件名。

目前的具体格式：

 1. 文件名不保留空格（首先空格会被替换为下划线，之后根据是否存在连续下划线来决定缩减）；
 #. 文件名中只保留下划线字符，其余的控制类字符会被替换为下划线；
 #. 多个连续的下划线字符会被缩减为一个下划线；
 #. 如果文件名首字符为下划线将会被删除；
 #. 除去扩展名后的文件名如果最后一个字符是下划线也会被删除；
 #. 在文件名中保留 `bash special parameters <https://www.gnu.org/software/bash/manual/html_node/Special-Parameters.html>`__;

安装
====

.. _bash版-1:

Bash版
~~~~~~

-  内容为\ `bash <https://www.gnu.org/software/bash/>`__\ 脚本，可以直接下载和执行;
-  将UFn.sh放置在/usr/local/bin 或者/usr/bin 或者其它\ `环境变量 <https://en.wikipedia.org/wiki/Environment_variable>`__\ 包含的目录，这样可以在任意目录执行该脚本；
-  如果需要其它目录执行可以考虑创建\ `软连接 <https://en.wikipedia.org/wiki/Ln_(Unix)>`__\ (ln -s UFn.sh TargetName)

.. _python-版-1:

Python 版
~~~~~~~~~
建议使用conda进行安装:

.. code-block:: shell

    $ conda install -c hobbymarks ufdn

示例
====

