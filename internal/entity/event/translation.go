package event

import "github.com/google/uuid"

const (
	TopicTranslation       = "translation"
	TypeTranslationCreated = "translation.created"
)

type TranslationCreatedPayload struct {
	TranslationID uuid.UUID `json:"translation_id"`
	UserID        uuid.UUID `json:"user_id"`
	Source        string    `json:"source"`
	Destination   string    `json:"destination"`
	Original      string    `json:"original"`
	Translation   string    `json:"translation"`
}

type TranslationCreated struct {
	Base
	payload TranslationCreatedPayload
}

func NewTranslationCreated(translationID, userID uuid.UUID, source, destination, original, translation string) *TranslationCreated {
	return &TranslationCreated{
		Base: NewBase(TypeTranslationCreated, "translation", translationID.String()),
		payload: TranslationCreatedPayload{
			TranslationID: translationID,
			UserID:        userID,
			Source:        source,
			Destination:   destination,
			Original:      original,
			Translation:   translation,
		},
	}
}

func (e *TranslationCreated) Payload() any {
	return e.payload
}
