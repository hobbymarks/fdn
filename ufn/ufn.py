#!/usr/bin/env python
import os
from pathlib import Path
import shutil
import sys

# From Third party
import click
import colorama
from colorama import Back
from colorama import Fore
from colorama import Style

# From This Project
import config
import utils


@click.command(
    context_settings={
        "ignore_unknown_options": True,
        "help_option_names": ["-h", "--help"],
        "max_content_width": shutil.get_terminal_size()[0]
    })
@click.argument("path", required=False, type=click.Path(exists=True), nargs=-1)
@click.option("--max-depth",
              "-d",
              default=1,
              type=int,
              help=f"Set travel directory tree with max depth.",
              show_default=True)
@click.option(
    "--type",
    "-t",
    default="file",
    type=click.Choice(["file", "dir"]),
    help=
    f"Set types.If the value is 'file' ,only change file names,If the value "
    f"is 'dir',only change directory names.",
    show_default=True)
@click.option("--in-place",
              "-i",
              default=False,
              type=bool,
              is_flag=True,
              help=f"Changes file name in place.",
              show_default=True)
@click.option("--confirm",
              "-c",
              default=False,
              type=bool,
              is_flag=True,
              help=f"Need confirmation before change to take effect.",
              show_default=True)
@click.option("--is-link",
              "-l",
              default=False,
              type=bool,
              is_flag=True,
              help=f"Follow the real path of a link.",
              show_default=True)
@click.option("--full-path",
              "-f",
              default=False,
              type=bool,
              is_flag=True,
              help="Show full path of file.",
              show_default=True)
@click.option("--roll-back",
              "-r",
              default=False,
              type=bool,
              is_flag=True,
              help=f"To roll back changed file names.",
              show_default=True)
@click.option("--overwrite",
              "-o",
              default=False,
              type=bool,
              is_flag=True,
              help=f"Overwrite exist files.",
              show_default=True)
@click.option("--pretty",
              "-p",
              default=False,
              type=bool,
              is_flag=True,
              help=f"Try to pretty output.",
              show_default=True)
@click.option("--enhanced-display",
              "-e",
              default=False,
              type=bool,
              is_flag=True,
              help=f"Enhanced display output.",
              show_default=True)
def ufn(path, max_depth, type, in_place, confirm, is_link, full_path,
        roll_back, overwrite, pretty, enhanced_display):
    """Files in PATH will be changed file names unified.
    
    You can direct set path such as ufn.py path ...
    """
    if not path:
        config.gParamDict["path"] = ["."]
    else:
        config.gParamDict["path"] = path
    config.gParamDict["max_depth"] = max_depth
    config.gParamDict["type"] = type
    config.gParamDict["in_place"] = in_place
    config.gParamDict["confirm"] = confirm
    config.gParamDict["is_link"] = is_link
    config.gParamDict["full_path"] = full_path
    config.gParamDict["overwrite"] = overwrite
    config.gParamDict["pretty"] = pretty
    config.gParamDict["enhanced_display"] = enhanced_display
    config.gParamDict["AllInPlace"] = False
    config.gParamDict["latest_confirm"] = utils.unify_confirm()  # Parameter is
    # Null to get default return
    for pth in config.gParamDict["path"]:
        if os.path.isfile(pth) and utils.type_matched(pth):
            if not roll_back:
                utils.one_file_ufn(pth)
            else:
                utils.one_file_rbk(pth)
        elif os.path.isdir(pth):
            if not roll_back:
                utils.one_dir_ufn(pth)
            else:
                utils.one_dir_rbk(pth)
        else:
            click.echo(f"{Fore.RED}Not valid:{pth}{Fore.RESET}")

    if (not config.gParamDict["in_place"]) and (
            not config.gParamDict["confirm"]):
        cols, _ = shutil.get_terminal_size(fallback=(79, 23))
        click.echo("*" * cols)
        click.echo("In order to take effect,add option '-i' or '-c'")


if __name__ == "__main__":
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

                config.gParamDict["LowerCaseWordSet"] = set(
                    list(map(lambda x: x.lower(), words.words())))
            except LookupError:
                nltk.download("words")
        from nltk.corpus import words

        config.gParamDict["LowerCaseWordSet"] = set(
            list(map(lambda x: x.lower(), words.words())))
        config.gParamDict["record_path"] = os.path.join(app_path, "rd_data")
        Path(config.gParamDict["record_path"]).mkdir(parents=True,
                                                     exist_ok=True)
        #######################################################################
        ufn()
    finally:
        colorama.deinit()

# TODO: **support edit config data
# TODO: display config data
# TODO: display total summary
# TODO: display progress bar at bottom ...
# TODO: support regular expression input path or directory as path argument
# TODO: exclude spec directory or file type
# TODO: %,'s,How to ...

# TODO: unify full_path request to a function ...
