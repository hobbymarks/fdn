#!/usr/bin/env python
import difflib
import os
import re
import string
import sys
from pathlib import Path

# From Third party
import click
import colorama
try:
    from click_default_group import DefaultGroup
except:
    from utils import DefaultGroup
from colorama import Back
from colorama import Fore
from colorama import Style
from unidecode import unidecode

# From This Project
import config
import utils


def is_hidden(f_path):
    """
    Check file is hidden or not
    :param f_path: string,file path
    :return: True if is hidden file,False if is not hidden file
    """
    if os.name == "nt":
        import win32api
        import win32con
    if os.name == "nt":
        attribute = win32api.GetFileAttributes(f_path)
        return attribute & (win32con.FILE_ATTRIBUTE_HIDDEN
                            | win32con.FILE_ATTRIBUTE_SYSTEM)
    else:
        return os.path.basename(f_path).startswith(".")  # linux, osx


def mask_RUW(s):
    """
    Mask a input string by RemainUnchangedWordList value
    :param s: string,input string to be masked
    :return: word list and mask list,if joined the word list,you will get the
    intact s.
    """
    ruw_list = config.gParamDict["RemainUnchangedWordList"]
    re_str = "|".join([re.escape(ruw) for ruw in ruw_list])
    word_list = re.split(f"({re_str})", s)
    mask_list = []
    for elm in word_list:
        if elm in ruw_list:
            mask_list.append(True)
        else:
            mask_list.append(False)
    return word_list, mask_list


def replace_char(f_path):
    c_dict = config.gParamDict["BeReplacedCharDictionary"]
    root, ext = os.path.splitext(f_path)
    word_list, mask_list = mask_RUW(root)
    for c_key, c_value in c_dict.items():
        new_word_list = []
        for word, mask in zip(word_list, mask_list):
            if not mask:
                new_word_list.append(word)
        if c_key in "".join(new_word_list):  # If need replace ,then go ...
            new_word_list = []
            for word, mask in zip(word_list, mask_list):
                if (c_key in word) and (not mask):
                    new_word_list.append(word.replace(c_key, c_value))
                else:
                    new_word_list.append(word)
            word_list = new_word_list
    return "".join(word_list) + ext


def process_head_tail_sep_char(f_path):
    sep_char = config.gParamDict["SeparatorChar"]
    root, ext = os.path.splitext(f_path)
    if root.startswith(sep_char):
        root = sep_char.join(root.split(sep_char)[1:])
    if root.endswith(sep_char):
        root = sep_char.join(root.split(sep_char)[0:-1])
    return root + ext


def process_head_tail(f_path=""):
    assert f_path
    new_name = process_head_tail_sep_char(f_path=f_path)
    # Capitalize The First Letter
    if new_name[0].islower():
        new_name = new_name[0].upper() + new_name[1:]
    return new_name


def process_white_space(f_path=""):
    sep_char = config.gParamDict["SeparatorChar"]
    root, ext = os.path.splitext(f_path)
    new_name = sep_char.join(root.split()) + ext
    return new_name


def check_starts_with_terminology(word=""):
    term_dict = config.gParamDict["TerminologyDictionary"]
    for key in term_dict.keys():
        if word.lower().startswith(key):
            return key
    return None


def process_terminology(f_path=""):
    root, ext = os.path.splitext(f_path)
    word_list = root.split(config.gParamDict["SeparatorChar"])
    new_word_list = []
    term_dict = config.gParamDict["TerminologyDictionary"]
    sep_char = config.gParamDict["SeparatorChar"]
    for word in word_list:
        if word.lower() in term_dict.keys():
            new_word_list.append(term_dict[word.lower()])
        elif key := check_starts_with_terminology(word):
            new_word = term_dict[key] + word[len(key):]
            new_word_list.append(new_word)
        else:
            new_word_list.append(word)
    return sep_char.join(new_word_list) + ext


def asc_head(f_path=""):
    lmt_len = config.gParamDict["ASCLen"]
    sep_char = config.gParamDict["SeparatorChar"]
    head_chars = config.gParamDict["HeadChars"]
    if f_path[0] in head_chars:
        return ""
    word = f_path.split(sep_char)[0]
    if len(word) > lmt_len:
        word = word[0:lmt_len]
    new_word = ""
    for c in word:
        if c not in head_chars:
            new_word += c
        else:
            break
    return "".join([elm[0] for elm in unidecode(new_word).split()])


