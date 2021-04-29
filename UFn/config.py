import string

_gSC = "_"

gParamDict = {
    "SeparatorChar": _gSC,
    "ASCLen": 3,
    "AlternateFlag":True,
    "HeadChars": string.ascii_letters + string.digits + string.punctuation
}

rd_prefix_str = "CNSH_ONCrypt_dict"

gParamDict["CharDictionary"] = {
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
    "…": _gSC,
    "“": _gSC,
    "”": _gSC,
    #     ".": _gSC,
    "•": _gSC,
    "，": _gSC,
    "–": _gSC,
    "—": _gSC,
    #     "一": _gSC,#It is a chinese number character, means one
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
    "____": _gSC,
    "___": _gSC,
    "__": _gSC,
    "._": _gSC,
    "What’s": "What_is",
    "what’s": "what_is"
}
gParamDict["TerminologyDictionary"] = {
    #     "apple": "Apple",
    "asciinema": "asciinema",
    "api": "API",
    "atm": "ATM",
    "cmsis": "CMSIS",
    "cypress": "CYPRESS",
    "dji": "DJI",
    #     "google": "Google",
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
}

gParamDict["KeepOriginalList"] = [
    "$*", "$@", "$#", "$?", "$-", "$$", "$!", "$0"
]

gParamDict["IgnoredDirectoryKeyList"] = [
    ".git", ".xcodeproj", ".cydsn", ".cywrk"
]
