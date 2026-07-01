-- Notification templates and their per-language translations.
-- Deep links are managed via application configuration
-- (configs/notifications.yaml) so URLs can vary per environment
-- (dev, staging, production) without a migration. The seed INSERTs
-- below populate the canonical template set in en and fr.

CREATE TABLE IF NOT EXISTS
notifications.notification_templates (
  id SERIAL PRIMARY KEY,
  event_type VARCHAR(50) NOT NULL UNIQUE,
  icon VARCHAR(100),
  sound VARCHAR(100) DEFAULT 'default',
  priority VARCHAR(20) NOT NULL DEFAULT 'high',
  android_channel_id VARCHAR(50),
  action VARCHAR(100),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT chk_notification_templates_priority_value CHECK (priority IN ('high', 'normal', 'low'))
);

CREATE TABLE IF NOT EXISTS
notifications.notification_template_translations (
  id SERIAL PRIMARY KEY,
  template_id INT NOT NULL REFERENCES notifications.notification_templates (id) ON DELETE CASCADE,
  language CHAR(2) NOT NULL CHECK (language ~ '^[a-z]{2}$'),
  title TEXT NOT NULL CHECK (title <> ''),
  body TEXT NOT NULL CHECK (body <> ''),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (template_id, language)
);

CREATE INDEX IF NOT EXISTS idx_notification_templates_event_type ON notifications.notification_templates (event_type);

CREATE INDEX IF NOT EXISTS idx_notification_templates_priority ON notifications.notification_templates (priority);

CREATE INDEX IF NOT EXISTS idx_notification_template_translations_template_id ON notifications.notification_template_translations (template_id);

CREATE INDEX IF NOT EXISTS idx_notification_template_translations_language ON notifications.notification_template_translations (language);

-- Lifecycle triggers: updated_at maintenance and created_at
-- immutability on both tables.
DROP TRIGGER IF EXISTS update_notification_templates_updated_at ON notifications.notification_templates;

CREATE TRIGGER update_notification_templates_updated_at BEFORE
UPDATE ON notifications.notification_templates FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS prevent_notification_templates_created_at_update ON notifications.notification_templates;

CREATE TRIGGER prevent_notification_templates_created_at_update BEFORE
UPDATE ON notifications.notification_templates FOR EACH ROW
EXECUTE FUNCTION prevent_update_created_at();

DROP TRIGGER IF EXISTS update_notification_template_translations_updated_at ON notifications.notification_template_translations;

CREATE TRIGGER update_notification_template_translations_updated_at BEFORE
UPDATE ON notifications.notification_template_translations FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS prevent_notification_template_translations_created_at_update ON notifications.notification_template_translations;

CREATE TRIGGER prevent_notification_template_translations_created_at_update BEFORE
UPDATE ON notifications.notification_template_translations FOR EACH ROW
EXECUTE FUNCTION prevent_update_created_at();

-- Seed templates. ON CONFLICT (event_type) DO NOTHING keeps the
-- migration idempotent.
INSERT INTO
notifications.notification_templates (event_type, priority, android_channel_id, action)
VALUES
(
  'booking.accepted',
  'high',
  'bookings',
  'open_trip_details'
),
(
  'booking.rejected',
  'normal',
  'bookings',
  'view_other_trips'
),
(
  'booking.canceled',
  'high',
  'bookings',
  'find_alternative'
),
(
  'booking.created',
  'high',
  'bookings',
  'review_booking'
),
(
  'trip.started',
  'high',
  'trips',
  'open_live_tracking'
),
(
  'trip.completed',
  'normal',
  'trips',
  'open_rating'
),
(
  'trip.canceled',
  'high',
  'trips',
  'find_alternative'
),
(
  'trip.reminder.1h',
  'normal',
  'reminders',
  'view_trip_details'
),
(
  'trip.reminder.30m',
  'high',
  'reminders',
  'get_ready'
),
('message.new', 'normal', 'messages', 'open_chat'),
(
  'rating.received',
  'low',
  'ratings',
  'view_rating'
),
(
  'trip.seats_full',
  'normal',
  'trips',
  'view_bookings'
),
(
  'trip.price_changed',
  'high',
  'trips',
  'view_trip_details'
) ON CONFLICT (event_type)
DO NOTHING;

-- Seed English translations. Body strings carry {placeholder}
-- tokens that the worker substitutes at dispatch time.
INSERT INTO
notifications.notification_template_translations (template_id, language, title, body)
SELECT
  id,
  'en',
  title,
  body
