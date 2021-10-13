import os
import re
import string
import sys
from pathlib import Path
from typing import List, Optional, Tuple, Union

# From Third Party
import click
from colorama import Back
from fdn.fdnlib import ucrypt, utils
# From This Project
from fdn.fdnlib.fdncfg import gParamDict as ugPD
from fdn.fdnlib.udb import FDNDB
from unidecode import unidecode


def _replace_char(s: str) -> str:
    """
    Replace char in s ,when the char in "BeReplacedCharDictionary"
    :param s:string,if "BeReplacedChar" in s ,then will be replaced
    :return:string,be replaced
    """
    def _mask_ruw(_s: str) -> Tuple[List[str], List[bool]]:
        """
        Mask a input string by RemainUnchangedWordList value
        :param _s: string,input string to be masked
        :return: word list and mask list
        """
        ruw_list = ugPD["RemainUnchangedWordList"]
        re_str = "|".join([re.escape(ruw) for ruw in ruw_list])
        w_list = re.split(f"({re_str})", _s)
        m_list = []
        for elm in w_list:
            if elm in ruw_list:
                m_list.append(True)
            else:
                m_list.append(False)
        return w_list, m_list

    c_dict = ugPD["BeReplacedCharDictionary"]
    sep_c = ugPD["SeparatorChar"]
    re_cns = re.compile(
        f"[{sep_c}]+")  # To recognize continuous separator char
    word_list, mask_list = _mask_ruw(s)

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


def _del_head_tail_sep(s: str) -> str:
    """
    Delete the beginning and the end separator in string.
    :param s: string
    :return: string
    """
    sep_char = ugPD["SeparatorChar"]
    if s.startswith(sep_char):
        s = sep_char.join(s.split(sep_char)[1:])
    if s.endswith(sep_char):
        s = sep_char.join(s.split(sep_char)[0:-1])
    return s


def _process_head_tail(s: str) -> str:
    # Capitalize The First Letter
    # if s[0].islower():
    #     return s[0].upper() + s[1:]
    return s


def _del_white_space(s: str) -> str:
    """
    All white space will be replaced by Separator Char.
    :param s:string,input string
    :return:string,all whitespace in input string will be replaced by
    Separator Char then return the processed input string
    """
    sep_char = ugPD["SeparatorChar"]
    return sep_char.join(s.split())


def _process_terminology(s: str) -> str:
    """

    :param s:
    :return:
    """
    def _swt(_s: str) -> Optional[str]:
        """
        Check if the s startswith a terminology word.If the s startswith a
        terminology word then return the word lower case ,else return None.
        :param _s:string,input
        :return:None or string
        """
        t_d = ugPD["TerminologyDictionary"]
        for k in t_d.keys():
            if _s.lower().startswith(k):
                return k
        return None

    sep_char = ugPD["SeparatorChar"]
    word_list = s.split(sep_char)
    term_dict = ugPD["TerminologyDictionary"]
    new_word_list = []
    for word in word_list:
        if word.lower() in term_dict.keys():
            new_word_list.append(term_dict[word.lower()])
        elif key := _swt(word):
            new_word = term_dict[key] + word[len(key):]
            new_word_list.append(new_word)
        else:
            new_word_list.append(word)
    return sep_char.join(new_word_list)


def _asc_head(s: str) -> str:
    """
    If s not starts with ascii letters ,will create at most "ASCLen" length
    asc letters then return ,else return null string.
    :param s:string
    :return:string ,null string or at most "ASCLen" length ascii letters
    """
    lmt_len = ugPD["ASCLen"]
    sep_char = ugPD["SeparatorChar"]
    head_chars = ugPD["HeadChars"]
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
    return "".join([elm[0].upper() for elm in unidecode(new_word).split()])


def _process_word(s: str) -> str:
    sep_char = ugPD["SeparatorChar"]
    word_list = s.split(sep_char)
    new_word_list = []
    word_set = ugPD["LowerCaseWordSet"]
    for word in word_list:
        if word.lower() in word_set:
            new_word_list.append(string.capwords(word))
        else:
            new_word_list.append(word)
    return sep_char.join(new_word_list)


def _confirm(p_i: str = "") -> str:
    v = click.prompt(f"{p_i}\nPlease confirm(y/n/A/q)",
                     type=click.Choice(
                         ["y", "yes", "n", "no", "A", "all", "q", "quit"]),
                     show_choices=False,
                     default=ugPD["latest_confirm"])

    return utils.unify_confirm(v)


