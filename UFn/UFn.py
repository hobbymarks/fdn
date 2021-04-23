#!/usr/bin/env python
import difflib
import hashlib
import collections
import os
import pickle
from datetime import datetime
import string
import sys
from unidecode import unidecode
import click
from rich.console import Console
from rich.theme import Theme
import re


# TODO: split to standalone file
class FileNameLog:
    def __init__(self, file_path="", md5_value=""):
        """

        :param file_path:
        :type file_path:
        :param md5_value:
        :type md5_value:
        """
        if not md5_value:
            assert os.path.exists(file_path)
            with open(file_path, "rb") as fh:
                data = fh.read()
                md5_value = hashlib.md5(data).hexdigest()
        # TODO:should add version info for save data structure
        self.md5_value = str(md5_value)
        self._currentName = os.path.basename(file_path)
        self._nameRecord = collections.OrderedDict()

    def change_file_name(self, new_name="", stamp=""):
        """Change file name

        This function will add the operation 'Change File Name' to record

        :param new_name: String,new file name
        :type new_name:
        :param stamp:
        :type stamp:
        :return:
        :rtype:
        """
        assert new_name
        if not stamp:
            stamp = datetime.now().strftime("%Y%m%d%H%M%S%f")
        self._nameRecord[stamp] = self._currentName
        self._currentName = new_name

    def get_history(self):
        """

        :return:
        :rtype:
        """
        return {
            self.md5_value: {
                "currentName": self._currentName,
                "nameRecord": self._nameRecord
            }
        }


def create_name_log(file_path="", md5_value=""):
    """

    :param file_path:
    :type file_path:
    :param md5_value:
    :type md5_value:
    :return:
    :rtype:
    """
    global gParamDict
    record_path = gParamDict["record_path"]
    if not md5_value:
        assert os.path.isfile(file_path)
        with open(file_path, "rb") as fh:
            data = fh.read()
            md5_value = hashlib.md5(data).hexdigest()
    file_record_path = os.path.join(record_path, str(md5_value) + "_HRd.pkl")
    if os.path.isfile(file_record_path):
        with open(file_record_path, "rb") as fh:
            rd = pickle.load(fh)
        return rd
    else:
        assert os.path.isfile(file_path)
        return FileNameLog(file_path)


def is_hidden(path):
    """

    :param path:
    :type path:
    :return:
    :rtype:
    """
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
    """

    :param s:
    :type s:
    :return:
    :rtype:
    """
    global gParamDict
    keep_original_list = gParamDict["KeepOriginalList"]
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
    """

    :param file_str:
    :type file_str:
    :return:
    :rtype:
    """
    global gParamDict
    char_dict = gParamDict["CharDictionary"]
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
            gParamDict["record_list"].append("".join(new_word_list) + ext)
            word_list = new_word_list
    return "".join(word_list) + ext


def process_head_tail_sep_char(file_str=""):
    """

    :param file_str:
    :type file_str:
    :return:
    :rtype:
    """
    global gParamDict
    sep_char = gParamDict["sep_char"]
    root, ext = os.path.splitext(file_str)
    if root.startswith(sep_char):
        root = sep_char.join(root.split(sep_char)[1:])
        gParamDict["record_list"].append(root + ext)
    if root.endswith(sep_char):
        root = sep_char.join(root.split(sep_char)[0:-1])
        gParamDict["record_list"].append(root + ext)
    return root + ext


def process_head_tail(file_str=""):
    """

    :param file_str:
    :type file_str:
    :return:
    :rtype:
    """
    global gParamDict
    assert file_str
    new_name = process_head_tail_sep_char(file_str=file_str)
    # Capitalize The First Letter
    if new_name[0].islower():
        new_name = new_name[0].upper() + new_name[1:]
        gParamDict["record_list"].append(new_name)

    return new_name


def process_white_space(file_str=""):
    """

    :param file_str:
    :type file_str:
    :return:
    :rtype:
    """
    global gParamDict
    sep_char = gParamDict["sep_char"]
    root, ext = os.path.splitext(file_str)
    new_name = sep_char.join(root.split()) + ext
    if new_name != file_str:
        gParamDict["record_list"].append(new_name)
    return new_name


def check_starts_with_terminology(word=""):
    """

    :param word:
    :type word:
    :return:
    :rtype:
    """
    global gParamDict
    term_dict = gParamDict["TerminologyDictionary"]
    for key in term_dict.keys():
        if word.lower().startswith(key):
            return key
    return None


def process_terminology(file_str=""):
    """

    :param file_str:
    :type file_str:
    :return:
    :rtype:
    """
    global gParamDict

    root, ext = os.path.splitext(file_str)
    word_list = root.split(gParamDict["sep_char"])
    new_word_list = []
    term_dict = gParamDict["TerminologyDictionary"]
    sep_char = gParamDict["sep_char"]
    for word in word_list:
        if word.lower() in term_dict.keys():
            new_word_list.append(term_dict[word.lower()])
        elif key := check_starts_with_terminology(word):
            new_word = term_dict[key] + word[len(key):]
            new_word_list.append(new_word)
        else:
            new_word_list.append(word)
    if new_word_list != word_list:
        gParamDict["record_list"].append(sep_char.join(new_word_list) + ext)
    return sep_char.join(new_word_list) + ext


def asc_head(file_str=""):
    """

    :param file_str:file name string
    :type file_str:string
    :return:
    :rtype:string
    """
    global gParamDict
    lmt_len = gParamDict["asc_len"]
    sep_char = gParamDict["sep_char"]
    head_chars = gParamDict["head_chars"]
    if file_str[0] in head_chars:
        return ""
    word = file_str.split(sep_char)[0]
    if len(word) > lmt_len:
        word = word[0:lmt_len]
    new_word = ""
    for c in word:
        if not c in head_chars:
            new_word += c
        else:
            break
    return "".join([elm[0] for elm in unidecode(new_word).split()])


