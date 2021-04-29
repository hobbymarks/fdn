import string
# From Third Party
from rich.console import Console
from rich.theme import Theme

console = Console(width=240, theme=Theme(inherit=False))
style = "black on white"

_gSC = "_"

gParamDict = {
    "console": (console, style),
    "sep_char": _gSC,
    "asc_len": 3,
    "head_chars": string.ascii_letters + string.digits + string.punctuation
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