def process_word(f_path=""):
    sep_char = config.gParamDict["SeparatorChar"]
    root, ext = os.path.splitext(f_path)
    word_list = root.split(sep_char)
    new_word_list = []
    word_set = config.gParamDict["LowerCaseWordSet"]
    for word in word_list:
        if word.lower() in word_set:
            new_word_list.append(string.capwords(word))
        else:
            new_word_list.append(word)
    return sep_char.join(new_word_list) + ext


def depth_walk(top_path, top_down=True, follow_links=False, max_depth=1):
    if str(max_depth).isnumeric():
        max_depth = int(max_depth)
    else:
        max_depth = 1
    names = os.listdir(top_path)
    dirs, non_dirs = [], []
    for name in names:
        if os.path.isdir(os.path.join(top_path, name)):
            dirs.append(name)
        else:
            non_dirs.append(name)
    if top_down:
        yield top_path, dirs, non_dirs
    if max_depth > 1:
        for name in dirs:
            new_path = os.path.join(top_path, name)
            if follow_links or not os.path.islink(new_path):
                for x in depth_walk(new_path, top_down, follow_links,
                                    1 if max_depth == 1 else max_depth - 1):
                    yield x
    if not top_down:
        yield top_path, dirs, non_dirs


def rich_style(original="", processed=""):
    rich_org = rich_proc = ""
    rich_org_dif_pos = rich_proc_dif_pos = 0
    for match in difflib.SequenceMatcher(a=original,
                                         b=processed).get_matching_blocks():
        if rich_org_dif_pos < match.a:
            rich_org += Fore.RED + original[rich_org_dif_pos:match.a].replace(
                " ", "▯") + Fore.RESET + original[match.a:match.a + match.size]
            rich_org_dif_pos = match.a + match.size
        else:
            rich_org += original[match.a:match.a + match.size]
            rich_org_dif_pos = match.a + match.size

        if rich_proc_dif_pos < match.b:
            rich_proc += Fore.GREEN + processed[
                rich_proc_dif_pos:match.b].replace(
                    " ",
                    "▯") + Fore.RESET + processed[match.b:match.b + match.size]
            rich_proc_dif_pos = match.b + match.size
        else:
            rich_proc += processed[match.b:match.b + match.size]
            rich_proc_dif_pos = match.b + match.size
    return rich_org, rich_proc


def out_info(file, new_name):
    rich_org, rich_proc = rich_style(file, new_name)
    if config.gParamDict["AlternateFlag"]:
        click.echo(Back.LIGHTWHITE_EX + " " * 3 + rich_org + Style.RESET_ALL)
    else:
        click.echo(" " * 3 + rich_org)
    if config.gParamDict["dry_run"]:
        if config.gParamDict["AlternateFlag"]:
            click.echo(Back.LIGHTWHITE_EX + "-->" + rich_proc +
                       Style.RESET_ALL)
        else:
            click.echo("-->" + rich_proc)
    else:
        if config.gParamDict["AlternateFlag"]:
            click.echo(Back.LIGHTWHITE_EX + "==>" + rich_proc +
                       Style.RESET_ALL)
        else:
            click.echo("==>" + rich_proc)
    config.gParamDict["AlternateFlag"] = not config.gParamDict["AlternateFlag"]


def one_file_ufn(f_path=""):
    subdir, file = os.path.split(f_path)
    new_name = file
    # all whitespace replace by sep_char
    new_name = process_white_space(f_path=new_name)
    # replace characters by defined Dictionary
    new_name = replace_char(f_path=new_name)
    # Capwords only when word in wordsSet
    new_name = process_word(f_path=new_name)
    # Pretty Terminology
    new_name = process_terminology(f_path=new_name)
    # process Head and Tail
    new_name = process_head_tail(f_path=new_name)
    # ascii head
    new_name = asc_head(new_name) + new_name

    # Create full path
    new_path = os.path.join(subdir, new_name)
    if not config.gParamDict["dry_run"]:
        # Create Or Update File Name Change Record and Save to File
        # then rename file name
        if new_name != file:
            utils.log_to_db(cur_name=file, new_name=new_name)
            os.rename(f_path, new_path)
    if new_name != file:
        out_info(file, new_name)


