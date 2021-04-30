import hashlib
import os

# From This Project
import config
import ucrypt
from udb import UDB


def sha2_id(s):
    if (not s) or (not type(s) is str):
        return None
    return hashlib.sha256(s.encode("UTF-8")).hexdigest()


def used_name_lookup(cur_name, db_path="", latest=True):
    if not db_path:
        db_path = os.path.join(config.gParamDict["record_path"], "rd.db")
    _cur_id = sha2_id(cur_name)
    db = UDB(db_path)
    df = db.checkout_rd(_cur_id)
    db.close()
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


def log_to_db(cur_name, new_name, db_path=""):
    if not db_path:
        db_path = os.path.join(config.gParamDict["record_path"], "rd.db")
    _cur_id = sha2_id(cur_name)
    _new_id = sha2_id(new_name)
    _cur_crypt = ucrypt.encrypt_b64_str(cur_name, new_name)
    db = UDB(db_path)
    db.insert_rd(_new_id, _cur_id, _cur_crypt)
    db.close()
