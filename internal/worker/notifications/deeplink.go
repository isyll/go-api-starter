package notifications

import (
	"strings"

	"github.com/isyll/go-api-starter/pkg/config"
)

// buildDeepLink constructs the deep link URL for a notification
// event. The notifications config carries a two-tier mapping:
// event_type -> template_name -> URL template. Missing entries at
// either tier return the empty string, signaling the push payload
// should not carry a click_action URL.
//
// Placeholders inside the template ({trip_id}, {booking_id}, ...)
// are substituted from data using literal {key} -> value
// replacement.
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
