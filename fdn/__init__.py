"""fdn use to uniformly change file or directory names and also support
rollback these operations. """
from __future__ import print_function

import sys

# From Third party
import click
import colorama
from colorama import Back
from colorama import Fore
from colorama import Style

# From Project
from fdn.fdnlib.fdncli import ufn

__version__ = "2022.02.16.2704"


def main() -> None:
    try:
        colorama.init()
        if (sys.version_info.major, sys.version_info.minor) < (3, 8):
            click.echo(
                f"{Fore.RED}current is {sys.version},\n"
                f"{Back.WHITE}Please upgrade to >=3.8.{Style.RESET_ALL}")
            sys.exit()
        #######################################################################
        ufn()
        #######################################################################
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
