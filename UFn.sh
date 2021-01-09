#!/bin/bash

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

rename -vS "…" _ *
rename -vS "“" _ *
rename -vS "”" _ *
rename -X -vS "." _ *
rename -vS "•" _ *
rename -vS "，" _ *
rename -vS "–" _ *
rename -vS "—" _ *
rename -vS "一" _ *
rename -vS "、" _ *
rename -vS "（" _ *
rename -vS "）" _ *
rename -vS "《" _ *
rename -vS "》" _ *
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

rename -vS "____" _ *
rename -vS "___" _ *
rename -vS "__" _ *
rename -vS "._" _ *

rename -v "s/^_//g"  *

rename -vS "What’s" "What_is" *
rename -vS "what’s" "what_is" *

rename -v 's/^([a-z])/_\U$1/' *

rename -v -X --trim *
