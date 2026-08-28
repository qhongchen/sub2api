package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
)

// Task type constants
const (
	TaskTypeVerifyCode           = "verify_code"
	TaskTypePasswordReset        = "password_reset"
	TaskTypeNotification         = "notification"
	notificationEmailMaxAttempts = 3
	notificationEmailTaskTimeout = 2 * time.Minute
)

// EmailTask 邮件发送任务
type EmailTask struct {
	Email    string
	SiteName string
	TaskType string
	ResetURL string // Only used for password_reset task type
	Locale   string // Optional Accept-Language locale hint

	NotificationInput *NotificationEmailSendInput
	FallbackSubject   string
	FallbackBody      string
}

// EmailQueueService 异步邮件队列服务
type EmailQueueService struct {
	emailService *EmailService
	taskChan     chan EmailTask
	wg           sync.WaitGroup
	workers      int

	enqueueMu sync.RWMutex
	stopOnce  sync.Once
	stopped   bool
}

// NewEmailQueueService 创建邮件队列服务
func NewEmailQueueService(emailService *EmailService, workers int) *EmailQueueService {
	if workers <= 0 {
		workers = 3 // 默认3个工作协程
	}

	service := &EmailQueueService{
		emailService: emailService,
		taskChan:     make(chan EmailTask, 100), // 缓冲100个任务
		workers:      workers,
	}

	// 启动工作协程
	service.start()

	return service
}

// start 启动工作协程
func (s *EmailQueueService) start() {
	for i := 0; i < s.workers; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}
	logger.LegacyPrintf("service.email_queue", "[EmailQueue] Started %d workers", s.workers)
}

// worker 工作协程
func (s *EmailQueueService) worker(id int) {
	defer s.wg.Done()
	defer logger.LegacyPrintf("service.email_queue", "[EmailQueue] Worker %d stopping", id)

	for task := range s.taskChan {
		s.processTask(id, task)
	}
}

// processTask 处理任务
func (s *EmailQueueService) processTask(workerID int, task EmailTask) {
	timeout := 30 * time.Second
	if task.TaskType == TaskTypeNotification {
		timeout = notificationEmailTaskTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	switch task.TaskType {
	case TaskTypeVerifyCode:
		if err := s.emailService.SendVerifyCode(ctx, task.Email, task.SiteName, task.Locale); err != nil {
			logger.LegacyPrintf("service.email_queue", "[EmailQueue] Worker %d failed to send verify code to %s: %v", workerID, task.Email, err)
		} else {
			logger.LegacyPrintf("service.email_queue", "[EmailQueue] Worker %d sent verify code to %s", workerID, task.Email)
		}
	case TaskTypePasswordReset:
		if err := s.emailService.SendPasswordResetEmailWithCooldown(ctx, task.Email, task.SiteName, task.ResetURL, task.Locale); err != nil {
			logger.LegacyPrintf("service.email_queue", "[EmailQueue] Worker %d failed to send password reset to %s: %v", workerID, task.Email, err)
		} else {
			logger.LegacyPrintf("service.email_queue", "[EmailQueue] Worker %d sent password reset to %s", workerID, task.Email)
		}
	case TaskTypeNotification:
		if task.NotificationInput == nil || s.emailService == nil {
			slog.Error("email_queue: notification email service is not configured", "worker_id", workerID)
			return
		}
		input := *task.NotificationInput
		if err := s.sendNotificationWithRetry(ctx, input, task.FallbackSubject, task.FallbackBody); err != nil {
			slog.Error("email_queue: notification email delivery failed",
				"worker_id", workerID,
				"event", input.Event,
				"source_type", input.SourceType,
				"source_id", input.SourceID,
				"recipient_hash", notificationEmailHash(input.RecipientEmail),
				"error", err)
		}
	default:
		logger.LegacyPrintf("service.email_queue", "[EmailQueue] Worker %d unknown task type: %s", workerID, task.TaskType)
	}
}

func (s *EmailQueueService) sendNotificationWithRetry(ctx context.Context, input NotificationEmailSendInput, fallbackSubject, fallbackBody string) error {
	maxAttempts := notificationEmailMaxAttempts
	channelStatus := isChannelStatusDelivery(input)
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		lastErr = s.sendNotification(ctx, input, fallbackSubject, fallbackBody)
		if lastErr == nil {
			return nil
		}
		if !isNotificationEmailDeliveryError(lastErr) || attempt == maxAttempts {
			return lastErr
		}
		if channelStatus {
			// 渠道状态通知只重试已确认尚未进入 SMTP DATA 的失败；写入或
			// 关闭 DATA 流后的超时结果未知，绝不能自动重发。
			if !isNotificationEmailDeliveryBeforeDataError(lastErr) {
				return lastErr
			}
		}
		backoff := time.Duration(attempt) * time.Second
		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return ctx.Err()
		case <-timer.C:
		}
	}
	return lastErr
}

