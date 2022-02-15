import importlib.resources
import json
import os
from pathlib import Path

# From Project
import fdn.data
# From Third party
import nltk
from nltk.corpus import words

gParamDict: dict = {}
with importlib.resources.path("fdn.data", "config.json") as cfg_path:
    with open(cfg_path, encoding="UTF-8") as fh:
        gParamDict.update(json.load(fh))

# TODO: can optimize to not require nltk package
nltk_path = os.path.dirname(fdn.data.__file__)
if os.path.isdir(nltk_path):
    nltk.data.path.append(nltk_path)
    if not os.path.isfile(os.path.join(nltk_path, "corpora", "words.zip")):
        nltk.download("words", download_dir=nltk_path)
else:
    try:
        from nltk.corpus import words

        gParamDict["LowerCaseWordSet"] = set(
            list(map(lambda x: x.lower(), words.words())))
    except LookupError:
        nltk.download("words")
gParamDict["LowerCaseWordSet"] = set(
    list(map(lambda x: x.lower(), words.words())))

gParamDict["record_path"] = os.path.join(Path.home(), ".fdn")
Path(gParamDict["record_path"]).mkdir(parents=True, exist_ok=True)
gParamDict["db_path"] = os.path.join(gParamDict["record_path"], "rdsa.db")
