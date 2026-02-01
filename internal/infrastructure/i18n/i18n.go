package i18n

import (
	"embed"
	"encoding/json"
	"strings"
	"sync"
)

//go:embed locales
var localesFS embed.FS

const (
	DefaultLanguage = "en"
	Vietnamese      = "vi"
	English         = "en"
)

var (
	translations = make(map[string]map[string]any)
	supportedLangs = []string{English, Vietnamese}
	mu             sync.RWMutex
)

type Translator struct {
	lang string
}

func Init(localesPath string) error {
	for _, lang := range supportedLangs {
		data, err := localesFS.ReadFile("locales/" + lang + "/messages.json")
		if err != nil {
			continue
		}

		var messages map[string]any
		if err := json.Unmarshal(data, &messages); err != nil {
			return err
		}

		mu.Lock()
		translations[lang] = messages
		mu.Unlock()
	}

	return nil
}

func LoadFromFS(fs embed.FS, basePath string) error {
	for _, lang := range supportedLangs {
		path := basePath + "/" + lang + "/messages.json"
		data, err := fs.ReadFile(path)
		if err != nil {
			continue
		}

		var messages map[string]any
		if err := json.Unmarshal(data, &messages); err != nil {
			return err
		}

		mu.Lock()
		translations[lang] = messages
		mu.Unlock()
	}

	return nil
}

func New(lang string) *Translator {
	if !isSupported(lang) {
		lang = DefaultLanguage
	}
	return &Translator{lang: lang}
}

func (t *Translator) T(key string, params ...map[string]string) string {
	return Translate(t.lang, key, params...)
}

func (t *Translator) Lang() string {
	return t.lang
}

func Translate(lang, key string, params ...map[string]string) string {
	if !isSupported(lang) {
		lang = DefaultLanguage
	}

	mu.RLock()
	langMessages, ok := translations[lang]
	mu.RUnlock()

	if !ok {
		return key
	}

	value := getNestedValue(langMessages, key)
	if value == "" {
		if lang != DefaultLanguage {
			return Translate(DefaultLanguage, key, params...)
		}
		return key
	}

	if len(params) > 0 {
		value = interpolate(value, params[0])
	}

	return value
}

func T(lang, key string, params ...map[string]string) string {
	return Translate(lang, key, params...)
}

func getNestedValue(data map[string]any, key string) string {
	parts := strings.Split(key, ".")
	current := data

	for i, part := range parts {
		if i == len(parts)-1 {
			if val, ok := current[part].(string); ok {
				return val
			}
			return ""
		}

		if next, ok := current[part].(map[string]any); ok {
			current = next
		} else {
			return ""
		}
	}

	return ""
}

func interpolate(text string, params map[string]string) string {
	for key, value := range params {
		placeholder := "{{" + key + "}}"
		text = strings.ReplaceAll(text, placeholder, value)
	}
	return text
}

func isSupported(lang string) bool {
	for _, l := range supportedLangs {
		if l == lang {
			return true
		}
	}
	return false
}

func SupportedLanguages() []string {
	return supportedLangs
}

func ParseAcceptLanguage(header string) string {
	if header == "" {
		return DefaultLanguage
	}

	langs := strings.Split(header, ",")
	for _, lang := range langs {
		lang = strings.TrimSpace(lang)
		lang = strings.Split(lang, ";")[0]
		lang = strings.Split(lang, "-")[0]
		lang = strings.ToLower(lang)

		if isSupported(lang) {
			return lang
		}
	}

	return DefaultLanguage
}
