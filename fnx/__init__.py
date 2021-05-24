"""fnx use to uniformly change file or directory names and also support
rollback these operations. """
from __future__ import print_function

import os
import sys
from pathlib import Path

import click
import colorama
from colorama import Back
from colorama import Fore
from colorama import Style

from fnx.fnxlib import fnxcli
from fnx.fnxlib import ucrypt
from fnx.fnxlib import udb
from fnx.fnxlib import utils
from fnx.fnxlib.fnxcfg import gParamDict as ugPD


def main() -> None:
    try:
        colorama.init()
        if (sys.version_info.major, sys.version_info.minor) < (3, 8):
            click.echo(
                f"{Fore.RED}current is {sys.version},\n"
                f"{Back.WHITE}Please upgrade to >=3.8.{Style.RESET_ALL}")
            sys.exit()
        #######################################################################
        app_path = os.path.dirname(os.path.realpath(__file__))
        nltk_path = os.path.join(app_path, "nltk_data")
        import nltk

        if os.path.isdir(nltk_path):
            nltk.data.path.append(nltk_path)
            if not os.path.isfile(
                    os.path.join(nltk_path, "corpora", "words.zip")):
                nltk.download("words", download_dir=nltk_path)
        else:
            try:
                from nltk.corpus import words

                ugPD["LowerCaseWordSet"] = set(
                    list(map(lambda x: x.lower(), words.words())))
            except LookupError:
                nltk.download("words")
        from nltk.corpus import words

        ugPD["LowerCaseWordSet"] = set(
            list(map(lambda x: x.lower(), words.words())))
        ugPD["record_path"] = os.path.join(Path.home(), ".fnx")
        Path(ugPD["record_path"]).mkdir(parents=True, exist_ok=True)
        ugPD["db_path"] = os.path.join(ugPD["record_path"], "rdsa.db")
        #######################################################################
        fnxcli.ufn()
    finally:
        colorama.deinit()


if __name__ == '__main__':
    try:
        sys.exit(main())
    except Exception as e:
        sys.stderr.write(f"Error:{str(e)}\n")
        sys.exit(1)
