#!/usr/bin/env python
import difflib
import os
from pathlib import Path
import pickle
import re
import string
import sys
# From Third party
import click
from unidecode import unidecode
# From This Project
import config
import utils


def is_hidden(path):
    if os.name == "nt":
        import win32api
        import win32con
    if os.name == "nt":
        attribute = win32api.GetFileAttributes(path)
        return attribute & (win32con.FILE_ATTRIBUTE_HIDDEN
                            | win32con.FILE_ATTRIBUTE_SYSTEM)
    else:
        return os.path.basename(path).startswith('.')  # linux, osx


def mask_original(s=""):
    keep_original_list = config.gParamDict["KeepOriginalList"]
    re_str = "|".join([re.escape(sepStr) for sepStr in keep_original_list])
    word_list = re.split(f'({re_str})', s)
    mask_list = []
    for elm in word_list:
        if elm in keep_original_list:
            mask_list.extend([True])
        else:
            mask_list.extend([False])

    return word_list, mask_list


def replace_char(file_str=""):
    char_dict = config.gParamDict["CharDictionary"]
    root, ext = os.path.splitext(file_str)
    word_list, mask_list = mask_original(root)
    for key, value in char_dict.items():
        new_word_list = []
        for word, mask in zip(word_list, mask_list):
            if not mask:
                new_word_list.append(word)
        if key in "".join(new_word_list):  # If need to replace ,then go ...
            new_word_list = []
            for word, mask in zip(word_list, mask_list):
                if (key in word) and (not mask):
                    new_word_list.append(word.replace(key, value))
                else:
                    new_word_list.append(word)
            config.gParamDict["record_list"].append("".join(new_word_list) +
                                                    ext)
            word_list = new_word_list
    return "".join(word_list) + ext


def process_head_tail_sep_char(file_str=""):
    sep_char = config.gParamDict["sep_char"]
    root, ext = os.path.splitext(file_str)
    if root.startswith(sep_char):
        root = sep_char.join(root.split(sep_char)[1:])
        config.gParamDict["record_list"].append(root + ext)
    if root.endswith(sep_char):
        root = sep_char.join(root.split(sep_char)[0:-1])
        config.gParamDict["record_list"].append(root + ext)
    return root + ext


def process_head_tail(file_str=""):
    assert file_str
    new_name = process_head_tail_sep_char(file_str=file_str)
    # Capitalize The First Letter
    if new_name[0].islower():
        new_name = new_name[0].upper() + new_name[1:]
        config.gParamDict["record_list"].append(new_name)

    return new_name


def process_white_space(file_str=""):
    sep_char = config.gParamDict["sep_char"]
    root, ext = os.path.splitext(file_str)
    new_name = sep_char.join(root.split()) + ext
    if new_name != file_str:
        config.gParamDict["record_list"].append(new_name)
    return new_name


def check_starts_with_terminology(word=""):
    term_dict = config.gParamDict["TerminologyDictionary"]
    for key in term_dict.keys():
        if word.lower().startswith(key):
            return key
    return None


def process_terminology(file_str=""):
    root, ext = os.path.splitext(file_str)
    word_list = root.split(config.gParamDict["sep_char"])
    new_word_list = []
    term_dict = config.gParamDict["TerminologyDictionary"]
    sep_char = config.gParamDict["sep_char"]
    for word in word_list:
        if word.lower() in term_dict.keys():
            new_word_list.append(term_dict[word.lower()])
        elif key := check_starts_with_terminology(word):
            new_word = term_dict[key] + word[len(key):]
            new_word_list.append(new_word)
        else:
            new_word_list.append(word)
    if new_word_list != word_list:
        config.gParamDict["record_list"].append(
            sep_char.join(new_word_list) + ext)
    return sep_char.join(new_word_list) + ext


def asc_head(file_str=""):
    lmt_len = config.gParamDict["asc_len"]
    sep_char = config.gParamDict["sep_char"]
    head_chars = config.gParamDict["head_chars"]
    if file_str[0] in head_chars:
        return ""
    word = file_str.split(sep_char)[0]
    if len(word) > lmt_len:
        word = word[0:lmt_len]
    new_word = ""
    for c in word:
        if c not in head_chars:
            new_word += c
        else:
            break
    return "".join([elm[0] for elm in unidecode(new_word).split()])


def process_word(file_str=""):
    sep_char = config.gParamDict["sep_char"]
    root, ext = os.path.splitext(file_str)
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


