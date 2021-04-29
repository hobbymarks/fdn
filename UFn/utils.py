from datetime import datetime
import hashlib
import os
import re
import sys
import pickle

# From Third Party
import click

# From This Project
import config
import ucrypt


def get_id(s=""):
    if (not s) or (not type(s) is str):
        return None
    return hashlib.sha1(s.encode("UTF-8")).hexdigest()


def get_id_enc(cur_name="", new_name=""):
    if (not cur_name) or (not new_name):
        return None, None
    id_nm = get_id(new_name)
    enc_str_cm = ucrypt.encrypt_b64_str(str_to_enc=cur_name,
                                        password_str=new_name)
    return id_nm, enc_str_cm


def get_stamp(enc_str=""):
    re_pattern = ".*_(\d{20}).pkl"
    re_match = re.match(re_pattern, enc_str)
    if re_match is not None:
        return re_match.groups()[0]
    return None


def load_log(rd_path=""):
    if not rd_path:
        click.echo(f"Parameters not set,please check again.")
        sys.exit()
    if not os.path.isfile(rd_path):
        click.echo(f"No file found:{rd_path},\nPlease check again.")
        sys.exit()
    try:
        with open(rd_path, "rb") as fh:
            rd_dict = pickle.load(fh)
            if type(rd_dict) is dict:
                return rd_dict
            return None

    except EOFError as e:
        click.echo(f"EOFError:{e}\n"
                   f"Please check again\n"
                   f"or delete {rd_path}\n "
                   f"or with option --force_run True")
        return None
    except pickle.UnpicklingError as e:
        click.echo(f"pickle.UnpicklingError:{e}\n"
                   f"Please check again\n"
                   f"or delete {rd_path}\n "
                   f"or with option --force_run True")
        return None


def scan_rd(rd_dir_path=""):
    stamp_rpath_dict = {}
    if not rd_dir_path:
        rd_dir_path = config.gParamDict["record_path"]
    for subdir, dirs, files in os.walk(rd_dir_path):
        for file in files:
            stamp = get_stamp(file)
            if stamp is None:
                continue
            stamp_rpath_dict[stamp] = os.path.join(subdir, file)
    return stamp_rpath_dict


def get_stamp_id_crypt(stamp_rpath_dict=None):
    stamp_id_crypt_dict = {}
    if not stamp_rpath_dict:
        stamp_rpath_dict = scan_rd()
    for stamp, rpath in stamp_rpath_dict.items():
        with open(rpath, "rb") as fh:
            rd_dict = pickle.load(fh)
            if not type(rd_dict) is dict:
                continue
            for key, value in rd_dict.items():
                stamp_id_crypt_dict[(stamp, key)] = value

    return stamp_id_crypt_dict


def stamp_id(password_str=""):
    stamp_id_list = []
    id_ = get_id(password_str)
    dict_ = config.gParamDict["stamp_id_crypt_dict"]
    for key, value in dict_.items():
        if id_ == key[1]:
            stamp_id_list.append(key)
    return stamp_id_list


def old_name(cur_name=""):
    stamp_id_list = stamp_id(cur_name)
    dict_ = config.gParamDict["stamp_id_crypt_dict"]
    stamp_nm_dict = {}
    for st_id in stamp_id_list:
        crypt = dict_[st_id]
        nm_str = ucrypt.b64_str_decrypt(enc_str=crypt, password_str=cur_name)
        if st_id[0] not in stamp_nm_dict.keys():
            stamp_nm_dict[st_id[0]] = nm_str
        else:
            click.echo(f"Skipping duplicate record found,{st_id}")

    return stamp_nm_dict


def log_to_file(cur_name="", new_name="", rd_path=""):
    id_nm, enc_str = get_id_enc(cur_name, new_name)
    if (not id_nm) or (not enc_str):
        click.echo(f"Parameters not set,please check again.")
        sys.exit()
    if not rd_path:
        stamp = datetime.now().strftime("%Y%m%d%H%M%S%f")
        rd_path = os.path.join(config.gParamDict["record_path"],
                               config.rd_prefix_str + "_" + stamp + ".pkl")
    with open(rd_path, "wb") as fh:
        pickle.dump({id_nm: enc_str}, fh, protocol=pickle.HIGHEST_PROTOCOL)
