fdn
===

Introduction
------------

**fdn** is a CLI tool to normalize file and directory names (underscore separator, configurable
rules) and to **roll back** those renames using a local SQLite journal. Optional **dry-run**
output shows proposed names before you apply them.

Naming behavior is driven by settings in a local database (separator character, term
replacements, and “replace with separator” substrings). Edit those with ``fdn config``.

Typical goals include:

1. Spaces in the basename become the configured separator (default ``_``), with runs of that
   separator collapsed.
2. Other punctuation and control-like characters are normalized using your **config** rules
   (not only a hard-coded underscore mapping).
3. Leading/trailing separator runs on the basename (excluding extension) are trimmed.
4. Bash-style special parameters (e.g. ``$0``, ``$?``) are preserved in the sense described in
   the `bash manual on special parameters <https://www.gnu.org/software/bash/manual/html_node/Special-Parameters.html>`__.
5. Configured **term words** can preserve or rewrite specific tokens (e.g. acronyms).

For exact behavior, run ``fdn`` on a copy of your files or use dry-run (default without
``-i`` / ``-c``).

Data directory
--------------

State is stored under the user data directory:

* **macOS / Linux:** ``~/.fdn/``
* **Windows:** ``%USERPROFILE%\.fdn\`` (when ``HOME`` is unset, the tool falls back similarly to
  Go’s user home directory rules)

Single database file:

* ``fdn.db`` — holds both **configuration** (separator, term words, separator-words) and
  **rename records** (journal for reverse renames).

On first startup, if ``fdn.db`` is missing but legacy ``cfg.db`` and/or ``rd.db`` exist in the
same folder, they are **merged or renamed** into ``fdn.db`` automatically.

Installation
------------

macOS or Linux (Homebrew)
~~~~~~~~~~~~~~~~~~~~~~~~~

If you use `Homebrew <https://brew.sh/>`__ and the tap is available:

.. code:: bash

   brew tap hobbymarks/release https://github.com/hobbymarks/release
   brew install --cask fdn

Any platform (install from source with Go)
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

1. Install `Go <https://go.dev/dl/>`__ **1.26** or newer (see ``go.mod`` in this repository).

2. Install or build the binary:

   .. code:: bash

      go install github.com/hobbymarks/fdn@latest

   The executable is placed in ``$(go env GOPATH)/bin``. Ensure that directory is on your
   ``PATH``.

   Alternatively, clone the repository and run:

   .. code:: bash

      git clone https://github.com/hobbymarks/fdn.git
      cd fdn
      go build -o fdn .

   Then move ``fdn`` to a directory on your ``PATH``.

Windows
~~~~~~~

**Scoop** (recommended). GoReleaser publishes a Scoop manifest to the `hobbymarks/release
<https://github.com/hobbymarks/release>`__ repository (``scoop-fdn/fdn.json``). With `Scoop
<https://scoop.sh/>`__ installed, run:

.. code:: powershell

   scoop install https://raw.githubusercontent.com/hobbymarks/release/main/scoop-fdn/fdn.json

Updates:

.. code:: powershell

   scoop update fdn

**Go (from source).** Install Go from https://go.dev/dl/ and use the same ``go install`` command
from **PowerShell** or **cmd** (ensure ``%USERPROFILE%\go\bin`` or your ``GOPATH\bin`` is on
``PATH``).

**Local build.** Clone the repo and run ``go build -o fdn.exe .``, then add the folder containing
``fdn.exe`` to ``PATH``.

Shell completion
~~~~~~~~~~~~~~~~

After installation, generate scripts with:

.. code:: bash

   fdn completion bash
   fdn completion zsh
   fdn completion fish
   fdn completion powershell

Follow the instructions printed by your shell’s plugin documentation to load the script.

Usage overview
--------------

.. code:: text

   fdn [flags]                    # scan path(s), print or apply renames
   fdn [command] [flags] [args]   # config, mv, completion, help

Run ``fdn --help`` or ``fdn <command> --help`` for the full flag list.

