#!/usr/bin/env python
import os
from pathlib import Path
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


@click.command(context_settings={
    "ignore_unknown_options": True,
    "help_option_names": ["-h", "--help"]
})
@click.argument("path", required=False, type=click.Path(exists=True), nargs=-1)
@click.option("--max-depth",
              "-d",
              default=1,
              type=int,
              help="Set travel directory tree with max depth.",
              show_default=True)
@click.option(
    "--type",
    "-t",
    default="file",
    type=click.Choice(["file", "dir"]),
    help="File types.If file ,only change file names,If dir,only change "
    "directory names.",
    show_default=True)
@click.option("--in-place",
              "-i",
              default=False,
              type=bool,
              is_flag=True,
              help="Make changes to file name in place.",
              show_default=True)
@click.option("--confirm",
              "-c",
              default=False,
              type=bool,
              is_flag=True,
              help="If confirm is True will need confirmation.",
              show_default=True)
@click.option("--is-link",
              "-l",
              default=False,
              type=bool,
              is_flag=True,
              help="Follow the real path of a link.",
              show_default=True)
@click.option("--full",
              "-f",
              default=False,
              type=bool,
              is_flag=True,
              help="Show full path.",
              show_default=True)
@click.option("--rollback",
              "-r",
              default=False,
              type=bool,
              is_flag=True,
              help="To roll back changed file names.",
              show_default=True)
@click.option("--overwrite",
              "-o",
              default=False,
              type=bool,
              is_flag=True,
              help="Overwrite exist files.",
              show_default=True)
@click.option("--pretty",
              "-p",
              default=False,
              type=bool,
              is_flag=True,
              help="Try to pretty output.",
              show_default=True)
def ufn(path, max_depth, type, in_place, confirm, is_link, full, rollback,
        overwrite,pretty):
    """Files in PATH will be changed file names unified.
    
    You can direct set path such as UFn.py path ...
    """
    if not path:
        config.gParamDict["path"] = ["."]
    else:
        config.gParamDict["path"] = path
    if str(max_depth).isnumeric():
        config.gParamDict["max_depth"] = int(str(max_depth))
    else:
        config.gParamDict["max_depth"] = 1
    config.gParamDict["type"] = type
    if in_place:
        config.gParamDict["in_place"] = True
    else:
        config.gParamDict["in_place"] = False
    if confirm:
        config.gParamDict["confirm"] = True
    else:
        config.gParamDict["confirm"] = False
    if is_link:
        config.gParamDict["is_link"] = True
    else:
        config.gParamDict["is_link"] = False
    if full:
        config.gParamDict["full_path"] = True
    else:
        config.gParamDict["full_path"] = False
    if overwrite:
        config.gParamDict["overwrite"] = True
    else:
        config.gParamDict["overwrite"] = False
    if pretty:
        config.gParamDict["pretty"] = True
    else:
        config.gParamDict["pretty"] = False
    config.gParamDict["AllInPlace"] = False
    if rollback:
        rb = True
    else:
        rb = False
    config.gParamDict["latest_confirm"] = utils.unify_confirm()  # Parameter is
    # Null to get default return
    for pth in config.gParamDict["path"]:
        if os.path.isfile(pth) and utils.type_matched(pth):
            if not rb:
                utils.one_file_ufn(pth)
            else:
                utils.one_file_rbk(pth)
        elif os.path.isdir(pth):
            if not rb:
                utils.one_dir_ufn(pth)
            else:
                utils.one_dir_rbk(pth)
        else:
            click.echo(f"{Fore.RED}Not valid:{pth}{Fore.RESET}")

    if (not config.gParamDict["in_place"]) and (
            not config.gParamDict["confirm"]):
        cols, _ = os.get_terminal_size()
        click.echo("*" * cols)
        click.echo(
            "In order to take effect,add option '-i' or '-c'")


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
