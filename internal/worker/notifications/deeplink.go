package notifications

import (
	"strings"

	"github.com/isyll/go-grpc-starter/pkg/config"
)

func buildDeepLink(
	c *config.NotificationsConfig,
	eventType string,
	data map[string]string,
) string {
	templateName, ok := c.EventLinks[eventType]
	if !ok {
		return ""
	}

	templatePath, ok := c.DeepLinks.Templates[templateName]
	if !ok {
		return ""
	}

	result := c.DeepLinks.BaseURL + templatePath

	for key, value := range data {
		placeholder := "{" + key + "}"
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}