def one_dir_ufn(tgt_path):
    for subdir, dirs, files in depth_walk(
            top_path=tgt_path, max_depth=config.gParamDict["max_depth"]):
        for file in files:
            f_path = os.path.join(subdir, file)
            if is_hidden(f_path):
                continue
            if not os.path.isfile(f_path):
                continue
            one_file_ufn(f_path=f_path)


def one_file_rbk(f_path=""):
    subdir, file = os.path.split(f_path)
    new_name = utils.used_name_lookup(file)
    if new_name is None:
        return None
    # Create full path
    new_path = os.path.join(subdir, new_name)
    if not config.gParamDict["dry_run"]:
        # Create Or Update File Name Change Record and Save to File
        # then rename file name
        if new_name != file:
            os.rename(f_path, new_path)
    if new_name != file:
        out_info(file, new_name)


def one_dir_rbk(tgt_path):
    for subdir, dirs, files in depth_walk(
            top_path=tgt_path, max_depth=config.gParamDict["max_depth"]):
        for file in files:
            f_path = os.path.join(subdir, file)
            if is_hidden(f_path):
                continue
            if not os.path.isfile(f_path):
                continue
            one_file_rbk(f_path=f_path)


@click.group(cls=DefaultGroup, default="ufn", default_if_no_args=True)
def cli():
    pass


@cli.command(context_settings={"ignore_unknown_options": True})
@click.argument("path", required=False, type=click.Path(exists=True), nargs=-1)
@click.option("--max_depth",
              default=1,
              type=int,
              help="Set travel directory tree with max depth.",
              show_default=True)
@click.option("--dry_run",
              default=True,
              type=bool,
              help="If dry_run is True will not change file name.",
              show_default=True)
def rbk(path, max_depth, dry_run):
    if not path:
        config.gParamDict["path"] = ["."]
    else:
        config.gParamDict["path"] = path
    if str(max_depth).isnumeric():
        config.gParamDict["max_depth"] = int(str(max_depth))
    else:
        config.gParamDict["max_depth"] = 1
    config.gParamDict["dry_run"] = dry_run
    for pth in config.gParamDict["path"]:
        if os.path.isfile(pth):
            one_file_rbk(pth)
        elif os.path.isdir(pth):
            one_dir_rbk(pth)
        else:
            click.echo(f"{Fore.RED}Not valid:{pth}{Fore.RESET}")

    if config.gParamDict["dry_run"]:
        click.echo("*" * 79)
        click.echo(
            "In order to take effect,run the CLI add option '--dry_run False'")


@cli.command(context_settings={"ignore_unknown_options": True})
@click.argument("path", required=False, type=click.Path(exists=True), nargs=-1)
@click.option("--max_depth",
              default=1,
              type=int,
              help="Set travel directory tree with max depth.",
              show_default=True)
@click.option(
    "--exclude",
    default="",
    help="Exclude all files in exclude path.Not valid in current version.",
    show_default=True)
@click.option("--dry_run",
              default=True,
              type=bool,
              help="If dry_run is True will not change file name.",
              show_default=True)
def ufn(path, max_depth, exclude, dry_run):
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
    config.gParamDict["exclude"] = exclude
    config.gParamDict["dry_run"] = dry_run

    for pth in config.gParamDict["path"]:
        if os.path.isfile(pth):
            one_file_ufn(pth)
        elif os.path.isdir(pth):
            one_dir_ufn(pth)
        else:
            click.echo(f"{Fore.RED}Not valid:{pth}{Fore.RESET}")

    if config.gParamDict["dry_run"]:
        click.echo("*" * 79)
        click.echo(
            "In order to take effect,run the CLI add option '--dry_run False'")

    return 0


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
        cli()
    finally:
        colorama.deinit()

# TODO: not input file path when in processing only file path exclude ext is
#       enough

# TODO: add verify before change take effect
# TODO: undo default not include extension or manually set include extension
# TODO: support interactive operation
# TODO: display config data
# TODO: support edit config data
# TODO: display total summary
# TODO: display progress bar at bottom ...
