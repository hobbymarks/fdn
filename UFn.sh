#!/bin/bash

RED='\033[0;31m' # Red
GREEN='\033[0;32m' #Green
NC='\033[0m' # No Color

if ! command -v rename $> /dev/null ;then
    echo -e "${RED}rename could not be found.${NC}"
    echo "(http://plasmasturm.org/code/rename/)"
    exit
fi

rename -v "s/\r//g"  *
rename -v "s/\n//g"  *

rename -vS "?" _ *
rename -vS "," _ *
rename -vS "!" _ *
rename -vS ":" _ *
rename -vS "&" _ *
rename -vS "@" _ *
rename -vS "·" _ *
rename -vS "\`" _ *
rename -vS " " _ *
rename -vS "(" _ *
rename -vS ")" _ *
rename -vS "'" _ *
rename -vS "+" _ *
rename -vS "-" _ *
rename -vS "=" _ *
rename -vS "|" _ *
rename -vS "[" _ *
rename -vS "]" _ *
rename -vS "{" _ *
rename -vS "}" _ *
rename -vS "»" _ *
rename -vS "«" _ *
rename -vS "\"" _ *
rename -vS "*" _ *
rename -vS "#" _ *
rename -vS "®" _ *

rename -vS "…" _ *
rename -vS "“" _ *
rename -vS "”" _ *
rename -X -vS "." _ *
rename -vS "•" _ *
rename -vS "，" _ *
rename -vS "–" _ *
rename -vS "—" _ *
rename -vS "、" _ *
rename -vS "（" _ *
rename -vS "）" _ *
rename -vS "《" _ *
rename -vS "》" _ *
rename -vS ">" _ *
rename -vS "【" _ *
rename -vS "】" _ *
rename -vS "「" _ *
rename -vS "」" _ *
rename -vS "｜" _ *
rename -vS "：" _ *
rename -vS "？" _ *
rename -vS "！" _ *

rename -vS "🚀" _ *
rename -vS "🚴" _ *
rename -vS "🌏" _ *
rename -vS "🐾" _ *

rename -vS "%2F" _ *

rename -vS "____" _ *
rename -vS "___" _ *
rename -vS "__" _ *
rename -vS "._" _ *

rename -v "s/^_//g"  *

rename -vS "What’s" "What_is" *
rename -vS "what’s" "what_is" *

rename -v 's/^([a-z])/_\U$1/' *

rename -v -X --trim *
