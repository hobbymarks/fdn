import string

_gSC = "_"

gParamDict = {
    "SeparatorChar": _gSC,
    "ASCLen": 3,
    "AlternateFlag": True,
    "HeadChars": string.ascii_letters + string.digits + string.punctuation
}

rd_prefix_str = "CNSH_ONCrypt_dict"

gParamDict["BeReplacedCharDictionary"] = {
    "\r": _gSC,
    "\n": _gSC,
    "?": _gSC,
    ",": _gSC,
    "!": _gSC,
    ":": _gSC,
    "&": _gSC,
    "@": _gSC,
    "·": _gSC,  # at the middle height of row
    "\`": _gSC,
    "`": _gSC,
    "\\": _gSC,
    " ": _gSC,  # space
    "(": _gSC,
    ")": _gSC,
    "'": _gSC,
    "+": _gSC,
    "-": _gSC,
    "=": _gSC,
    "|": _gSC,
    "[": _gSC,
    "]": _gSC,
    "{": _gSC,
    "}": _gSC,
    "»": _gSC,
    "«": _gSC,
    "\"": _gSC,
    "*": _gSC,
    "#": _gSC,
    "®": _gSC,
    "™": _gSC,
    "…": _gSC,
    "“": _gSC,
    "”": _gSC,
    #     ".": _gSC,
    "•": _gSC,
    "，": _gSC,
    "。": _gSC,
    "–": _gSC,
    "—": _gSC,
    #     "一": _gSC,# It is a chinese number character, means one
    "、": _gSC,
    "（": _gSC,
    "）": _gSC,
    "《": _gSC,
    "》": _gSC,
    ">": _gSC,
    "【": _gSC,
    "】": _gSC,
    "「": _gSC,
    "」": _gSC,
    "｜": _gSC,
    "：": _gSC,
    "；": _gSC,
    "？": _gSC,
    "！": _gSC,
    "🚀": _gSC,
    "🚴": _gSC,
    "🌏": _gSC,
    "🐾": _gSC,
    "❤️": _gSC,
    "%2F": _gSC,
    # "____": _gSC,# This is a rougher way to deal with continuous separator
    # "___": _gSC,# character.
    # "__": _gSC,# Now replaced by a regex,so delete them
    "._": _gSC,
    "What’s": "What_is",
    "what’s": "what_is"
}
gParamDict["TerminologyDictionary"] = {
    # "apple": "Apple",# Already in nltk words and also match capwords so
    # delete
    "asciinema": "asciinema",
    "api": "API",
    "atm": "ATM",
    "cmsis": "CMSIS",
    "cypress": "CYPRESS",
    "dji": "DJI",
    # "google": "Google",# Already in nltk words and also match capwords so
    # delete
    "i2c": "I2C",
    "kicad": "KiCAD",
    "mbed": "Mbed",
    "mosfet": "MOSFET",
    "mux": "MUX",
    "nltk": "NLTK",
    "nucleo": "NUCLEO",
    "pcb": "PCB",
    "pcie": "PCIe",
    "psoc": "PSoC",
    "rohs": "ROHS",
    "spi": "SPI",
    "stm32": "STM32",
    "stmicroelectronics": "STMicroelectronics",
    "ti": "TI",
    "usb": "USB",
    "vishay": "VISHAY",
    "vim": "Vim",
}

gParamDict["RemainUnchangedWordList"] = [
    "$*", "$@", "$#", "$?", "$-", "$$", "$!", "$0"
]

gParamDict["IgnoredDirectoryKeyList"] = [
    ".git", ".xcodeproj", ".cydsn", ".cywrk"
]
