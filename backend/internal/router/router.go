package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/soft-test-study/backend/internal/config"
	"github.com/soft-test-study/backend/internal/handler"
	"github.com/soft-test-study/backend/internal/middleware"
	"github.com/soft-test-study/backend/internal/repository"
	"github.com/soft-test-study/backend/internal/service"
	"gorm.io/gorm"
)

func Setup(db *gorm.DB, r *gin.Engine, cfg *config.Config) {
	api := r.Group("/api/v1")

	// 启动 AI 异步任务清理后台 goroutine
	service.StartTaskJanitor()

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
	essayScoreRepo := repository.NewEssayScoreRepo(db)
	studyMaterialRepo := repository.NewStudyMaterialRepo(db)
	checkInRepo := repository.NewCheckInRepo(db)
	reviewCardRepo := repository.NewReviewCardRepo(db)
	studyPlanRepo := repository.NewStudyPlanRepo(db)
	rankingRepo := repository.NewRankingRepo(db)

	notifySvc := service.NewNotificationService(db)

	userSvc := service.NewUserService(userRepo, examLevelRepo, subjectRepo, cfg.JWTSecret, cfg.JWTExpiresIn)
	examLevelSvc := service.NewExamLevelService(examLevelRepo)
	subjectSvc := service.NewSubjectService(subjectRepo)
	subSubjectSvc := service.NewSubSubjectService(subSubjectRepo)
	chapterSvc := service.NewChapterService(chapterRepo)
	questionSvc := service.NewQuestionService(questionRepo)
	reviewSvc := service.NewReviewService(reviewCardRepo, questionRepo, practiceRecordRepo, wrongRepo)
	practiceSvc := service.NewPracticeService(practiceRecordRepo, questionRepo, reviewSvc)
	bookmarkSvc := service.NewBookmarkService(bookmarkRepo)
	markSvc := service.NewQuestionMarkService(markRepo)
	wrongSvc := service.NewWrongQuestionService(wrongRepo, questionRepo, practiceRecordRepo, reviewSvc)
	examSvc := service.NewExamService(examTemplateRepo, examRecordRepo, questionRepo, reviewSvc)
	statsRepo := repository.NewStatsRepo(db)
	statsSvc := service.NewStatsService(statsRepo, subjectRepo)
	aiSvc := service.NewAiService(aiRepo, questionRepo, subjectRepo, chapterRepo, examRecordRepo, essayScoreRepo, practiceRecordRepo, notifySvc)
	studyMaterialSvc := service.NewStudyMaterialService(studyMaterialRepo, subjectRepo)
	checkInSvc := service.NewCheckInService(checkInRepo, questionRepo, practiceRecordRepo, reviewSvc)
	studyPlanSvc := service.NewStudyPlanService(studyPlanRepo, practiceRecordRepo)
	rankingSvc := service.NewRankingService(rankingRepo, userRepo)

	userH := handler.NewUserHandler(userSvc)
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

	rateLimiter := middleware.RateLimit(5, time.Minute)

	api.POST("/auth/register", rateLimiter, userH.Register)
	api.POST("/auth/login", rateLimiter, userH.Login)

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

	auth := api.Group("")
	auth.Use(middleware.Auth(cfg.JWTSecret))
	{
		auth.GET("/auth/user-info", userH.GetUserInfo)
		auth.PUT("/auth/profile", userH.UpdateProfile)
		auth.PUT("/auth/password", userH.ChangePassword)

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

		aiRateLimiter := middleware.RateLimit(10, time.Minute)
		auth.POST("/ai/generate", aiRateLimiter, aiH.GenerateQuestions)
		// 异步 AI 出题：立即返回 task_id，后台生成完成后通过通知告知
		auth.POST("/ai/generate/async", aiRateLimiter, aiH.SubmitGenerateAsync)
		auth.GET("/ai/tasks", aiH.ListGenerateTasks)
		auth.GET("/ai/tasks/:id", aiH.GetGenerateTask)
		auth.GET("/ai/history", aiH.ListHistory)
		auth.GET("/ai/history/:batch_id", aiH.GetBatchQuestions)
		auth.POST("/ai/analyze", aiRateLimiter, aiH.Analyze)
		auth.POST("/ai/exam/start", aiRateLimiter, aiH.StartExam)
		auth.POST("/ai/essay-score", aiRateLimiter, aiH.EssayScore)
		auth.GET("/ai/essay-score/check", aiH.CheckEssayScore)

		// 通知中心
		auth.GET("/notifications", notifyH.List)
		auth.GET("/notifications/unread-count", notifyH.UnreadCount)
		auth.POST("/notifications/:id/read", notifyH.MarkRead)
		auth.POST("/notifications/read-all", notifyH.MarkAllRead)
	}
}
