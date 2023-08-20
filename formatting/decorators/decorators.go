package decorators

// Decorator is a type of text decorator (bold, underline, etc.)
type Decorator string

var (
	Obfuscated    Decorator = "obfuscated"
	Bold          Decorator = "bold"
	Strikethrough Decorator = "strikethrough"
	Underline     Decorator = "underline"
	Italic        Decorator = "italic"
)

func (d Decorator) ToRaw() string {
	switch d {
	case Obfuscated:
		return "\u00A7k"
	case Bold:
		return "\u00A7l"
	case Strikethrough:
		return "\u00A7m"
	case Underline:
		return "\u00A7n"
	case Italic:
		return "\u00A7o"
	default:
		return ""
	}
}

// Parse attempts to return a Decorator type based on a formatting code string, formatting name string, or a Decorator type itself
func Parse(value interface{}) Decorator {
	switch value {
	case 'k', "k", "obfuscated", Obfuscated:
		return Obfuscated
	case 'l', "l", "bold", Bold:
		return Bold
	case 'm', "m", "strikethrough", Strikethrough:
		return Strikethrough
	case 'n', "n", "underline", Underline:
		return Underline
	case 'o', "o", "italic", Italic:
		return Italic
	default:
		return "unknown"
	}
}