def _in_place(p_i: str = "") -> bool:
    if ugPD["all_in_place_flag"]:
        return True
    if ugPD["need_confirmation_flag"]:
        if (c := _confirm(p_i)) == utils.unify_confirm("yes"):
            ugPD["latest_confirm"] = c
            return True
        elif c == utils.unify_confirm("no"):
            ugPD["latest_confirm"] = c
            return False
        elif c == utils.unify_confirm("all"):
            ugPD["latest_confirm"] = c
            ugPD["all_in_place_flag"] = True
            return True
        elif c == utils.unify_confirm("quit"):
            ugPD["latest_confirm"] = c
            sys.exit()  # TODO: roughly process ...
    else:
        if ugPD["in_place_flag"]:
            return True
        else:
            return False


def _out_info(file: str, new_name: str, take_effect: bool = False) -> None:
    def _b_d(s="", b_d=""):
        if ugPD["enhanced_display_flag"]:
            return b_d + s + Back.RESET
        else:
            return s

    rich_org, rich_proc = utils.rich_style(
        file,
        new_name,
        pretty=ugPD["pretty_flag"],
        enhanced_display=ugPD["enhanced_display_flag"])
    if ugPD["AlternateFlag"]:
        click.echo(" " * 3 + (_b_d(rich_org, Back.WHITE)))
    else:
        click.echo(" " * 3 + (_b_d(rich_org, Back.LIGHTWHITE_EX)))
    if take_effect:
        if ugPD["AlternateFlag"]:
            click.echo("==>" + (_b_d(rich_proc, Back.WHITE)))
        else:
            click.echo("==>" + (_b_d(rich_proc, Back.LIGHTWHITE_EX)))
    else:
        if ugPD["AlternateFlag"]:
            click.echo("-->" + (_b_d(rich_proc, Back.WHITE)))
        else:
            click.echo("-->" + (_b_d(rich_proc, Back.LIGHTWHITE_EX)))
    ugPD["AlternateFlag"] = not ugPD["AlternateFlag"]


def type_matched(f_path: str) -> bool:
    if (ugPD["type"] == "file") and (os.path.isfile(f_path)):
        return True
    if (ugPD["type"] == "dir") and (os.path.isdir(f_path)):
        return True
    return False


def one_file_ufn(f_path: Path) -> None:
    if os.path.islink(f_path) != ugPD["is_link"]:
        return None

    subdir, file = os.path.split(f_path)
    root, ext = os.path.splitext(file)
    # all whitespace replace by sep_char
    root = _del_white_space(root)
    # replace characters by defined Dictionary
    root = _replace_char(root)
    # Capwords only when word in wordsSet
    root = _process_word(root)
    # process Head and Tail
    root = _process_head_tail(root)
    root = _del_head_tail_sep(root)  # special terminology may conflict
    # Pretty Terminology
    root = _process_terminology(root)
    # Add ascii head at the beginning
    new_name = _asc_head(root) + root + ext

    # Create full path
    new_path = os.path.join(subdir, new_name)

    if new_name == file:
        return None

    _ap = ugPD["absolute_path_flag"]
    _fp = ugPD["full_path_flag"]

    if _ap:
        ip = _in_place(os.path.abspath(f_path))
    elif _fp:
        ip = _in_place(str(f_path))
    else:
        ip = _in_place(file)
    if ip:
        if (os.path.exists(new_path)) and (not ugPD["overwrite_flag"]):
            click.echo(f"{Back.RED}Exist:{Back.RESET}"
                       f"{new_path if _fp else new_name}\n"
                       f"Skipped:{f_path if _fp else file}\n"
                       f"With option '-o' to enable overwrite.")
            return None

        else:
            log_to_db(cur_name=file, new_path=Path(new_path))
            os.rename(f_path, new_path)
    if _ap:
        _out_info(str(os.path.abspath(f_path)),
                  os.path.abspath(new_path),
                  take_effect=ip)
    elif _fp:
        _out_info(str(f_path), new_path, take_effect=ip)
    else:
        _out_info(file, new_name, take_effect=ip)

    ugPD["target_appeared"] = True


