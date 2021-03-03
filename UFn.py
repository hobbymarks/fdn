
import os

data_path = "./data/"


def isHiddenFile(path):
    import os
    if os.name == "nt":
        import win32api, win32con
    if os.name == "nt":
        attribute = win32api.GetFileAttributes(path)
        return attribute & (win32con.FILE_ATTRIBUTE_HIDDEN |
                            win32con.FILE_ATTRIBUTE_SYSTEM)
    else:
        return os.path.basename(path).startswith('.')  #linux-osx


def replaceChar(charDict={}, tgtString=""):
    assert charDict
    assert tgtString
    import os
    import click
    root, ext = os.path.splitext(tgtString)
    for key, value in charDict.items():
        if key in root:
            root = root.replace(key, value)


#             click.echo("---%s" % root + ext)
    return root + ext


def processHeadTailChar(tgtString=""):
    assert tgtString
    specChar = "_"
    import os
    root, ext = os.path.splitext(tgtString)
    if root.startswith(specChar):
        root = specChar.join(root.split(specChar)[1:])
    if root.endswith(specChar):
        root = specChar.join(root.split(specChar)[0:-1])
    return root + ext


def processTerminology(termDictionary={}, tgtString=""):
    assert termDictionary
    assert tgtString
    specChar = "_"
    import os
    root, ext = os.path.splitext(tgtString)
    wordList = root.split(specChar)
    newWordList = []
    for word in wordList:
        if word.lower() in termDictionary.keys():
            newWordList.append(termDictionary[word.lower()])
        else:
            newWordList.append(word)
    return specChar.join(newWordList) + ext


def processWord(wordSet=set(), tgtString=""):
    assert wordSet
    assert tgtString
    specChar = "_"
    import os
    import string
    root, ext = os.path.splitext(tgtString)
    wordList = root.split(specChar)
    newWordList = []
    for word in wordList:
        if word.lower() in wordSet:
            newWordList.append(string.capwords(word))
        else:
            newWordList.append(word)

    return specChar.join(newWordList) + ext


import click


@click.command()
@click.option("-p",
              "--path",
              prompt="target path",
              help="files in path will be changed name.")
@click.option("-E",
              "--exclude",
              default="",
              help="exclude all files in exclude path")
@click.option("-d",
              "--dry",
              default=True,
              type=bool,
              help="if dry is True will not change file name.Default is True.")
@click.option(
    "-s",
    "--simple",
    default=True,
    type=bool,
    help="if simple is True Only print changed file name.Default is True.")
def ufn(path, exclude, dry, simple):
    """Files in PATH will be changed file names unified."""
    import os
    if not os.path.isdir(path):
        click.echo("%s is not valid path.")
        return -1

    import pickle
    with open(os.path.join(data_path, "CharDictionary.pkl"), "rb") as fhand:
        CharDictionary = pickle.load(fhand)
    with open(os.path.join(data_path, "TerminologyDictionary.pkl"),
              "rb") as fhand:
        TerminologyDictionary = pickle.load(fhand)
    with open(os.path.join(data_path, "LowerCaseWordSet.pkl"), "rb") as fhand:
        LowerCaseWordSet = pickle.load(fhand)
    for subdir, dirs, files in os.walk(path):
        for file in files:
            #             if not os.path.isfile(file):
            #                 continue
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
            newName = newName[0].upper() + newName[1:]
            # Create full path
            newNamePath = os.path.join(subdir, newName)
            if not dry:
                # rename file name
                os.rename(oldNamePath, newNamePath)
            if (not simple) or (newName != file):
                click.echo("   %s" % file)
                click.echo("==>%s" % newName)

    return 0


if __name__ == "__main__":
    ufn()
