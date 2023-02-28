package description

import (
	"fmt"
	"html"
)

type FormatItem struct {
	Text          string `json:"text"`
	Color         Color  `json:"color"`
	Obfuscated    bool   `json:"obfuscated"`
	Bold          bool   `json:"bold"`
	Strikethrough bool   `json:"strikethrough"`
	Underline     bool   `json:"underline"`
	Italic        bool   `json:"italic"`
}

func (i FormatItem) Raw() (result string) {
	if i.Color != White {
		result += i.Color.ToRaw()
	}

	if i.Obfuscated {
		result += "\u00A7k"
	}

	if i.Bold {
		result += "\u00A7l"
	}

	if i.Strikethrough {
		result += "\u00A7m"
	}

	if i.Underline {
		result += "\u00A7n"
	}

	if i.Italic {
		result += "\u00A7o"
	}

	result += i.Text

	return
}

func (i FormatItem) Clean() string {
	return i.Text
}

func (i FormatItem) HTML() (result string) {
	classes := make([]string, 0)

	styles := map[string]string{
		"color": i.Color.ToHex(),
	}

	if i.Obfuscated {
		classes = append(classes, "minecraft-format-obfuscated")
	}

	if i.Bold {
		styles["font-weight"] = "bold"
	}

	if i.Strikethrough {
		if _, ok := styles["text-decoration"]; ok {
			styles["text-decoration"] += " "
		}

		styles["text-decoration"] += "line-through"
	}

	if i.Underline {
		if _, ok := styles["text-decoration"]; ok {
			styles["text-decoration"] += " "
		}

		styles["text-decoration"] += "underline"
	}

	if i.Italic {
		styles["font-style"] = "italic"
	}

	result += "<span"

	if len(classes) > 0 {
		result += " class=\""

		for _, v := range classes {
			result += v
		}

		result += "\""
	}

	if len(styles) > 0 {
		result += " style=\""

		keys := make([]string, 0, len(styles))

		for k := range styles {
			keys = append(keys, k)
		}

		for i, l := 0, len(keys); i < l; i++ {
			key := keys[i]
			value := styles[key]

			result += fmt.Sprintf("%s: %s;", key, value)

			if i+1 != l {
				result += " "
			}
		}

		result += "\""
	}

	result += fmt.Sprintf(">%s</span>", html.EscapeString(i.Text))

	return
}
