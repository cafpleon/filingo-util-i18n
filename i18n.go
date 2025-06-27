// package i18n proporciona un wrapper simplificado sobre la librería go-i18n.
package i18n

import (
	"embed"
	"fmt"
	"io/fs"
	"log/slog"

	// Importamos para manejar rutas
	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

// BundleKey define un tipo para las llaves de nuestros paquetes de traducción,
// dándonos seguridad de tipos.
type BundleKey string

// Constantes para cada paquete de traducción en nuestro ecosistema.
const (
	BundleKeySeguridad BundleKey = "seguridad"
	BundleKeyMaestros  BundleKey = "maestros"
	BundlePcaseBot     BundleKey = "pcase-bot"
	// Si mañana creas 'filingo-core-facturacion', añadirías aquí:
	// BundleKeyFacturacion BundleKey = "facturacion"
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

	// Usamos WalkDir para recorrer todos los archivos en el FS incrustado.
	// 	// Usamos WalkDir para recorrer todos los archivos en el FS incrustado.
	err := fs.WalkDir(translationsFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil // Saltamos los directorios
		}

		// --- LOG DE DEPURACIÓN CLAVE ---
		// Esto nos confirmará en la consola si los archivos se están viendo.
		slog.Debug("Intentando cargar archivo de traducción desde embed", "ruta", path)

		// Cargamos el archivo usando su ruta completa dentro del FS.
		if _, loadErr := bundle.LoadMessageFileFS(translationsFS, path); loadErr != nil {
			slog.Warn("No se pudo cargar o parsear un archivo de traducción", "ruta", path, "error", loadErr)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error recorriendo el directorio de traducciones incrustado: %w", err)
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
