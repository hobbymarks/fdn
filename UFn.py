#!/usr/bin/env python
import difflib
import hashlib
import collections
import os
import pickle
from datetime import datetime
import string
import sys
import click
from rich.console import Console
from rich.theme import Theme
import re


class FileNameLog:

    def __init__(self, filePath="", md5Value=""):
        if not md5Value:
            assert os.path.exists(filePath)
            with open(filePath, "rb") as fhand:
                data = fhand.read()
                md5Value = hashlib.md5(data).hexdigest()
        self.md5Value = str(md5Value)
        self._currentName = os.path.basename(filePath)
        self._nameRecord = collections.OrderedDict()

    def changeFileName(self, newFileName="", stamp=""):
        assert newFileName
        if not stamp:
            stamp = datetime.now().strftime("%Y%m%d%H%M%S%f")
        self._nameRecord[stamp] = self._currentName
        self._currentName = newFileName

    def getFileNameHistory(self):
        return {
            self.md5Value: {
                "currentName": self._currentName,
                "nameRecord": self._nameRecord
            }
        }


def createFileNameLog(filePath="", md5Value=""):
    global gParamDict
    rdPath = gParamDict["recordPath"]
    if not md5Value:
        assert os.path.isfile(filePath)
        with open(filePath, "rb") as fhand:
            data = fhand.read()
            md5Value = hashlib.md5(data).hexdigest()
    fRecordPath = os.path.join(rdPath, str(md5Value) + "_HRd.pkl")
    if os.path.isfile(fRecordPath):
        with open(fRecordPath, "rb") as fhand:
            rd = pickle.load(fhand)
        return rd
    else:
        assert os.path.isfile(filePath)
        return FileNameLog(filePath)


def isHiddenFile(path):
    if os.name == "nt":
        import win32api, win32con
    if os.name == "nt":
        attribute = win32api.GetFileAttributes(path)
        return attribute & (win32con.FILE_ATTRIBUTE_HIDDEN |
                            win32con.FILE_ATTRIBUTE_SYSTEM)
    else:
        return os.path.basename(path).startswith('.')  #linux-osx


def maskOriginal(s=""):
    global gParamDict
    kpOrgList = gParamDict["KeepOriginalList"]
    sepChar = gParamDict["sepChar"]
    reStr = "|".join([re.escape(sepStr) for sepStr in kpOrgList])
    wordList = re.split(f'({reStr})', s)
    maskList = []
    for elm in wordList:
        if elm in kpOrgList:
            maskList.extend([True])
        else:
            maskList.extend([False])

    return (wordList, maskList)


def replaceChar(fNameStr=""):
    global gParamDict
    charDict = gParamDict["CharDictionary"]
    root, ext = os.path.splitext(fNameStr)
    wordList = []
    maskList = []
    wordList, maskList = maskOriginal(root)
    for key, value in charDict.items():
        newWordList = []
        for word, mask in zip(wordList, maskList):
            if not mask:
                newWordList.append(word)
        if key in "".join(newWordList):  #If need repalce ,then go ...
            newWordList = []
            for word, mask in zip(wordList, maskList):
                if (key in word) and (not mask):
                    newWordList.append(word.replace(key, value))
                else:
                    newWordList.append(word)
            gParamDict["nameRdList"].append("".join(newWordList) + ext)
            wordList = newWordList
    return "".join(wordList) + ext


def processHeadTailSepChar(fNameStr=""):
    """
    Process When globalSepChar at Head or Tail of the file name exclude extension.
    """
    global gParamDict
    sepChar = gParamDict["sepChar"]
    root, ext = os.path.splitext(fNameStr)
    if root.startswith(sepChar):
        root = sepChar.join(root.split(sepChar)[1:])
        gParamDict["nameRdList"].append(root + ext)
    if root.endswith(sepChar):
        root = sepChar.join(root.split(sepChar)[0:-1])
        gParamDict["nameRdList"].append(root + ext)
    return root + ext


def processHeadTail(fNameStr=""):
    """
    Process Head and Tail of the file name.
    """
    global gParamDict
    assert fNameStr
    newName = processHeadTailSepChar(fNameStr=fNameStr)
    # Capitalize The First Letter
    if newName[0].islower():
        newName = newName[0].upper() + newName[1:]
        gParamDict["nameRdList"].append(newName)

    return newName


