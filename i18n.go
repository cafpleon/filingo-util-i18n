// package i18n proporciona un wrapper simplificado sobre la librería go-i18n.
package i18n

import (
	"embed"
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

// Bundle es el objeto principal que contiene todas las traducciones cargadas.
type Bundle struct {
	i18nBundle *i18n.Bundle
}

// New crea y configura un nuevo Bundle de i18n a partir de un sistema de archivos incrustado.
// Recibe el embed.FS y el idioma que se usará por defecto si no se encuentra una traducción.
func New(translationsFS embed.FS, defaultLang language.Tag) (*Bundle, error) {
	// 2. Creamos el "bundle" principal, definiendo el idioma por defecto.
	bundle := i18n.NewBundle(defaultLang)

	// Registramos los decodificadores para los formatos de archivo que soportaremos.
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
	bundle.RegisterUnmarshalFunc("yml", yaml.Unmarshal) // Añadimos yml como alias de yaml

	// Leemos el directorio raíz del sistema de archivos incrustado.
	files, err := fs.ReadDir(translationsFS, ".")
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer el directorio de traducciones incrustado: %w", err)
	}
	// 5. Iteramos sobre los archivos encontrados y los cargamos en el bundle.
	// Esto carga automáticamente cualquier idioma que añadas (active.es.yaml, active.fr.yaml, etc.).
	for _, file := range files {
		if !file.IsDir() {
			slog.Debug("Cargando archivo de traducción", "archivo", file.Name())
			_, err := bundle.LoadMessageFileFS(translationsFS, file.Name())
			if err != nil {
				slog.Warn("No se pudo cargar o parsear un archivo de traducción", "archivo", file.Name(), "error", err)
			}
		}
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
