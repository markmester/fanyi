/*
 * File: flag.go
 * Project: bot
 * File Created: Wednesday, 25th January 2023 2:47:57 pm
 * Author: Mark Mester (mmester6016@gmail.com)
 * -----
 * Last Modified: Wednesday, 25th January 2023 9:42:50 pm
 * Modified By: Mark Mester (mmester6016@gmail.com>)
 */
package slackbot

import "strings"

// flag emoji to language
var flagMap = map[string]string{
	"flag-ac": "English",
	"flag-ad": "Catalan",
	"flag-ae": "Arabic",
	"flag-af": "Pashto",
	"flag-ag": "English",
	"flag-ai": "English",
	"flag-al": "Albanian",
	"flag-am": "Armenian",
	"flag-ao": "Portuguese",
	"flag-ar": "Spanish",
	"flag-as": "English",
	"flag-at": "German",
	"flag-au": "English",
	"flag-aw": "Dutch",
	"flag-ax": "Swedish",
	"flag-az": "Spanish",
	"flag-ba": "Bosnian",
	"flag-bb": "English",
	"flag-bd": "Bengali",
	"flag-be": "Dutch",
	"flag-bf": "French",
	"flag-bg": "Bulgarian",
	"flag-bh": "Arabic",
	"flag-bi": "French",
	"flag-bj": "French",
	"flag-bl": "French",
	"flag-bn": "English",
	"flag-bm": "Malay",
	"flag-bo": "Spanish",
	"flag-bq": "Dutch",
	"flag-br": "Portuguese",
	"flag-bs": "English",
	"flag-bt": "Dzongkha",
	"flag-bv": "Norwegian",
	"flag-bw": "English",
	"flag-by": "Belarusian",
	"flag-bz": "English",
	"flag-ca": "English",
	"flag-cc": "Malay",
	"flag-cd": "French",
	"flag-cf": "French",
	"flag-cg": "French",
	"flag-ch": "German",
	"flag-ci": "French",
	"flag-ck": "English",
	"flag-cl": "Spanish",
	"flag-cm": "French",
	"flag-cn": "Chinese Simplified",
	"flag-co": "Spanish",
	"flag-cp": "French",
	"flag-cr": "Spanish",
	"flag-cu": "Spanish",
	"flag-cv": "Portuguese",
	"flag-cw": "Dutch",
	"flag-cx": "English",
	"flag-cy": "Greek",
	"flag-cz": "Czech",
	"flag-de": "German",
	"flag-dg": "English",
	"flag-dj": "French",
	"flag-dk": "Danish",
	"flag-dm": "English",
	"flag-do": "Spanish",
	"flag-dz": "Arabic",
	"flag-ea": "Spanish",
	"flag-ec": "Spanish",
	"flag-ee": "Estonian",
	"flag-eg": "Arabic",
	"flag-eh": "Arabic",
	"flag-er": "Arabic",
	"flag-es": "Spanish",
	"flag-et": "Oromo",
	"flag-fi": "Finnish",
	"flag-fj": "English",
	"flag-fk": "English",
	"flag-fm": "English",
	"flag-fr": "French",
	"flag-ga": "French",
	"flag-gb": "English",
	"flag-gd": "English",
	"flag-ge": "Georgian",
	"flag-gf": "French",
	"flag-gg": "English",
	"flag-gh": "English",
	"flag-gi": "English",
	"flag-gl": "Danish",
	"flag-gm": "English",
	"flag-gn": "French",
	"flag-gp": "French",
	"flag-gq": "Spanish",
	"flag-gr": "Greek",
	"flag-gs": "English",
	"flag-gt": "Spanish",
	"flag-gu": "English",
	"flag-gw": "Portuguese",
	"flag-gy": "English",
	"flag-hk": "Chinese Traditional",
	"flag-hn": "Spanish",
	"flag-hr": "Croatian",
	"flag-ht": "Haitian Creole",
	"flag-hu": "Hungarian",
	"flag-ic": "Spanish",
	"flag-id": "Indonesian",
	"flag-ie": "Irish",
	"flag-il": "Hebrew",
	"flag-im": "English",
	"flag-in": "Hindi",
	"flag-io": "English",
	"flag-iq": "Arabic",
	"flag-ir": "Persian",
	"flag-is": "Icelandic",
	"flag-it": "Italian",
	"flag-je": "English",
	"flag-jm": "English",
	"flag-jo": "Arabic",
	"flag-jp": "Japanese",
	"flag-ke": "English",
	"flag-kg": "Kyrgyz",
	"flag-kh": "Khmer",
	"flag-ki": "English",
	"flag-kn": "English",
	"flag-kp": "Korean",
	"flag-kr": "Korean",
	"flag-kw": "Arabic",
	"flag-ky": "English",
	"flag-kz": "Kazakh",
	"flag-la": "Lao",
	"flag-lb": "Arabic",
	"flag-lc": "English",
	"flag-li": "German",
	"flag-lk": "Sinhala",
	"flag-lr": "English",
	"flag-ls": "Sesotho",
	"flag-lt": "Lithuanian",
	"flag-lu": "Luxembourgish",
	"flag-lv": "Latvian",
	"flag-ly": "Arabic",
	"flag-ma": "Arabic",
	"flag-mc": "French",
	"flag-md": "Romanian",
	"flag-mg": "Malagasy",
	"flag-mh": "Marshallese",
	"flag-mk": "Macedonian",
	"flag-ml": "French",
	"flag-mm": "Burmese",
	"flag-mn": "Mongolian",
	"flag-mo": "Chinese Traditional",
	"flag-mp": "English",
	"flag-mq": "French",
	"flag-mr": "Arabic",
	"flag-ms": "English",
	"flag-mt": "Maltese",
	"flag-mu": "English",
	"flag-mv": "Dhivehi",
	"flag-mw": "English",
	"flag-mx": "Spanish",
	"flag-my": "Malay",
	"flag-mz": "Portuguese",
	"flag-na": "English",
	"flag-nc": "French",
	"flag-ne": "French",
	"flag-nf": "English",
	"flag-ng": "English",
	"flag-ni": "Spanish",
	"flag-nl": "Dutch",
	"flag-no": "Norwegian",
	"flag-np": "Nepali",
	"flag-nr": "Nauru",
	"flag-nu": "Niuean",
	"flag-nz": "English",
	"flag-om": "Arabic",
	"flag-pa": "Spanish",
	"flag-pe": "Spanish",
	"flag-pf": "French",
	"flag-pg": "English",
	"flag-ph": "Tagalog",
	"flag-pk": "Urdu",
	"flag-pl": "Polish",
	"flag-pm": "French",
	"flag-pn": "English",
	"flag-pr": "Spanish",
	"flag-ps": "Arabic",
	"flag-pt": "Portuguese",
	"flag-pw": "English",
	"flag-py": "Spanish",
	"flag-qa": "Arabic",
	"flag-re": "French",
	"flag-ro": "Romanian",
	"flag-rs": "Serbian",
	"flag-ru": "Russian",
	"flag-rw": "Kinyarwanda",
	"flag-sa": "Arabic",
	"flag-sb": "English",
	"flag-sc": "English",
	"flag-sd": "Arabic",
	"flag-se": "Swedish",
	"flag-sg": "English",
	"flag-sh": "English",
	"flag-si": "Slovenian",
	"flag-sj": "Norwegian",
	"flag-sk": "Slovak",
	"flag-sl": "English",
	"flag-sm": "Italian",
	"flag-sn": "French",
	"flag-so": "Somali",
	"flag-sr": "Dutch",
	"flag-ss": "English",
	"flag-st": "Portuguese",
	"flag-sv": "Spanish",
	"flag-sx": "Dutch",
	"flag-sw": "Arabic",
	"flag-sz": "Swati",
	"flag-ta": "English",
	"flag-tc": "English",
	"flag-td": "French",
	"flag-tf": "French",
	"flag-tg": "French",
	"flag-th": "Thai",
	"flag-tj": "Tajik",
	"flag-tk": "Tokelau",
	"flag-tl": "Tetum",
	"flag-tm": "Turkmen",
	"flag-tn": "Arabic",
	"flag-tr": "Turkish",
	"flag-tt": "English",
	"flag-tv": "Tuvalua",
	"flag-tw": "Chinese Traditional",
	"flag-tz": "Swahili",
	"flag-ua": "Ukrainian",
	"flag-ug": "English",
	"flag-um": "English",
	"flag-us": "English",
	"flag-uy": "Spanish",
	"flag-uz": "Uzbek",
	"flag-va": "Italian",
	"flag-vc": "English",
	"flag-ve": "Spanish",
	"flag-vg": "English",
	"flag-vi": "English",
	"flag-vn": "Vietnamese",
	"flag-vu": "English",
	"flag-wf": "French",
	"flag-ws": "Samoan",
	"flag-xk": "Albanian",
	"flag-ye": "Arabic",
	"flag-yt": "French",
	"flag-za": "Afrikaans",
	"flag-zm": "English",
	"flag-zw": "English",
	"flag-to": "",
	"flag-me": "",
	"flag-km": "",
	"flag-hm": "",
	"flag-mf": "Saint Martin",
	"flag-fo": "Faroe Islands",
	"flag-eu": "EU",
	"flag-aq": "Antarctica",
}
var alternateFlagMap = func() map[string]string {
	alternateFlagMap := map[string]string{}
	for k, v := range flagMap {
		alternateFlagMap[strings.TrimPrefix(k, "flag-")] = v
	}
	return alternateFlagMap
}()

// GetLanguageCode is
func GetLanguageCode(flag string) (code string, ok bool) {
	code, ok = flagMap[flag]
	if !ok {
		code, ok = alternateFlagMap[flag]
	}
	return
}