def processWhiteSpace(fNameStr=""):
    """
    Process all whitespace char ...
    """
    global gParamDict
    sepChar = gParamDict["sepChar"]
    root, ext = os.path.splitext(fNameStr)
    newName = sepChar.join(root.split()) + ext
    if newName != fNameStr:
        gParamDict["nameRdList"].append(newName)
    return newName


def checkStartsWithTerminology(word=""):
    global gParamDict
    termDict = gParamDict["TerminologyDictionary"]
    for key in termDict.keys():
        if word.lower().startswith(key):
            return key
    return None


def processTerminology(fNameStr=""):
    global gParamDict

    root, ext = os.path.splitext(fNameStr)
    wordList = root.split(gParamDict["sepChar"])
    newWordList = []
    termDict = gParamDict["TerminologyDictionary"]
    sepChar = gParamDict["sepChar"]
    for word in wordList:
        if word.lower() in termDict.keys():
            newWordList.append(termDict[word.lower()])
        elif key := checkStartsWithTerminology(word):
            newWord = termDict[key] + word[len(key):]
            newWordList.append(newWord)
        else:
            newWordList.append(word)
    if newWordList != wordList:
        gParamDict["nameRdList"].append(sepChar.join(newWordList) + ext)
    return sepChar.join(newWordList) + ext


def processWord(fNameStr=""):
    global gParamDict
    sepChar = gParamDict["sepChar"]
    root, ext = os.path.splitext(fNameStr)
    wordList = root.split(sepChar)
    newWordList = []
    wordSet = gParamDict["LowerCaseWordSet"]
    for word in wordList:
        if word.lower() in wordSet:
            newWordList.append(string.capwords(word))
        else:
            newWordList.append(word)

    return sepChar.join(newWordList) + ext


def depthWalk(topPath, topDown=True, followLinks=False, maxDepth=1):
    if str(maxDepth).isnumeric():
        maxDepth = int(maxDepth)
    else:
        maxDepth = 1
    names = os.listdir(topPath)
    dirs, nondirs = [], []
    for name in names:
        if os.path.isdir(os.path.join(topPath, name)):
            dirs.append(name)
        else:
            nondirs.append(name)
    if topDown:
        yield topPath, dirs, nondirs
    if maxDepth is None or maxDepth > 1:
        for name in dirs:
            newPath = os.path.join(topPath, name)
            if followLinks or not os.path.islink(newPath):
                for x in depthWalk(newPath, topDown, followLinks,
                                   None if maxDepth is None else maxDepth - 1):
                    yield x
    if not topDown:
        yield topPath, dirs, nondirs


def richStyle(originString="", processedString=""):
    richS1 = richS2 = ""
    richS1DifPos = richS2DifPos = 0
    for match in difflib.SequenceMatcher(0, originString,
                                         processedString).get_matching_blocks():
        if richS1DifPos < match.a:
            richS1 += "[bold red]" + originString[richS1DifPos:match.a].replace(
                " ", "▯") + "[/bold red]" + originString[match.a:match.a +
                                                         match.size]
            richS1DifPos = match.a + match.size
        else:
            richS1 += originString[match.a:match.a + match.size]
            richS1DifPos = match.a + match.size

        if richS2DifPos < match.b:
            richS2 += "[bold green]" + processedString[
                richS2DifPos:match.b].replace(
                    " ",
                    "▯") + "[/bold green]" + processedString[match.b:match.b +
                                                             match.size]
            richS2DifPos = match.b + match.size
        else:
            richS2 += processedString[match.b:match.b + match.size]
            richS2DifPos = match.b + match.size

    return richS1, richS2


def oneFileNameUFn(fNameStr=""):
    """
    Process Only one file name
    """
    global gParamDict
    newName = fNameStr
    # all whitespace replace by sepChar
    newName = processWhiteSpace(fNameStr=newName)
    # replace characters by defined Dictionary
    newName = replaceChar(fNameStr=newName)
    # Capwords only when word in wordsSet
    newName = processWord(fNameStr=newName)
    # Pretty Terminology
    newName = processTerminology(fNameStr=newName)
    # process Head and Tail
    newName = processHeadTail(fNameStr=newName)

    return newName