def one_dir_ufn(tgt_path: Path) -> None:
    for subdir, dirs, files in utils.depth_walk(top_path=tgt_path,
                                                max_depth=ugPD["max_depth"]):
        if ugPD["type"] == "file":
            for file in files:
                f_path = Path(os.path.join(subdir, file))
                if utils.is_hidden(f_path):
                    continue
                if os.path.isfile(f_path):
                    one_file_ufn(f_path)
        elif ugPD["type"] == "dir":
            for d in dirs:
                f_path = Path(os.path.join(subdir, d))
                if utils.is_hidden(f_path):
                    continue
                if os.path.isdir(f_path):
                    one_file_ufn(f_path)
    if ugPD["type"] == "dir":
        one_file_ufn(tgt_path)


def one_file_rbk(f_path: Path) -> None:
    if os.path.islink(f_path) != ugPD["is_link"]:
        return None

    subdir, file = os.path.split(f_path)
    if not (new_name := _used_name_lookup(file)):
        return None

    # Create full path
    new_path = os.path.join(subdir, new_name)

    if new_name == file:
        return None

    _ap = ugPD["absolute_path_flag"]
    _fp = ugPD["full_path_flag"]

    if _ap:
        ip = _in_place(os.path.abspath(f_path))
    elif _fp := ugPD["full_path_flag"]:
        ip = _in_place(str(f_path))
    else:
        ip = _in_place(file)
    if ip:
        if (os.path.exists(new_path)) and (not ugPD["overwrite_flag"]):
            click.echo(f"{new_path if _fp else new_name} exist.\n"
                       f"{f_path if _fp else file} Skipped."
                       f"You can with option '-o' to enable overwrite.")
            return None

        else:
            os.replace(f_path, new_path)  # os.place is atomic and match
            # cross platform :https://docs.python.org/dev/library/os.html
    if _ap:
        _out_info(str(os.path.abspath(f_path)),
                  os.path.abspath(new_path),
                  take_effect=ip)
    elif _fp:
        _out_info(str(f_path), new_path, take_effect=ip)
    else:
        _out_info(file, new_name, take_effect=ip)

    ugPD["target_appeared"] = True


def one_dir_rbk(tgt_path: Path) -> None:
    for subdir, dirs, files in utils.depth_walk(top_path=tgt_path,
                                                max_depth=ugPD["max_depth"]):
        if ugPD["type"] == "file":
            for file in files:
                f_path = Path(os.path.join(subdir, file))
                if utils.is_hidden(f_path):
                    continue
                one_file_rbk(f_path)
        elif ugPD["type"] == "dir":
            for d in dirs:
                f_path = Path(os.path.join(subdir, d))
                if utils.is_hidden(f_path):
                    continue
                one_file_rbk(f_path)
    if ugPD["type"] == "dir":
        one_file_rbk(tgt_path)


###############################################################################
def _used_name_lookup(cur_name: str,
                      db_path: Optional[Path] = None,
                      latest: bool = True) -> Union[None, str, List[str]]:
    if not db_path:
        db_path = ugPD["db_path"]
    _cur_id = utils.sha2_id(cur_name)
    db = FDNDB(db_path)
    df = db.checkout_rd(_cur_id)
    if "curCrypt" not in df.columns:
        return None
    rlt = list(df["curCrypt"])
    if len(rlt) == 0:
        return None
    if latest:
        return ucrypt.b64_str_decrypt(rlt[0], cur_name)
    else:
        return [
            ucrypt.b64_str_decrypt(elm, cur_name)
            for elm in list(dict.fromkeys(rlt))
        ]


def log_to_db(cur_name: str,
              new_path: Path,
              db_path: Optional[Path] = None) -> None:
    if not db_path:
        db_path = ugPD["db_path"]
    _f_p = os.path.abspath(new_path)
    _new = os.path.basename(_f_p)
    abs_dp = os.path.dirname(_f_p)
    _cur_id = utils.sha2_id(cur_name)
    _new_id = utils.sha2_id(_new)
    _cur_crypt = ucrypt.encrypt_b64_str(cur_name, _new)
    _sep_dp_id = "_|_".join([utils.sha2_id(d) for d in abs_dp.split(os.sep)])
    _abs_dp_id = utils.sha2_id(abs_dp)
    db = FDNDB(db_path)
    db.insert_rd(_new_id, _cur_id, _cur_crypt, _sep_dp_id, _abs_dp_id)
