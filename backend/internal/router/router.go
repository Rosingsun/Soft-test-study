package router

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/soft-test-study/backend/internal/config"
	"github.com/soft-test-study/backend/internal/handler"
	"github.com/soft-test-study/backend/internal/middleware"
	"github.com/soft-test-study/backend/internal/repository"
	"github.com/soft-test-study/backend/internal/service"
	"gorm.io/gorm"
)

func Setup(db *gorm.DB, r *gin.Engine, cfg *config.Config, ctx context.Context) {
	api := r.Group("/api/v1")

	userRepo := repository.NewUserRepo(db)
	examLevelRepo := repository.NewExamLevelRepo(db)
	subjectRepo := repository.NewSubjectRepo(db)
	subSubjectRepo := repository.NewSubSubjectRepo(db)
	chapterRepo := repository.NewChapterRepo(db)
	questionRepo := repository.NewQuestionRepo(db)
	practiceRecordRepo := repository.NewPracticeRecordRepo(db)
	bookmarkRepo := repository.NewBookmarkRepo(db)
	markRepo := repository.NewQuestionMarkRepo(db)
	wrongRepo := repository.NewWrongQuestionRepo(db)
	examTemplateRepo := repository.NewExamTemplateRepo(db)
	examRecordRepo := repository.NewExamRecordRepo(db)
	aiRepo := repository.NewAiRepo(db)
	// 修复 #20：AI 异步出题任务持久化仓库（DB 为权威存储，内存仅作热缓存）
	aiTaskRepo := repository.NewAiGeneratedTaskRepo(db)
	essayScoreRepo := repository.NewEssayScoreRepo(db)
	studyMaterialRepo := repository.NewStudyMaterialRepo(db)
	checkInRepo := repository.NewCheckInRepo(db)
	reviewCardRepo := repository.NewReviewCardRepo(db)
	studyPlanRepo := repository.NewStudyPlanRepo(db)
	rankingRepo := repository.NewRankingRepo(db)
	emailCodeRepo := repository.NewEmailVerificationCodeRepo(db)
	invitationCodeRepo := repository.NewInvitationCodeRepo(db)

	notifyRepo := repository.NewNotificationRepo(db)
	notifySvc := service.NewNotificationService(notifyRepo)
	knowledgePointRepo := repository.NewKnowledgePointRepo(db)

	invitationCodeSvc := service.NewInvitationCodeService(invitationCodeRepo)
	userSvc := service.NewUserService(
		db,
		userRepo, examLevelRepo, subjectRepo,
		invitationCodeSvc,
		cfg.JWTSecret, cfg.JWTExpiresIn,
		cfg.AdminBypassUsernames,
	)
	examLevelSvc := service.NewExamLevelService(examLevelRepo)
	subjectSvc := service.NewSubjectService(subjectRepo)
	subSubjectSvc := service.NewSubSubjectService(subSubjectRepo)
	chapterSvc := service.NewChapterService(chapterRepo, questionRepo)
	questionSvc := service.NewQuestionService(questionRepo)
	reviewSvc := service.NewReviewService(reviewCardRepo, questionRepo, practiceRecordRepo, wrongRepo)
	practiceSvc := service.NewPracticeService(practiceRecordRepo, questionRepo, reviewSvc)
	bookmarkSvc := service.NewBookmarkService(bookmarkRepo)
	markSvc := service.NewQuestionMarkService(markRepo)
	wrongSvc := service.NewWrongQuestionService(wrongRepo, questionRepo, practiceRecordRepo, reviewSvc)
	examSvc := service.NewExamService(db, examTemplateRepo, examRecordRepo, questionRepo, reviewSvc)
	// 注册并启动考试超时 janitor（每 30s 扫描 pending 记录并自动交卷）
	// OPT-20: 接收根 ctx，便于服务关闭时退出
	service.SetExamJanitor(examSvc)
	service.StartExamJanitor(ctx)
	statsRepo := repository.NewStatsRepo(db)
	statsSvc := service.NewStatsService(statsRepo, subjectRepo)
	// OPT-19: AI 任务管理改为通过 AiTaskManager 依赖注入（替代旧包级 var）
	aiTaskMgr := service.NewAiTaskManager(aiTaskRepo)
	aiSvc := service.NewAiService(aiRepo, aiTaskMgr, questionRepo, subjectRepo, chapterRepo, examRecordRepo, essayScoreRepo, practiceRecordRepo, notifySvc)
	// 注册 AI 出题超时通知回调（janitor 标记失败后调用）
	aiTaskMgr.SetNotifier(func(userID uint, taskID, msg string) {
		short := msg
		if len(short) > 200 {
			short = short[:200] + "..."
		}
		if err := notifySvc.Push(userID, "ai_generate_failed", "AI 出题失败",
			"任务超时："+short+"。请稍后重试或调整参数。",
			"/ai/practice?task_id="+taskID); err != nil {
			log.Printf("[ai task] timeout notifier push failed user_id=%d task_id=%s: %v",
				userID, taskID, err)
		}
	})
	// 启动 AI 出题超时 janitor（每 1min 扫描 pending > 5min / running > 30min 的任务并标记 failed）
	// OPT-20: 接收根 ctx 便于服务关闭时退出
	aiTaskMgr.StartJanitor(ctx)
	// 启动时恢复 in-flight 任务（把 DB 中 status='running' 的行加回内存）
	if err := aiSvc.RecoverInflightTasks(); err != nil {
		log.Printf("[ai task] recover inflight failed: %v", err)
	}
	studyMaterialSvc := service.NewStudyMaterialService(studyMaterialRepo, subjectRepo)
	checkInSvc := service.NewCheckInService(checkInRepo, questionRepo, practiceRecordRepo, reviewSvc)
	studyPlanSvc := service.NewStudyPlanService(studyPlanRepo, practiceRecordRepo)
	rankingSvc := service.NewRankingService(rankingRepo, userRepo)
	knowledgePointSvc := service.NewKnowledgePointService(knowledgePointRepo, questionRepo, subjectRepo)

	// 管理端服务：复用现有 user / stats / exam 仓储，不引入新数据源
	adminSvc := service.NewAdminService(
		userRepo, examLevelRepo, subjectRepo,
		invitationCodeRepo, invitationCodeSvc, statsSvc,
		examRecordRepo, examTemplateRepo,
		questionRepo, subSubjectRepo, chapterRepo,
	)

	// 邮件发送器选择顺序：
	//   1. MAIL_PROVIDER=resend 或已配置 RESEND_API_KEY → Resend HTTP API（无需邮箱授权码）
	//   2. MAIL_PROVIDER=smtp 或已配置完整 SMTP → SMTP 直连
	//   3. 都未配置 → 降级为日志发送器（开发模式，验证码打进日志）
	var emailSender service.EmailSender
	provider := strings.ToLower(strings.TrimSpace(cfg.MailProvider))
	hasResend := cfg.ResendAPIKey != ""
	hasSMTP := cfg.SMTPHost != "" && cfg.SMTPUser != "" && cfg.SMTPPassword != ""

	switch {
	case provider == "resend" || (provider == "" && hasResend):
		if !hasResend {
			log.Println("[email] MAIL_PROVIDER=resend 但未配置 RESEND_API_KEY，降级为日志发送器")
			emailSender = service.NewLogEmailSender()
			break
		}
		emailSender = service.NewResendEmailSender(cfg.ResendAPIKey, cfg.ResendFrom)
		log.Printf("[email] 使用 Resend API 发送器 from=%s", cfg.ResendFrom)
	case provider == "smtp" || (provider == "" && hasSMTP):
		if !hasSMTP {
			log.Println("[email] MAIL_PROVIDER=smtp 但 SMTP 配置不完整，降级为日志发送器")
			emailSender = service.NewLogEmailSender()
			break
		}
		emailSender = service.NewSMTPEmailSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPUser, cfg.SMTPFromName)
		log.Printf("[email] 使用 SMTP 发送器 host=%s port=%d", cfg.SMTPHost, cfg.SMTPPort)
	default:
		emailSender = service.NewLogEmailSender()
		log.Println("[email] 未配置 Resend / SMTP，验证码将仅打印到日志（开发模式）")
	}
	emailSvc := service.NewEmailService(emailCodeRepo, userRepo, emailSender, cfg.EmailDailyMax)

	userH := handler.NewUserHandler(userSvc, emailSvc)
	examLevelH := handler.NewExamLevelHandler(examLevelSvc)
	subjectH := handler.NewSubjectHandler(subjectSvc)
	subSubjectH := handler.NewSubSubjectHandler(subSubjectSvc)
	chapterH := handler.NewChapterHandler(chapterSvc)
	questionH := handler.NewQuestionHandler(questionSvc)
	practiceH := handler.NewPracticeHandler(practiceSvc)
	bookmarkH := handler.NewBookmarkHandler(bookmarkSvc)
	markH := handler.NewQuestionMarkHandler(markSvc)
	wrongH := handler.NewWrongQuestionHandler(wrongSvc)
	examH := handler.NewExamHandler(examSvc)
	statsH := handler.NewStatsHandler(statsSvc)
	aiH := handler.NewAiHandler(aiSvc)
	studyMaterialH := handler.NewStudyMaterialHandler(studyMaterialSvc)
	checkInH := handler.NewCheckInHandler(checkInSvc)
	reviewH := handler.NewReviewHandler(reviewSvc)
	studyPlanH := handler.NewStudyPlanHandler(studyPlanSvc)
	rankingH := handler.NewRankingHandler(rankingSvc)
	notifyH := handler.NewNotificationHandler(notifySvc)
	knowledgePointH := handler.NewKnowledgePointHandler(knowledgePointSvc)
	adminH := handler.NewAdminHandler(adminSvc)

	rateLimiter := middleware.RateLimit(5, time.Minute)

	api.POST("/auth/register", rateLimiter, userH.Register)
	api.POST("/auth/login", rateLimiter, userH.Login)
	// 发送邮箱验证码：需要登录态（避免匿名刷验证码），加限流防刷
	api.POST("/auth/email/send-code", middleware.Auth(cfg.JWTSecret), rateLimiter, userH.SendEmailCode)
	// 未登录重置密码：公开接口，加限流防刷
	api.POST("/auth/password/reset-code", rateLimiter, userH.SendResetPasswordCode)
	api.POST("/auth/password/reset", rateLimiter, userH.ResetPasswordByCode)

	api.GET("/exam-levels", examLevelH.List)
	api.GET("/subjects", subjectH.ListByLevel)
	api.GET("/subjects/all", subjectH.ListAll)
	api.GET("/subjects/:id/sub-subjects", subSubjectH.ListBySubject)
	api.GET("/sub-subjects/:id/chapters", chapterH.ListBySubSubject)
	api.GET("/questions", questionH.ListByChapter)
	api.GET("/questions/case-studies", questionH.CaseStudies)
	api.GET("/exam-templates", examH.ListTemplates)
	api.GET("/ai/providers", aiH.GetProviders)
	api.GET("/materials", studyMaterialH.List)
	api.GET("/materials/:id", studyMaterialH.Detail)
	api.GET("/materials/:id/content", studyMaterialH.Content)

	auth := api.Group("")
	auth.Use(middleware.Auth(cfg.JWTSecret))
	{
		auth.GET("/auth/user-info", userH.GetUserInfo)
		auth.PUT("/auth/profile", userH.UpdateProfile)
		auth.PUT("/auth/password", userH.ChangePassword)
		auth.POST("/auth/email/verify", userH.VerifyEmailCode)

		auth.POST("/practice/submit", practiceH.Submit)

		auth.GET("/questions/random", questionH.Random)
		auth.GET("/questions/special", questionH.Special)
		auth.GET("/questions/essays", questionH.EssayList)
		auth.GET("/questions/:id", questionH.GetByID)

		auth.POST("/favorite-folders", bookmarkH.CreateFolder)
		auth.GET("/favorite-folders", bookmarkH.ListFolders)
		auth.POST("/questions/:id/favorite", bookmarkH.AddFavorite)
		auth.DELETE("/questions/:id/favorite", bookmarkH.RemoveFavorite)
		auth.GET("/questions/:id/favorited", bookmarkH.CheckFavorited)
		auth.GET("/favorites", bookmarkH.ListFavorites)

		auth.POST("/questions/:id/mark", markH.AddMark)
		auth.DELETE("/questions/:id/mark", markH.RemoveMark)
		auth.GET("/questions/:id/marked", markH.CheckMarked)
		auth.GET("/marks", markH.ListMarks)

		auth.GET("/wrong-questions", wrongH.List)
		auth.GET("/wrong-questions/count", wrongH.Count)
		auth.POST("/wrong-questions/practice", wrongH.PracticeSubmit)
		auth.DELETE("/wrong-questions/:id", wrongH.Remove)

		auth.POST("/exam-templates/:id/start", examH.StartExam)
		auth.POST("/exam-records/:id/submit-answer", examH.SubmitAnswer)
		auth.POST("/exam-records/:id/submit", examH.SubmitExam)
		auth.GET("/exam-records", examH.ListRecords)
		auth.GET("/exam-records/:id/result", examH.GetResult)

		auth.GET("/materials/:id/download", studyMaterialH.Download)

		auth.GET("/stats/overview", statsH.Overview)
		auth.GET("/stats/daily", statsH.Daily)
		auth.GET("/stats/calendar", statsH.Calendar)
		auth.GET("/stats/subject-progress", statsH.SubjectProgress)
		auth.GET("/stats/chapter-progress", statsH.ChapterProgress)

		auth.GET("/check-in/today", checkInH.Today)
		auth.POST("/check-in/answer", checkInH.Answer)
		auth.POST("/check-in/complete", checkInH.Complete)
		auth.POST("/check-in/reset", checkInH.Reset)
		auth.GET("/check-in/history", checkInH.History)
		auth.GET("/check-in/stats", checkInH.Stats)

		auth.GET("/review/today", reviewH.Today)
		auth.POST("/review/answer", reviewH.Answer)
		auth.GET("/review/overview", reviewH.Overview)

		auth.GET("/study-plans", studyPlanH.List)
		auth.POST("/study-plans", studyPlanH.Create)
		auth.PUT("/study-plans/:id", studyPlanH.Update)
		auth.DELETE("/study-plans/:id", studyPlanH.Delete)

		auth.GET("/rankings", rankingH.Get)
		auth.GET("/rankings/estimated-section-scores", rankingH.GetEstimatedScores)

		aiRateLimiter := middleware.RateLimit(10, time.Minute)
		auth.POST("/ai/generate", aiRateLimiter, aiH.GenerateQuestions)
		// 异步 AI 出题：立即返回 task_id，后台生成完成后通过通知告知
		auth.POST("/ai/generate/async", aiRateLimiter, aiH.SubmitGenerateAsync)
		auth.GET("/ai/tasks", aiH.ListGenerateTasks)
		auth.GET("/ai/tasks/inflight", aiH.ListInflightTasks)
		auth.GET("/ai/tasks/:id", aiH.GetGenerateTask)
		auth.GET("/ai/history", aiH.ListHistory)
		auth.GET("/ai/history/:batch_id", aiH.GetBatchQuestions)
		auth.POST("/ai/analyze", aiRateLimiter, aiH.Analyze)
		auth.POST("/ai/exam/start", aiRateLimiter, aiH.StartExam)
		auth.POST("/ai/essay-score", aiRateLimiter, aiH.EssayScore)
		auth.GET("/ai/essay-score/check", aiH.CheckEssayScore)
		auth.POST("/ai/knowledge-points/extract", aiRateLimiter, knowledgePointH.Extract)

		auth.GET("/knowledge-points", knowledgePointH.List)
		auth.POST("/knowledge-points", knowledgePointH.Add)
		auth.POST("/knowledge-points/batch", knowledgePointH.BatchAdd)
		auth.GET("/knowledge-points/stats", knowledgePointH.Stats)
		auth.PATCH("/knowledge-points/:id", knowledgePointH.Update)
		auth.DELETE("/knowledge-points/:id", knowledgePointH.Remove)
		auth.POST("/knowledge-points/:id/resolve-duplicate", knowledgePointH.ResolveDuplicate)

		// 通知中心
		auth.GET("/notifications", notifyH.List)
		auth.GET("/notifications/unread-count", notifyH.UnreadCount)
		auth.POST("/notifications/:id/read", notifyH.MarkRead)
		auth.POST("/notifications/read-all", notifyH.MarkAllRead)
	}

	// 管理员专用路由组：Auth + RequireAdmin 双重保护，任何 /admin/* 路由都不得绕过
	admin := api.Group("/admin", middleware.Auth(cfg.JWTSecret), middleware.RequireAdmin(userRepo))
	{
		admin.GET("/invitation-codes", adminH.ListInvitationCodes)
		admin.POST("/invitation-codes", adminH.CreateInvitationCode)

		// 用户管理
		admin.GET("/users", adminH.ListUsers)
		admin.POST("/users", adminH.CreateUser)
		admin.PUT("/users/:id", adminH.UpdateUser)
		admin.PATCH("/users/:id/status", adminH.UpdateUserStatus)
		admin.POST("/users/:id/reset-password", adminH.ResetUserPassword)
		admin.DELETE("/users/:id", adminH.DeleteUser)
		admin.GET("/users/:id", adminH.GetUserDetail)

		// 科目管理
		admin.GET("/subjects", adminH.ListSubjects)
		admin.POST("/subjects", adminH.CreateSubject)
		admin.PUT("/subjects/:id", adminH.UpdateSubject)
		admin.PATCH("/subjects/:id/status", adminH.SetSubjectStatus)
		admin.DELETE("/subjects/:id", adminH.DeleteSubject)

		// 题库管理（管理员）
		admin.GET("/questions", adminH.ListQuestions)
		admin.POST("/questions", adminH.CreateQuestion)
		admin.GET("/questions/:id", adminH.GetQuestion)
		admin.PUT("/questions/:id", adminH.UpdateQuestion)
		admin.DELETE("/questions/:id", adminH.DeleteQuestion)
		// 批量启停：先于 /questions/:id 匹配（避免把 status 路由当成 :id）
		admin.PATCH("/questions/status", adminH.BatchUpdateQuestionStatus)
		admin.GET("/stats/questions", adminH.QuestionStats)
	}
}
