package decorators

// Decorator is a type of text decorator (bold, underline, etc.).
type Decorator string

var (
	// Obfuscated is a decorator used to specify text is obfuscated (§k).
	Obfuscated Decorator = "obfuscated"
	// Bold is a decorator used to specify text is bold (§k).
	Bold Decorator = "bold"
	// Strikethrough is a decorator used to specify text has a strikethrough (§m).
	Strikethrough Decorator = "strikethrough"
	// Underlined is a decorator used to specify text has an underline (§n).
	Underlined Decorator = "underline"
	// Italic is a decorator used to specify text uses italics (§o).
	Italic Decorator = "italic"
	// Unknown is an unknown parsed decorator.
	Unknown Decorator = "unknown"
)

var (
	// PropertyMap is a key-value map of Minecraft property names to decorator types.
	PropertyMap map[string]Decorator = map[string]Decorator{
		"obfuscated":    Obfuscated,
		"bold":          Bold,
		"strikethrough": Strikethrough,
		"underlined":    Underlined,
		"italic":        Italic,
	}
)

// ToRaw returns the Minecraft formatted text of the decorator (§ + code).
func (d Decorator) ToRaw() string {
	switch d {
	case Obfuscated:
		return "\u00A7k"
	case Bold:
		return "\u00A7l"
	case Strikethrough:
		return "\u00A7m"
	case Underlined:
		return "\u00A7n"
	case Italic:
		return "\u00A7o"
	default:
		return ""
	}
}

// Parse attempts to return a Decorator type based on a formatting code string, formatting name string, or a Decorator type itself.
func Parse(value interface{}) (Decorator, bool) {
	switch value {
	case 'k', "k", "obfuscated", Obfuscated:
		return Obfuscated, true
	case 'l', "l", "bold", Bold:
		return Bold, true
	case 'm', "m", "strikethrough", Strikethrough:
		return Strikethrough, true
	case 'n', "n", "underline", Underlined:
		return Underlined, true
	case 'o', "o", "italic", Italic:
		return Italic, true
	default:
		return Unknown, false
	}
}
