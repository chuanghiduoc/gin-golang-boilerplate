package middleware

import (
	"github.com/gin-gonic/gin"

	httpctx "backend-gin/internal/adapter/handler/http/context"
	"backend-gin/internal/infrastructure/i18n"
)

const TranslatorKey = "translator"

func I18n() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.Query("lang")
		if lang == "" {
			lang = c.GetHeader("Accept-Language")
		}

		parsedLang := i18n.ParseAcceptLanguage(lang)
		translator := i18n.New(parsedLang)

		c.Set(string(httpctx.LangKey), parsedLang)
		c.Set(TranslatorKey, translator)

		c.Next()
	}
}

func GetTranslator(c *gin.Context) *i18n.Translator {
	if t, exists := c.Get(TranslatorKey); exists {
		if translator, ok := t.(*i18n.Translator); ok {
			return translator
		}
	}
	return i18n.New(i18n.DefaultLanguage)
}

func GetLang(c *gin.Context) string {
	return httpctx.GetLang(c)
}