func (s *EmailQueueService) sendNotification(ctx context.Context, input NotificationEmailSendInput, fallbackSubject, fallbackBody string) error {
	if s.emailService.notificationEmailService != nil {
		return s.emailService.notificationEmailService.SendWithFallback(ctx, input, fallbackSubject, fallbackBody)
	}
	if isChannelStatusDelivery(input) {
		return notificationEmailConfigErr(fmt.Errorf("channel status notification service is not configured"))
	}
	if err := s.emailService.SendEmail(ctx, input.RecipientEmail, fallbackSubject, fallbackBody); err != nil {
		return notificationEmailDeliveryErr(err)
	}
	return nil
}

// EnqueueNotification 异步投递模板通知，并复制变量避免调用方后续修改任务内容。
func (s *EmailQueueService) EnqueueNotification(input NotificationEmailSendInput, fallbackSubject, fallbackBody string) error {
	input.Variables = cloneEmailVariables(input.Variables)
	input.RawHTMLVariables = cloneEmailVariables(input.RawHTMLVariables)
	task := EmailTask{
		Email:             input.RecipientEmail,
		TaskType:          TaskTypeNotification,
		NotificationInput: &input,
		FallbackSubject:   fallbackSubject,
		FallbackBody:      fallbackBody,
	}

	return s.enqueue(task)
}

func cloneEmailVariables(input map[string]string) map[string]string {
	if input == nil {
		return nil
	}
	output := make(map[string]string, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}

// EnqueueVerifyCode 将验证码发送任务加入队列
func (s *EmailQueueService) EnqueueVerifyCode(email, siteName string, locale ...string) error {
	task := EmailTask{
		Email:    email,
		SiteName: siteName,
		TaskType: TaskTypeVerifyCode,
		Locale:   firstEmailLocale(locale),
	}

	if err := s.enqueue(task); err != nil {
		return err
	}
	logger.LegacyPrintf("service.email_queue", "[EmailQueue] Enqueued verify code task for %s", email)
	return nil
}

// EnqueuePasswordReset 将密码重置邮件任务加入队列
func (s *EmailQueueService) EnqueuePasswordReset(email, siteName, resetURL string, locale ...string) error {
	task := EmailTask{
		Email:    email,
		SiteName: siteName,
		TaskType: TaskTypePasswordReset,
		ResetURL: resetURL,
		Locale:   firstEmailLocale(locale),
	}

	if err := s.enqueue(task); err != nil {
		return err
	}
	logger.LegacyPrintf("service.email_queue", "[EmailQueue] Enqueued password reset task for %s", email)
	return nil
}

func (s *EmailQueueService) enqueue(task EmailTask) error {
	if s == nil || s.taskChan == nil {
		return fmt.Errorf("email queue is not configured")
	}
	s.enqueueMu.RLock()
	defer s.enqueueMu.RUnlock()
	if s.stopped {
		return fmt.Errorf("email queue is stopped")
	}
	select {
	case s.taskChan <- task:
		return nil
	default:
		return fmt.Errorf("email queue is full")
	}
}

// Stop 停止队列服务
func (s *EmailQueueService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		s.enqueueMu.Lock()
		s.stopped = true
		if s.taskChan != nil {
			close(s.taskChan)
		}
		s.enqueueMu.Unlock()

		s.wg.Wait()
		logger.LegacyPrintf("service.email_queue", "%s", "[EmailQueue] All workers stopped")
	})
}