def oneDirPathUFn(targetPath, console, style):
    """
    Process Only one Directory Path
    """
    global gParamDict
    for subdir, dirs, files in depthWalk(topPath=targetPath,
                                         maxDepth=gParamDict["maxdepth"]):
        for file in files:
            #             if not os.path.isfile(file):
            #                 continue
            gParamDict["nameRdList"] = []
            oldNamePath = os.path.join(subdir, file)
            if isHiddenFile(oldNamePath):
                continue
            newName = oneFileNameUFn(fNameStr=file)
            # Create full path
            newNamePath = os.path.join(subdir, newName)
            if not gParamDict["dry"]:
                # Create Or Update File Name Change Record and Save to File
                # then rename file name
                if newName != file:
                    fileHistoryRecord = createFileNameLog(filePath=oldNamePath)
                    with open(
                            os.path.join(
                                gParamDict["recordPath"],
                                str(fileHistoryRecord.md5Value) + "_HRd.pkl"),
                            "wb") as fhand:
                        fileHistoryRecord.changeFileName(newFileName=newName)
                        pickle.dump(fileHistoryRecord, fhand)
                    os.rename(oldNamePath, newNamePath)

            if newName != file:
                richFile, richNewName = richStyle(file, newName)
                console.print(" " * 3 + richFile, style=style)
                if not gParamDict["simple"]:
                    for fName in gParamDict["nameRdList"]:
                        console.print("---" + fName, style=style)
                if gParamDict["dry"]:
                    console.print("-->" + richNewName, style=style)
                else:
                    console.print("==>" + richNewName, style=style)


@click.command(context_settings={"ignore_unknown_options": True})
# @click.command()
@click.argument("argpath",
                required=False,
                type=click.Path(exists=True),
                nargs=-1)
@click.option(
    "--path",
    #     prompt="target path",
    default=".",
    help="Recursively traverse the path,All file's name will be changed.",
    show_default=True)
@click.option("--maxdepth",
              default=1,
              type=str,
              help="Set travel directory tree with max depth.",
              show_default=True)
@click.option(
    "--exclude",
    default="",
    help="Exclude all files in exclude path.Not valid in current version.",
    show_default=True)
@click.option("--dry",
              default=True,
              type=bool,
              help="If dry is True will not change file name.",
              show_default=True)
@click.option("--simple",
              default=True,
              type=bool,
              help="If simple is True Only print changed file name.",
              show_default=True)
def ufn(argpath, path, maxdepth, exclude, dry, simple):
    """Files in PATH will be changed file names unified.
    
    You can direct set path such as UFn.py path ...
    """
    global gParamDict
    gParamDict["argpath"] = argpath
    gParamDict["path"] = path
    gParamDict["maxdepth"] = maxdepth
    gParamDict["exclude"] = exclude
    gParamDict["dry"] = dry
    gParamDict["simple"] = simple

    console = Console(width=240, theme=Theme(inherit=False))
    style = "black on white"

    targetPath = ""

    if gParamDict["argpath"]:
        targetPath = gParamDict["argpath"]
    else:
        if not os.path.isdir(gParamDict["path"]):
            click.echo("%s is not valid path.")
            return -1
        targetPath = gParamDict["path"]

    with open(os.path.join(gParamDict["dataPath"], "CharDictionary.pkl"),
              "rb") as fhand:
        gParamDict["CharDictionary"] = pickle.load(fhand)
    with open(os.path.join(gParamDict["dataPath"], "TerminologyDictionary.pkl"),
              "rb") as fhand:
        gParamDict["TerminologyDictionary"] = pickle.load(fhand)
    with open(os.path.join(gParamDict["dataPath"], "LowerCaseWordSet.pkl"),
              "rb") as fhand:
        gParamDict["LowerCaseWordSet"] = pickle.load(fhand)
    with open(os.path.join(gParamDict["dataPath"], "KeepOriginalList.pkl"),
              "rb") as fhand:
        gParamDict["KeepOriginalList"] = pickle.load(fhand)
    for path in targetPath:
        oneDirPathUFn(path, console, style)

    if gParamDict["dry"]:
        console.print("*" * 80)
        console.print(
            "In order to take effect,run the CLI add option '--dry False'")

    return 0


if __name__ == "__main__":
    if (sys.version_info.major, sys.version_info.minor) < (3, 8):
        console.print(
            f"current Version is {sys.version},\n Please upgrade to at least 3.8."
        )
        sys.exit()
    scriptDirPath = os.path.dirname(os.path.realpath(__file__))

    gParamDict = {}
    gParamDict["sepChar"] = "_"
    gParamDict["dataPath"] = os.path.join(scriptDirPath, "data")
    gParamDict["recordPath"] = os.path.join(gParamDict["dataPath"], "hRdDir")
    gParamDict["nameRdList"] = []

    ufn()
