from datetime import datetime
import hashlib
import os
import sys
import pickle
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


def load_log(rd_path=""):
    if not rd_path:
        print(f"Parameters not set,please check again.")
        sys.exit()
    if not os.path.isfile(rd_path):
        print(f"No file found:{rd_path},\nPlease check again.")
        sys.exit()
    try:
        with open(rd_path, "rb") as fh:
            rd_dict = pickle.load(fh)
            if type(rd_dict) is dict:
                return rd_dict
            if not config.gParamDict["force_run"]:
                print(f"File not valid :{rd_path}\n"
                      f"Please check again\n"
                      f"or delete {rd_path}\n "
                      f"or with option --force_run True")
                return None
            else:
                return {}

    except EOFError as e:
        print(f"EOFError:{e}\n"
              f"Please check again\n"
              f"or delete {rd_path}\n "
              f"or with option --force_run True")
        return None
    except pickle.UnpicklingError as e:
        print(f"pickle.UnpicklingError:{e}\n"
              f"Please check again\n"
              f"or delete {rd_path}\n "
              f"or with option --force_run True")
        return None


def log_to_file(cur_name="", new_name="", rd_path=""):

    id_nm, enc_str = get_id_enc(cur_name, new_name)
    if (not id_nm) or (not enc_str):
        print(f"Parameters not set,please check again.")
        sys.exit()
    if not rd_path:
        stamp = datetime.now().strftime("%Y%m%d%H%M%S%f")
        rd_path = os.path.join(config.gParamDict["record_path"],
                               config.rd_prefix_str + "_" + stamp + ".pkl")
    with open(rd_path, "wb") as fh:
        pickle.dump({id_nm: enc_str}, fh, protocol=pickle.HIGHEST_PROTOCOL)
