// package i18n proporciona un wrapper simplificado sobre la librería go-i18n.
package i18n

import (
	"fmt"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

// Bundle es el objeto principal que contiene todas las traducciones cargadas.
type Bundle struct {
	i18nBundle *i18n.Bundle
}

// New crea y configura un nuevo Bundle de i18n.
// Recibe la ruta al directorio que contiene los archivos de traducción (ej: "./i18n").
func New(translationsPath string) (*Bundle, error) {
	// 1. Creamos el "bundle" principal, definiendo el idioma por defecto.
	bundle := i18n.NewBundle(language.Spanish) // Español como fallback

	// 2. Le decimos al bundle cómo "decodificar" nuestros archivos de traducción.
	// Usaremos TOML, pero también podríamos registrar JSON.
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)

	// 3. Cargamos los archivos de traducción desde la ruta especificada.
	// Ahora cargamos archivos que terminan en .yaml (o .yml).
	// La librería buscará archivos que coincidan con el patrón "active.*.yaml".
	_, err := bundle.LoadMessageFile(fmt.Sprintf("%s/active.es.yaml", translationsPath))
	if err != nil {
		return nil, fmt.Errorf("error cargando archivo de traducción para español: %w", err)
	}
	_, err = bundle.LoadMessageFile(fmt.Sprintf("%s/active.en.yaml", translationsPath))
	if err != nil {
		return nil, fmt.Errorf("error cargando archivo de traducción para inglés: %w", err)
	}

	return &Bundle{i18nBundle: bundle}, nil
}

// GetLocalizer crea un "localizador" para un idioma específico.
// Recibe los idiomas deseados en orden de preferencia (ej: "es-CO", "es", "en-US").
func (b *Bundle) GetLocalizer(langs ...string) *i18n.Localizer {
	return i18n.NewLocalizer(b.i18nBundle, langs...)
}

// T es un helper para traducir un texto de forma rápida.
// Recibe un localizador, el ID del mensaje y datos opcionales para plantillas.
// Ejemplo de uso: T(localizador, "WelcomeMessage", map[string]string{"Name": "Carlos"})
func T(localizer *i18n.Localizer, messageID string, templateData any, pluralCount ...int) string {
	pc := -1
	if len(pluralCount) > 0 {
		pc = pluralCount[0]
	}

	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: templateData,
		PluralCount:  pc,
	})
	if err != nil {
		// En un caso real, loggearías este error.
		// Devolvemos el ID del mensaje para que el desarrollador sepa qué falló.
		return messageID
	}

	return msg
}
