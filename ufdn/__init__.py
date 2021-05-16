"""
UFN use to uniformly change file names
"""
from __future__ import absolute_import
from __future__ import print_function

from pathlib import Path
import os
import sys

import click
import colorama
from colorama import Back
from colorama import Fore
from colorama import Style

from ufdn.ufdnlib import uconfig
from ufdn.ufdnlib import ucrypt
from ufdn.ufdnlib import udb
from ufdn.ufdnlib import ufncli
from ufdn.ufdnlib import utils


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

                uconfig.gParamDict["LowerCaseWordSet"] = set(
                    list(map(lambda x: x.lower(), words.words())))
            except LookupError:
                nltk.download("words")
        from nltk.corpus import words

        uconfig.gParamDict["LowerCaseWordSet"] = set(
            list(map(lambda x: x.lower(), words.words())))
        uconfig.gParamDict["record_path"] = os.path.join(Path.home(), ".ufdn")
        Path(uconfig.gParamDict["record_path"]).mkdir(parents=True,
                                                      exist_ok=True)
        #######################################################################
        ufncli.ufn()
    finally:
        colorama.deinit()


if __name__ == '__main__':
    try:
        sys.exit(main())
    except Exception as e:
        sys.stderr.write(f"Error:{str(e)}\n")
        sys.exit(1)
