package notification

import (
	"context"
	"fmt"

	"github.com/evrone/go-clean-template/internal/entity/notification"
	"github.com/evrone/go-clean-template/internal/repo"
	"github.com/evrone/go-clean-template/pkg/notify"
	"github.com/google/uuid"
)

type Service struct {
	notificationRepo repo.NotificationRepo
	prefsRepo        repo.NotificationPreferencesRepo
	pushTokenRepo    repo.PushTokenRepo
	deliveryLogRepo  repo.DeliveryLogRepo
	emailSender      notify.EmailSender
	pushSender       notify.PushSender
}

type ServiceDeps struct {
	NotificationRepo repo.NotificationRepo
	PrefsRepo        repo.NotificationPreferencesRepo
	PushTokenRepo    repo.PushTokenRepo
	DeliveryLogRepo  repo.DeliveryLogRepo
	EmailSender      notify.EmailSender
	PushSender       notify.PushSender
}

func NewService(deps *ServiceDeps) *Service {
	return &Service{
		notificationRepo: deps.NotificationRepo,
		prefsRepo:        deps.PrefsRepo,
		pushTokenRepo:    deps.PushTokenRepo,
		deliveryLogRepo:  deps.DeliveryLogRepo,
		emailSender:      deps.EmailSender,
		pushSender:       deps.PushSender,
	}
}

func (s *Service) SendInApp(ctx context.Context, msg *notification.InAppMessage) error {
	prefs, err := s.prefsRepo.Get(ctx, msg.UserID)
	if err == nil && !prefs.InAppEnabled {
		return nil
	}

	n := &notification.InAppNotification{
		UserID: msg.UserID,
		Type:   msg.Type,
		Title:  msg.Title,
		Body:   msg.Body,
		Data:   msg.Data,
		Read:   false,
	}

	if msg.ActionURL != "" {
		n.ActionURL = &msg.ActionURL
	}

	if msg.ImageURL != "" {
		n.ImageURL = &msg.ImageURL
	}

	if err := s.notificationRepo.Store(ctx, n); err != nil {
		return fmt.Errorf("Service - SendInApp - s.notificationRepo.Store: %w", err)
	}

	return nil
}

func (s *Service) SendPush(ctx context.Context, msg *notification.PushMessage) error {
	prefs, err := s.prefsRepo.Get(ctx, msg.UserID)
	if err == nil && !prefs.PushEnabled {
		return nil
	}

	tokens, err := s.pushTokenRepo.GetByUserID(ctx, msg.UserID)
	if err != nil {
		return fmt.Errorf("Service - SendPush - s.pushTokenRepo.GetByUserID: %w", err)
	}

	if len(tokens) == 0 {
		return nil
	}

	tokenStrings := make([]string, 0, len(tokens))
	for i := range tokens {
		if tokens[i].Active {
			tokenStrings = append(tokenStrings, tokens[i].Token)
		}
	}

	if len(tokenStrings) == 0 {
		return nil
	}

	if s.pushSender == nil {
		return nil
	}

	if err := s.pushSender.Send(ctx, msg, tokenStrings); err != nil {
		s.logDelivery(ctx, uuid.Nil, msg.UserID, notification.ChannelPush, notification.StatusFailed, err.Error())

		return fmt.Errorf("Service - SendPush - s.pushSender.Send: %w", err)
	}

	s.logDelivery(ctx, uuid.Nil, msg.UserID, notification.ChannelPush, notification.StatusSent, "")

	return nil
}

func (s *Service) SendEmail(ctx context.Context, msg *notification.EmailMessage) error {
	if s.emailSender == nil {
		return nil
	}

	if err := s.emailSender.Send(ctx, msg); err != nil {
		return fmt.Errorf("Service - SendEmail - s.emailSender.Send: %w", err)
	}

	return nil
}

func (s *Service) logDelivery(ctx context.Context, notificationID, userID uuid.UUID, channel notification.Channel, status notification.Status, errMsg string) {
	log := &notification.DeliveryLog{
		NotificationID: notificationID,
		UserID:         userID,
		Channel:        channel,
		Status:         status,
		Attempts:       1,
	}

	if errMsg != "" {
		log.ErrorMessage = &errMsg
	}

	//nolint:errcheck // fire and forget - delivery logging should not fail the main operation
	s.deliveryLogRepo.Store(ctx, log)
}