Root command (bulk rename)
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. list-table::
   :widths: 20 30 50
   :header-rows: 1

   * - Short
     - Long
     - Description
   * - ``-c``
     - ``--confirm``
     - Prompt per file: apply rename, skip, apply all, or quit (stdin).
   * - ``-d``
     - ``--directory``
     - Only consider directories (default is regular files only).
   * - ``-f``
     - ``--fullpath``
     - Show full paths in output instead of basenames.
   * - ``-i``
     - ``--inplace``
     - Apply renames (otherwise dry-run).
   * - ``-l``
     - ``--level``
     - Max directory depth when recursing (default ``1``).
   * - ``-o``
     - ``--overwrite``
     - Allow replacing an existing destination when applying a rename.
   * - ``-p``
     - ``--path``
     - Input path(s); repeatable (default ``.``).
   * - ``-r``
     - ``--reverse``
     - Use the journal to restore previous names where possible.
   * - ``-s``
     - ``--plainstyle``
     - Plain text output (no color / diff styling).
   * - ``-e``
     - ``--pretty``
     - Richer diff alignment in colored output.
   * - ``-V``
     - ``--verbose``
     - More log output.
   * - ``-v``
     - ``--version``
     - Print version and exit.
   * - ``-h``
     - ``--help``
     - Help for ``fdn``.

Confirm mode prints ``Please confirm (all,yes,no,quit):`` and accepts responses such as
``all``/``a``, ``yes``/``y``, ``no``/``n``, ``quit``/``q``.

``fdn config``
~~~~~~~~~~~~~~

Manage ``fdn.db`` settings (separator, term list, separator-word list). Typical patterns:

.. code:: bash

   fdn config -l sep
   fdn config -c sep _
   fdn config -c twl "MyBrand:mybrand"
   fdn config -c swl "·"
   fdn config -d twl mybrand

See ``fdn config --help`` for full examples and flag aliases (``sep`` / ``twl`` / ``swl``).

``fdn mv``
~~~~~~~~~~

Move or rename a single file or directory and update the same journal as the main command:

.. code:: bash

   fdn mv ./a.txt ./b.txt
   fdn mv ./doc.pdf ./backup/
   fdn mv "My File.txt" ./inbox/My_File.txt

If the destination exists and is a directory, the source is placed inside it. Existing
non-directory destinations are **not** overwritten unless your workflow uses other tools.

Output modes
------------

When you run **without** ``-i`` (in-place) or without confirming with ``-c``, **fdn** only
**shows** proposed renames (dry run).

* ``-->`` — proposed change (not applied).
* ``==>`` — applied rename (in-place or after confirmation).
* In colored mode, spaces may be shown as ``▯`` for readability; that symbol is **display-only**.

Examples
--------

Dry-run on the current directory:

.. code:: bash

   fdn

Apply renames under ``./mydir`` (depth 1):

.. code:: bash

   fdn -p mydir -i

Reverse (journal) on a path, then apply:

.. code:: bash

   fdn -p mydir -r -i

Use with `fd <https://github.com/sharkdp/fd>`__ (invoke **fdn** per match; pass flags after
``--`` if your shell requires it). On some Linux packages the executable is ``fdfind`` instead
of ``fd``.

.. code:: bash

   fd -e html -x fdn -p {} -i

简介（中文）
------------

**fdn** 是用于统一整理文件/目录名称的 CLI，可对照本地数据库中的规则进行规范化，并支持依据
日志做**撤销式**回滚。默认不写入磁盘（干跑）；使用 ``-i`` 或 ``-c`` 才会真正改名。

配置与改名记录保存在用户目录下的 ``.fdn/fdn.db``（已从旧的 ``cfg.db`` / ``rd.db`` 自动合并
或迁移）。使用 ``fdn config`` 维护分隔符与词条等规则。

安装方式见 `Installation`_：Homebrew（若可用）、Windows 上可用 `Scoop <https://scoop.sh/>`__
安装发布清单，或使用 Go 源码安装 / 本地 ``go build`` 并配置 ``PATH``。

更多参数与示例请运行 ``fdn --help``、``fdn config --help``、``fdn mv --help``。