FROM
  (
    VALUES
    (
      'booking.accepted',
      'Booking Accepted',
      '{driver} accepted your booking for {route}'
    ),
    (
      'booking.rejected',
      'Booking Declined',
      '{driver} declined your booking request'
    ),
    (
      'booking.canceled',
      'Booking Canceled',
      'Your booking for {route} has been canceled'
    ),
    (
      'booking.created',
      'New Booking Request',
      '{passenger} wants to join your trip to {destination}'
    ),
    (
      'trip.started',
      'Trip Started',
      'Your trip has started. Track your driver now'
    ),
    (
      'trip.completed',
      'Trip Completed',
      'You have arrived at your destination. Please rate your driver'
    ),
    (
      'trip.canceled',
      'Trip Canceled',
      '{driver} canceled the trip to {destination}'
    ),
    (
      'trip.reminder.1h',
      'Trip Reminder',
      'Your trip starts in 1 hour'
    ),
    (
      'trip.reminder.30m',
      'Departure Soon',
      'Your trip departs in 30 minutes'
    ),
    (
      'message.new',
      'New Message',
      '{sender}: {message_preview}'
    ),
    (
      'rating.received',
      'New Rating',
      '{rater} rated you {stars} stars'
    ),
    (
      'trip.seats_full',
      'Trip Full',
      'All seats are now booked'
    ),
    (
      'trip.price_changed',
      'Trip Price Updated',
      '{driver} changed the price for {route} from {old_price} to {new_price}'
    )
  ) AS t (event_type, title, body)
INNER JOIN notifications.notification_templates AS nt ON t.event_type = nt.event_type ON CONFLICT (template_id, language)
DO NOTHING;

-- Seed French translations.
INSERT INTO
notifications.notification_template_translations (template_id, language, title, body)
SELECT
  id,
  'fr',
  title,
  body
FROM
  (
    VALUES
    (
      'booking.accepted',
      'Réservation acceptée',
      '{driver} a accepté votre réservation pour {route}'
    ),
    (
      'booking.rejected',
      'Réservation refusée',
      '{driver} a refusé votre demande de réservation'
    ),
    (
      'booking.canceled',
      'Réservation annulée',
      'Votre réservation pour {route} a été annulée'
    ),
    (
      'booking.created',
      'Nouvelle demande de réservation',
      '{passenger} souhaite rejoindre votre trajet vers {destination}'
    ),
    (
      'trip.started',
      'Trajet commencé',
      'Votre trajet a commencé. Suivez votre conducteur maintenant'
    ),
    (
      'trip.completed',
      'Trajet terminé',
      'Vous êtes arrivé à destination. Veuillez noter votre conducteur'
    ),
    (
      'trip.canceled',
      'Trajet annulé',
      '{driver} a annulé le trajet vers {destination}'
    ),
    (
      'trip.reminder.1h',
      'Rappel de trajet',
      'Votre trajet commence dans 1 heure'
    ),
    (
      'trip.reminder.30m',
      'Départ imminent',
      'Votre trajet part dans 30 minutes'
    ),
    (
      'message.new',
      'Nouveau message',
      '{sender} : {message_preview}'
    ),
    (
      'rating.received',
      'Nouvelle note',
      '{rater} vous a donné {stars} étoiles'
    ),
    (
      'trip.seats_full',
      'Trajet complet',
      'Tous les sièges sont maintenant réservés'
    ),
    (
      'trip.price_changed',
      'Prix du trajet mis à jour',
      '{driver} a modifié le prix du trajet {route} de {old_price} à {new_price}'
    )
  ) AS t (event_type, title, body)
INNER JOIN notifications.notification_templates AS nt ON t.event_type = nt.event_type ON CONFLICT (template_id, language)
DO NOTHING;

COMMENT ON TABLE notifications.notification_templates IS 'Notification templates with language-agnostic metadata. Deep links are configured in application config files (configs/notifications.yaml).';

COMMENT ON TABLE notifications.notification_template_translations IS 'Per-language translations for notification templates. UNIQUE (template_id, language) keeps one translation per pair.';

COMMENT ON COLUMN notifications.notification_templates.event_type IS 'Unique event-type identifier (e.g. booking.accepted). Mirrors notification_logs.event_type.';

COMMENT ON COLUMN notifications.notification_templates.icon IS 'Icon resource name shown by the client when the notification is rendered.';

COMMENT ON COLUMN notifications.notification_templates.sound IS 'Sound asset to play when the notification is received.';

COMMENT ON COLUMN notifications.notification_templates.priority IS 'Notification priority: high, normal, or low. CHECK enforces the closed set.';

COMMENT ON COLUMN notifications.notification_templates.android_channel_id IS 'Android notification channel ID used to route the alert into the correct user-facing category.';

COMMENT ON COLUMN notifications.notification_templates.action IS 'Client-side action key to perform when the notification is clicked.';

COMMENT ON COLUMN notifications.notification_template_translations.template_id IS 'Reference to the parent notification template. CASCADE delete removes translations when the template is deleted.';

COMMENT ON COLUMN notifications.notification_template_translations.language IS 'ISO 639-1 two-letter language code (e.g. en, fr, ar).';

COMMENT ON COLUMN notifications.notification_template_translations.title IS 'Notification title in this language.';

COMMENT ON COLUMN notifications.notification_template_translations.body IS 'Notification body in this language. May contain {placeholder} tokens substituted at dispatch time.';
