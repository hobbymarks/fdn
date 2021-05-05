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
except ModuleNotFoundError:
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


def mask_ruw(s):
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


def replace_char(s):
    """
    Replace char in s ,when the char in "BeReplacedCharDictionary"
    :param s:string,if "BeReplacedChar" in s ,then will be replaced
    :return:string,be replaced
    """
    c_dict = config.gParamDict["BeReplacedCharDictionary"]
    sep_c = config.gParamDict["SeparatorChar"]
    re_cns = re.compile(
        f"[{sep_c}]+")  # To recognize continuous separator char
    word_list, mask_list = mask_ruw(s)

    new_word_list = []
    for word, mask in zip(word_list, mask_list):
        if mask:
            new_word_list.append(word)
        else:
            # TODO: May Be Not Highest Efficiency
            for c_key, c_value in c_dict.items():
                if c_key in word:  # If need replace ,then go ...
                    word = word.replace(c_key, c_value)
            word = re_cns.sub(sep_c, word)  # When continuous separator,remain
            # only one
            new_word_list.append(word)
    return "".join(new_word_list)


def process_head_tail_sep(s):
    """
    Delete the beginning and the end separator in string.
    :param s: string
    :return: string
    """
    sep_char = config.gParamDict["SeparatorChar"]
    if s.startswith(sep_char):
        s = sep_char.join(s.split(sep_char)[1:])
    if s.endswith(sep_char):
        s = sep_char.join(s.split(sep_char)[0:-1])
    return s


def process_head_tail(s):
    """

    :param s:
    :return:
    """
    root = process_head_tail_sep(s)
    # Capitalize The First Letter
    if root[0].islower():
        root = root[0].upper() + root[1:]
    return root


def process_white_space(s):
    """
    All white space will be replaced by Separator Char.
    :param s:string,input string
    :return:string,all whitespace in input string will be replaced by
    Separator Char then return the processed input string
    """
    sep_char = config.gParamDict["SeparatorChar"]
    return sep_char.join(s.split())


def starts_with_terminology(s):
    """
    Check if the s startswith a terminology word.If the s startswith a
    terminology word then return the word lower case ,else return None.
    :param s:string,input
    :return:None or string
    """
    term_dict = config.gParamDict["TerminologyDictionary"]
    for key in term_dict.keys():
        if s.lower().startswith(key):
            return key
    return None


def process_terminology(s):
    """

    :param s:
    :return:
    """
    sep_char = config.gParamDict["SeparatorChar"]
    word_list = s.split(sep_char)
    term_dict = config.gParamDict["TerminologyDictionary"]
    new_word_list = []
    for word in word_list:
        if word.lower() in term_dict.keys():
            new_word_list.append(term_dict[word.lower()])
        elif key := starts_with_terminology(word):
            new_word = term_dict[key] + word[len(key):]
            new_word_list.append(new_word)
        else:
            new_word_list.append(word)
    return sep_char.join(new_word_list)


def asc_head(s):
    """
    If s not starts with ascii letters ,will create at most "ASCLen" length
    asc letters then return ,else return null string.
    :param s:string
    :return:string ,null string or at most "ASCLen" length ascii letters
    """
    lmt_len = config.gParamDict["ASCLen"]
    sep_char = config.gParamDict["SeparatorChar"]
    head_chars = config.gParamDict["HeadChars"]
    if s[0] in head_chars:
        return ""
    word = s.split(sep_char)[0]
    if len(word) > lmt_len:
        word = word[0:lmt_len]
    new_word = ""
    for c in word:
        if c not in head_chars:
            new_word += c
        else:
            break
    return "".join([elm[0] for elm in unidecode(new_word).split()])


def process_word(s):
    """

    :param s:
    :return:
    """
    sep_char = config.gParamDict["SeparatorChar"]
    word_list = s.split(sep_char)
    new_word_list = []
    word_set = config.gParamDict["LowerCaseWordSet"]
    for word in word_list:
        if word.lower() in word_set:
            new_word_list.append(string.capwords(word))
        else:
            new_word_list.append(word)
    return sep_char.join(new_word_list)


def depth_walk(top_path, top_down=True, follow_links=False, max_depth=1):
    """

    :param top_path:
    :param top_down:
    :param follow_links:
    :param max_depth:
    :return:
    """
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
    """

    :param original:
    :param processed:
    :return:
    """
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
        click.echo(Back.LIGHTWHITE_EX + " " * 3 + Style.RESET_ALL + rich_org)
    else:
        click.echo(" " * 3 + rich_org)
    if config.gParamDict["dry_run"]:
        if config.gParamDict["AlternateFlag"]:
            click.echo(Back.LIGHTWHITE_EX + "-->" + Style.RESET_ALL +
                       rich_proc)
        else:
            click.echo("-->" + rich_proc)
    else:
        if config.gParamDict["AlternateFlag"]:
            click.echo(Back.LIGHTWHITE_EX + "==>" + Style.RESET_ALL +
                       rich_proc)
        else:
            click.echo("==>" + rich_proc)
    config.gParamDict["AlternateFlag"] = not config.gParamDict["AlternateFlag"]


def one_file_ufn(f_path):
    subdir, file = os.path.split(f_path)
    root, ext = os.path.splitext(file)
    # all whitespace replace by sep_char
    root = process_white_space(root)
    # replace characters by defined Dictionary
    root = replace_char(root)
    # Capwords only when word in wordsSet
    root = process_word(root)
    # Pretty Terminology
    root = process_terminology(root)
    # process Head and Tail
    root = process_head_tail(root)
    # Add ascii head at the beginning
    new_name = asc_head(root) + root + ext

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


def one_file_rbk(f_path):
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

# TODO: add verify before change take effect
# TODO: undo default not include extension or manually set include extension
# TODO: support interactive operation
# TODO: display config data
# TODO: support edit config data
# TODO: display total summary
# TODO: display progress bar at bottom ...

# TODO: support regular expression input path or directory as path argument
# TODO: support directory change name only by interactive way
