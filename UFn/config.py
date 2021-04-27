from datetime import datetime
import os
import pickle
import string
from rich.console import Console
from rich.theme import Theme

console = Console(width=240, theme=Theme(inherit=False))
style = "black on white"

gParamDict = {
    "console": (console, style),
    "sep_char": "_",
    "record_list": [],
    "asc_len": 3,
    "head_chars": string.ascii_letters + string.digits + string.punctuation,
    "force_run": False
}

# scriptDirPath = os.path.dirname(os.path.realpath(__file__))
# gParamDict["data_path"] = os.path.join(scriptDirPath, "data")
rd_prefix_str = "CNSH_ONCrypt_dict"
# stamp = datetime.now().strife("%Y%m%d%H%M%S%f")
# gParamDict["record_path"] = os.path.join(scriptDirPath, "data",
#                                          rd_prefix_str + "_" + stamp + ".pkl")
# if not os.path.isfile(gParamDict["record_path"]):
#     with open(gParamDict["record_path"], "wb") as fh:
#         pickle.dump({}, fh)
