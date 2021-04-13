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
    global globalHistoryRecordPath
    if not md5Value:
        assert os.path.isfile(filePath)
        with open(filePath, "rb") as fhand:
            data = fhand.read()
            md5Value = hashlib.md5(data).hexdigest()
    fRecordPath = os.path.join(globalHistoryRecordPath,
                               str(md5Value) + "_HRd.pkl")
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


def replaceChar(charDict={}, tgtString=""):
    global globalFileNameHistoryRecordList
    assert charDict
    assert tgtString
    root, ext = os.path.splitext(tgtString)
    for key, value in charDict.items():
        if key in root:
            root = root.replace(key, value)
            globalFileNameHistoryRecordList.append(root + ext)
    return root + ext


def processHeadTailChar(tgtString=""):
    global globalFileNameHistoryRecordList
    assert tgtString
    specChar = "_"
    root, ext = os.path.splitext(tgtString)
    if root.startswith(specChar):
        root = specChar.join(root.split(specChar)[1:])
        globalFileNameHistoryRecordList.append(root + ext)
    if root.endswith(specChar):
        root = specChar.join(root.split(specChar)[0:-1])
        globalFileNameHistoryRecordList.append(root + ext)
    return root + ext


def checkStartsWithTerminology(termDictionary={}, word=""):
    for key in termDictionary.keys():
        if word.lower().startswith(key):
            return key
    return None


def processTerminology(termDictionary={}, tgtString=""):
    global globalFileNameHistoryRecordList
    assert termDictionary
    assert tgtString
    specChar = "_"
    root, ext = os.path.splitext(tgtString)
    wordList = root.split(specChar)
    newWordList = []
    for word in wordList:
        if word.lower() in termDictionary.keys():
            newWordList.append(termDictionary[word.lower()])
        elif key := checkStartsWithTerminology(termDictionary, word):
            newWord = termDictionary[key] + word[len(key):]
            newWordList.append(newWord)
        else:
            newWordList.append(word)
    if newWordList != wordList:
        globalFileNameHistoryRecordList.append(specChar.join(newWordList) + ext)
    return specChar.join(newWordList) + ext


def processWord(wordSet=set(), tgtString=""):
    assert wordSet
    assert tgtString
    specChar = "_"
    root, ext = os.path.splitext(tgtString)
    wordList = root.split(specChar)
    newWordList = []
    for word in wordList:
        if word.lower() in wordSet:
            newWordList.append(string.capwords(word))
        else:
            newWordList.append(word)

    return specChar.join(newWordList) + ext


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


def onlyOnePathUFn(globalParameterDictionary, targetPath, CharDictionary,
                   TerminologyDictionary, LowerCaseWordSet, console, style):
    global globalFileNameHistoryRecordList
    for subdir, dirs, files in depthWalk(
            topPath=targetPath, maxDepth=globalParameterDictionary["maxdepth"]):
        for file in files:
            #             if not os.path.isfile(file):
            #                 continue
            globalFileNameHistoryRecordList = []
            oldNamePath = os.path.join(subdir, file)
            if isHiddenFile(oldNamePath):
                continue
            newName = file
            # replace characters by defined Dictionary
            newName = replaceChar(charDict=CharDictionary, tgtString=newName)
            # process Head and Tail character exclude file name extension
            newName = processHeadTailChar(tgtString=newName)
            # Capwords only when word in wordsSet
            newName = processWord(wordSet=LowerCaseWordSet, tgtString=newName)
            # Pretty Terminology
            newName = processTerminology(termDictionary=TerminologyDictionary,
                                         tgtString=newName)
            # Capitalize The First Letter
            if newName[0].islower():
                newName = newName[0].upper() + newName[1:]
                globalFileNameHistoryRecordList.append(newName)
            # Create full path
            newNamePath = os.path.join(subdir, newName)
            if not globalParameterDictionary["dry"]:
                # Create Or Update File Name Change Record and Save to File
                # then rename file name
                if newName != file:
                    fileHistoryRecord = createFileNameLog(filePath=oldNamePath)
                    with open(
                            os.path.join(
                                globalHistoryRecordPath,
                                str(fileHistoryRecord.md5Value) + "_HRd.pkl"),
                            "wb") as fhand:
                        fileHistoryRecord.changeFileName(newFileName=newName)
                        pickle.dump(fileHistoryRecord, fhand)
                    os.rename(oldNamePath, newNamePath)

            if newName != file:
                richFile, richNewName = richStyle(file, newName)
                console.print(" " * 3 + richFile, style=style)
                if not globalParameterDictionary["simple"]:
                    for fName in globalFileNameHistoryRecordList:
                        console.print("---" + fName, style=style)
                if globalParameterDictionary["dry"]:
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
    global globalDataPath
    global globalParameterDictionary
    global globalFileNameHistoryRecordList
    """Files in PATH will be changed file names unified.
    
    You can direct set path such as UFn.py path ...
    """
    globalParameterDictionary["argpath"] = argpath
    globalParameterDictionary["path"] = path
    globalParameterDictionary["maxdepth"] = maxdepth
    globalParameterDictionary["exclude"] = exclude
    globalParameterDictionary["dry"] = dry
    globalParameterDictionary["simple"] = simple

    console = Console(width=240, theme=Theme(inherit=False))
    style = "black on white"

    targetPath = ""

    if globalParameterDictionary["argpath"]:
        targetPath = globalParameterDictionary["argpath"]
    else:
        if not os.path.isdir(globalParameterDictionary["path"]):
            click.echo("%s is not valid path.")
            return -1
        targetPath = globalParameterDictionary["path"]

    with open(os.path.join(globalDataPath, "CharDictionary.pkl"),
              "rb") as fhand:
        CharDictionary = pickle.load(fhand)
    with open(os.path.join(globalDataPath, "TerminologyDictionary.pkl"),
              "rb") as fhand:
        TerminologyDictionary = pickle.load(fhand)
    with open(os.path.join(globalDataPath, "LowerCaseWordSet.pkl"),
              "rb") as fhand:
        LowerCaseWordSet = pickle.load(fhand)
    for path in targetPath:
        onlyOnePathUFn(globalParameterDictionary, path, CharDictionary,
                       TerminologyDictionary, LowerCaseWordSet, console, style)

    if globalParameterDictionary["dry"]:
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

    globalDataPath = os.path.join(scriptDirPath, "data")
    globalHistoryRecordPath = os.path.join(globalDataPath, "hRdDir")

    globalParameterDictionary = {}
    globalFileNameHistoryRecordList = []

    ufn()