def rich_style(org_str="", proc_str=""):
    rich_org = rich_proc = ""
    rich_org_dif_pos = rich_proc_dif_pos = 0
    for match in difflib.SequenceMatcher(a=org_str,
                                         b=proc_str).get_matching_blocks():
        if rich_org_dif_pos < match.a:
            rich_org += "[bold red]" + org_str[
                rich_org_dif_pos:match.a].replace(
                    " ", "▯") + "[/bold red]" + org_str[match.a:match.a +
                                                        match.size]
            rich_org_dif_pos = match.a + match.size
        else:
            rich_org += org_str[match.a:match.a + match.size]
            rich_org_dif_pos = match.a + match.size

        if rich_proc_dif_pos < match.b:
            rich_proc += "[bold green]" + proc_str[
                rich_proc_dif_pos:match.b].replace(
                    " ", "▯") + "[/bold green]" + proc_str[match.b:match.b +
                                                           match.size]
            rich_proc_dif_pos = match.b + match.size
        else:
            rich_proc += proc_str[match.b:match.b + match.size]
            rich_proc_dif_pos = match.b + match.size

    return rich_org, rich_proc


def one_file_ufn(file_path=""):
    config.gParamDict["record_list"] = []
    console, style = config.gParamDict["console"]
    subdir, file = os.path.split(file_path)
    new_name = file
    # all whitespace replace by sep_char
    new_name = process_white_space(file_str=new_name)
    # replace characters by defined Dictionary
    new_name = replace_char(file_str=new_name)
    # Capwords only when word in wordsSet
    new_name = process_word(file_str=new_name)
    # Pretty Terminology
    new_name = process_terminology(file_str=new_name)
    # process Head and Tail
    new_name = process_head_tail(file_str=new_name)
    # ascii head
    new_name = asc_head(new_name) + new_name

    # Create full path
    new_path = os.path.join(subdir, new_name)
    if not config.gParamDict["dry_run"]:
        # Create Or Update File Name Change Record and Save to File
        # then rename file name
        if new_name != file:
            utils.log_to_file(cur_name=file, new_name=new_name)
            os.rename(file_path, new_path)

    if new_name != file:
        rich_org, rich_proc = rich_style(file, new_name)
        console.print(" " * 3 + rich_org, style=style)
        if not config.gParamDict["simple"]:
            for fName in config.gParamDict["record_list"]:
                console.print("---" + fName, style=style)
        if config.gParamDict["dry_run"]:
            console.print("-->" + rich_proc, style=style)
        else:
            console.print("==>" + rich_proc, style=style)


def one_dir_ufn(tgt_path):
    for subdir, dirs, files in depth_walk(
            top_path=tgt_path, max_depth=config.gParamDict["max_depth"]):
        for file in files:
            f_path = os.path.join(subdir, file)
            if is_hidden(f_path):
                continue
            if not os.path.isfile(f_path):
                continue
            one_file_ufn(file_path=f_path)


@click.command(context_settings={"ignore_unknown_options": True})
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
@click.option("--force_run",
              default=False,
              type=bool,
              help="If force_run is True will change file name forcefully.",
              show_default=True)
@click.option("--simple",
              default=True,
              type=bool,
              help="If simple is True Only print changed file name.",
              show_default=True)
def ufn(path, max_depth, exclude, dry_run, force_run, simple):
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
    config.gParamDict["force_run"] = force_run
    config.gParamDict["simple"] = simple

    with open(
            os.path.join(config.gParamDict["data_path"], "CharDictionary.pkl"),
            "rb") as fh:
        config.gParamDict["CharDictionary"] = pickle.load(fh)
    with open(
            os.path.join(config.gParamDict["data_path"],
                         "TerminologyDictionary.pkl"), "rb") as fh:
        config.gParamDict["TerminologyDictionary"] = pickle.load(fh)
    with open(
            os.path.join(config.gParamDict["data_path"],
                         "LowerCaseWordSet.pkl"), "rb") as fh:
        config.gParamDict["LowerCaseWordSet"] = pickle.load(fh)
    with open(
            os.path.join(config.gParamDict["data_path"],
                         "KeepOriginalList.pkl"), "rb") as fh:
        config.gParamDict["KeepOriginalList"] = pickle.load(fh)
    tgt_path = config.gParamDict["path"]
    for path in tgt_path:
        if os.path.isfile(path):
            one_file_ufn(path)
        elif os.path.isdir(path):
            one_dir_ufn(path)
        else:
            print(f"Not valid:{path}")

    if config.gParamDict["dry_run"]:
        console.print("*" * 80)
        console.print(
            "In order to take effect,run the CLI add option '--dry_run False'")

    return 0


if __name__ == "__main__":
    console, style = config.gParamDict["console"]

    if (sys.version_info.major, sys.version_info.minor) < (3, 8):
        console.print(
            f"current Version is {sys.version},\n Please upgrade to >= 3.8.")
        sys.exit()
    scriptDirPath = os.path.dirname(os.path.realpath(__file__))
    config.gParamDict["data_path"] = os.path.join(scriptDirPath, "data")
    config.gParamDict["record_path"] = os.path.join(scriptDirPath, "data",
                                                    "rd")
    Path(config.gParamDict["record_path"]).mkdir(parents=True, exist_ok=True)
    ############################################################################
    ufn()

# TODO: support undo operation
# TODO: support one file undo change name

# TODO: add verify before change take effect
# TODO: undo default not include extension or manually set include extension
