import os
import shutil
from pathlib import Path
from typing import List, Optional

# From Third party
import click
from colorama import Fore
# From This Project
from fdn.fdnlib import fdnutils, utils
from fdn.fdnlib.fdncfg import gParamDict as ugPD


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
    "--file-type",
    "-t",
    default="file",
    type=click.Choice(["file", "dir"]),
    help=f"Set type.If 'file',operations are only valid for file,If 'dir',"
    f"operations are only valid for directory.",
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
              help=f"Show full path of file.Relative to the input path.",
              show_default=True)
@click.option("--absolute-path",
              "-a",
              default=False,
              type=bool,
              is_flag=True,
              help=f"Show absolute path of file.",
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
@click.version_option(version="2022.02.16.2704")
def ufn(path: Optional[List[Path]], max_depth: int, file_type: str,
        in_place: bool, confirm: bool, is_link: bool, full_path: bool,
        absolute_path: bool, roll_back: bool, overwrite: bool, pretty: bool,
        enhanced_display: bool):
    """Files in PATH will be changed file names unified.
    """
    if not path:
        ugPD["path"] = ["."]
    else:
        ugPD["path"] = path
    ugPD["max_depth"] = max_depth
    ugPD["type"] = file_type
    ugPD["in_place_flag"] = in_place
    ugPD["need_confirmation_flag"] = confirm
    ugPD["is_link"] = is_link
    ugPD["full_path_flag"] = full_path
    ugPD["absolute_path_flag"] = absolute_path
    ugPD["overwrite_flag"] = overwrite
    ugPD["pretty_flag"] = pretty
    ugPD["enhanced_display_flag"] = enhanced_display
    ugPD["all_in_place_flag"] = False
    ugPD["latest_confirm"] = utils.unify_confirm(
    )  # Parameter is Null to get default return
    ugPD["target_appeared"] = False
    for pth in ugPD["path"]:
        pth = (pth[:-1] if pth.endswith(os.sep) else pth)
        if os.path.isfile(pth) and fdnutils.type_matched(pth):
            if not roll_back:
                fdnutils.one_file_ufn(pth)
            else:
                fdnutils.one_file_rbk(pth)
        elif os.path.isdir(pth):
            if not roll_back:
                fdnutils.one_dir_ufn(pth)
            else:
                fdnutils.one_dir_rbk(pth)
        else:
            click.echo(f"{Fore.RED}Not valid:{pth}{Fore.RESET}")

    if (not ugPD["in_place_flag"]) and (
            not ugPD["need_confirmation_flag"]) and (ugPD["target_appeared"]):
        cols, _ = shutil.get_terminal_size(fallback=(79, 23))
        click.echo("*" * cols)
        click.echo("In order to take effect,add option '-i' or '-c'")
        ugPD["target_appeared"] = False


# TODO: global version how to
# TODO: **support edit config data
# TODO: display config data
# TODO: display total summary
# TODO: support regular expression input path or directory as path argument
# TODO: exclude spec directory or file type
# TODO: %,'s,How to ...

# TODO: unify full_path request to a function ...
# TODO: Embed nltk ,but need download first time ,why?
# TODO: multi dir may accure same name conflict so record path ID?
# TODO: support user select roll back when multi target

# TODO: bug:mac osx begin with ._ hidden check wrong
# TODO: bug:fd -x can not get terminal size .so fall back and no color display
# TODO: autocomplete
# TODO: different language can use separator for terminology
# TODO: C++ --> C ...?!