def process_word(file_str=""):
    """

    :param file_str:
    :type file_str:
    :return:
    :rtype:
    """
    global gParamDict
    sep_char = gParamDict["sep_char"]
    root, ext = os.path.splitext(file_str)
    word_list = root.split(sep_char)
    new_word_list = []
    word_set = gParamDict["LowerCaseWordSet"]
    for word in word_list:
        if word.lower() in word_set:
            new_word_list.append(string.capwords(word))
        else:
            new_word_list.append(word)

    return sep_char.join(new_word_list) + ext


def depth_walk(top_path, top_down=True, follow_links=False, max_depth=1):
    """

    :param top_path:
    :type top_path:
    :param top_down:
    :type top_down:
    :param follow_links:
    :type follow_links:
    :param max_depth:
    :type max_depth:
    :return:
    :rtype:
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


def rich_style(org_str="", proc_str=""):
    """

    :param org_str:
    :type org_str:
    :param proc_str:
    :type proc_str:
    :return:
    :rtype:
    """
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


def one_file_ufn(file_str=""):
    """

    :param file_str:
    :type file_str:
    :return:
    :rtype:
    """
    global gParamDict
    new_name = file_str
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

    return new_name


def one_dir_uf(tgt_path, console, style):
    """

    :param tgt_path:
    :type tgt_path:
    :param console:
    :type console:
    :param style:
    :type style:
    :return:
    :rtype:
    """
    global gParamDict
    for subdir, dirs, files in depth_walk(top_path=tgt_path,
                                          max_depth=gParamDict["max_depth"]):
        for file in files:
            #             if not os.path.isfile(file):
            #                 continue
            gParamDict["record_list"] = []
            old_path = os.path.join(subdir, file)
            if is_hidden(old_path):
                continue
            new_name = one_file_ufn(file_str=file)
            # Create full path
            new_path = os.path.join(subdir, new_name)
            if not gParamDict["dry_run"]:
                # Create Or Update File Name Change Record and Save to File
                # then rename file name
                if new_name != file:
                    history_record = create_name_log(file_path=old_path)
                    with open(
                            os.path.join(
                                gParamDict["record_path"],
                                str(history_record.md5_value) + "_HRd.pkl"),
                            "wb") as fh:
                        history_record.change_file_name(new_name=new_name)
                        pickle.dump(history_record, fh)
                    os.rename(old_path, new_path)

            if new_name != file:
                rich_org, rich_proc = rich_style(file, new_name)
                console.print(" " * 3 + rich_org, style=style)
                if not gParamDict["simple"]:
                    for fName in gParamDict["record_list"]:
                        console.print("---" + fName, style=style)
                if gParamDict["dry_run"]:
                    console.print("-->" + rich_proc, style=style)
                else:
                    console.print("==>" + rich_proc, style=style)


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
@click.option("--simple",
              default=True,
              type=bool,
              help="If simple is True Only print changed file name.",
              show_default=True)
def ufn(path, max_depth, exclude, dry_run, simple):
    """Files in PATH will be changed file names unified.
    
    You can direct set path such as UFn.py path ...
    """
    global gParamDict
    if not path:
        gParamDict["path"] = "."
    else:
        gParamDict["path"] = path
    if str(max_depth).isnumeric():
        gParamDict["max_depth"] = int(str(max_depth))
    else:
        gParamDict["max_depth"] = 1
    gParamDict["exclude"] = exclude
    gParamDict["dry_run"] = dry_run
    gParamDict["simple"] = simple

    tgt_path = ""

    if gParamDict["path"]:
        tgt_path = gParamDict["path"]
    else:
        if not os.path.isdir(gParamDict["path"]):
            click.echo("%s is not valid path.")
            return -1
        tgt_path = gParamDict["path"]

    with open(os.path.join(gParamDict["data_path"], "CharDictionary.pkl"),
              "rb") as fh:
        gParamDict["CharDictionary"] = pickle.load(fh)
    with open(
            os.path.join(gParamDict["data_path"], "TerminologyDictionary.pkl"),
            "rb") as fh:
        gParamDict["TerminologyDictionary"] = pickle.load(fh)
    with open(os.path.join(gParamDict["data_path"], "LowerCaseWordSet.pkl"),
              "rb") as fh:
        gParamDict["LowerCaseWordSet"] = pickle.load(fh)
    with open(os.path.join(gParamDict["data_path"], "KeepOriginalList.pkl"),
              "rb") as fh:
        gParamDict["KeepOriginalList"] = pickle.load(fh)
    for path in tgt_path:
        one_dir_uf(path, console, style)

    if gParamDict["dry_run"]:
        console.print("*" * 80)
        console.print(
            "In order to take effect,run the CLI add option '--dry_run False'")

    return 0


if __name__ == "__main__":
    console = Console(width=240, theme=Theme(inherit=False))
    style = "black on white"

    if (sys.version_info.major, sys.version_info.minor) < (3, 8):
        console.print(
            f"current Version is {sys.version},\n Please upgrade to >= 3.8.")
        sys.exit()
    scriptDirPath = os.path.dirname(os.path.realpath(__file__))

    gParamDict = {
        "console": (console, style),
        "sep_char": "_",
        "data_path": os.path.join(scriptDirPath, "data"),
        "record_path": os.path.join(scriptDirPath, "data", "hRdDir"),
        "record_list": [],
        "asc_len": 5,
        "head_chars": string.ascii_letters + string.digits + string.punctuation
    }
    ############################################################################
    ufn()

# TODO: add verify before change take effect
# TODO: support undo operation
