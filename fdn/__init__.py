"""fdn use to uniformly change file or directory names and also support
rollback these operations. """
import click

import os
import sys
from pathlib import Path

# From Third party

import colorama
from colorama import Back
from colorama import Fore
from colorama import Style

from fdn.fdnlib.fdncfg import gParamDict as ugPD
from fdn.fdnlib.fdncli import ufn

__version__ = "2021.05.26.2637"


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
        ugPD["record_path"] = os.path.join(Path.home(), ".fdn")
        Path(ugPD["record_path"]).mkdir(parents=True, exist_ok=True)
        ugPD["db_path"] = os.path.join(ugPD["record_path"], "rdsa.db")
        #######################################################################
        ufn()
    finally:
        colorama.deinit()


def run_main():
    try:
        sys.exit(main())
    except Exception as e:
        sys.stderr.write(f"Error:{str(e)}\n")
        sys.exit(1)


if __name__ == '__main__':
    run_main()
