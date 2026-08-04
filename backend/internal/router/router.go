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

	userRepo := repository.NewUserRepo(db)
	examLevelRepo := repository.NewExamLevelRepo(db)
	subjectRepo := repository.NewSubjectRepo(db)
	subSubjectRepo := repository.NewSubSubjectRepo(db)
	chapterRepo := repository.NewChapterRepo(db)
	questionRepo := repository.NewQuestionRepo(db)
	practiceRecordRepo := repository.NewPracticeRecordRepo(db)
	bookmarkRepo := repository.NewBookmarkRepo(db)
	wrongRepo := repository.NewWrongQuestionRepo(db)
	examTemplateRepo := repository.NewExamTemplateRepo(db)
	examRecordRepo := repository.NewExamRecordRepo(db)
	aiRepo := repository.NewAiRepo(db)
	essayScoreRepo := repository.NewEssayScoreRepo(db)

	userSvc := service.NewUserService(userRepo, examLevelRepo, subjectRepo, cfg.JWTSecret, cfg.JWTExpiresIn)
	examLevelSvc := service.NewExamLevelService(examLevelRepo)
	subjectSvc := service.NewSubjectService(subjectRepo)
	subSubjectSvc := service.NewSubSubjectService(subSubjectRepo)
	chapterSvc := service.NewChapterService(chapterRepo)
	questionSvc := service.NewQuestionService(questionRepo)
	practiceSvc := service.NewPracticeService(practiceRecordRepo, questionRepo, wrongRepo)
	bookmarkSvc := service.NewBookmarkService(bookmarkRepo)
	wrongSvc := service.NewWrongQuestionService(wrongRepo, questionRepo, practiceRecordRepo)
	examSvc := service.NewExamService(examTemplateRepo, examRecordRepo, questionRepo)
	statsRepo := repository.NewStatsRepo(db)
	statsSvc := service.NewStatsService(statsRepo, subjectRepo)
	aiSvc := service.NewAiService(aiRepo, questionRepo, subjectRepo, chapterRepo, examRecordRepo, essayScoreRepo, practiceRecordRepo)

	userH := handler.NewUserHandler(userSvc)
	examLevelH := handler.NewExamLevelHandler(examLevelSvc)
	subjectH := handler.NewSubjectHandler(subjectSvc)
	subSubjectH := handler.NewSubSubjectHandler(subSubjectSvc)
	chapterH := handler.NewChapterHandler(chapterSvc)
	questionH := handler.NewQuestionHandler(questionSvc)
	practiceH := handler.NewPracticeHandler(practiceSvc)
	bookmarkH := handler.NewBookmarkHandler(bookmarkSvc)
	wrongH := handler.NewWrongQuestionHandler(wrongSvc)
	examH := handler.NewExamHandler(examSvc)
	statsH := handler.NewStatsHandler(statsSvc)
	aiH := handler.NewAiHandler(aiSvc)

	rateLimiter := middleware.RateLimit(5, time.Minute)

	api.POST("/auth/register", rateLimiter, userH.Register)
	api.POST("/auth/login", rateLimiter, userH.Login)

	api.GET("/exam-levels", examLevelH.List)
	api.GET("/subjects", subjectH.ListByLevel)
	api.GET("/subjects/all", subjectH.ListAll)
	api.GET("/subjects/:id/sub-subjects", subSubjectH.ListBySubject)
	api.GET("/sub-subjects/:id/chapters", chapterH.ListBySubSubject)
	api.GET("/questions", questionH.ListByChapter)
	api.GET("/questions/:id", questionH.GetByID)
	api.GET("/exam-templates", examH.ListTemplates)
	api.GET("/ai/providers", aiH.GetProviders)

	auth := api.Group("")
	auth.Use(middleware.Auth(cfg.JWTSecret))
	{
		auth.GET("/auth/user-info", userH.GetUserInfo)
		auth.PUT("/auth/profile", userH.UpdateProfile)
		auth.PUT("/auth/password", userH.ChangePassword)

		auth.POST("/practice/submit", practiceH.Submit)

		auth.GET("/questions/random", questionH.Random)
		auth.GET("/questions/special", questionH.Special)

		auth.POST("/favorite-folders", bookmarkH.CreateFolder)
		auth.GET("/favorite-folders", bookmarkH.ListFolders)
		auth.POST("/questions/:id/favorite", bookmarkH.AddFavorite)
		auth.DELETE("/questions/:id/favorite", bookmarkH.RemoveFavorite)
		auth.GET("/questions/:id/favorited", bookmarkH.CheckFavorited)
		auth.GET("/favorites", bookmarkH.ListFavorites)

		auth.GET("/wrong-questions", wrongH.List)
		auth.GET("/wrong-questions/count", wrongH.Count)
		auth.POST("/wrong-questions/practice", wrongH.PracticeSubmit)
		auth.DELETE("/wrong-questions/:id", wrongH.Remove)

		auth.POST("/exam-templates/:id/start", examH.StartExam)
		auth.POST("/exam-records/:id/submit-answer", examH.SubmitAnswer)
		auth.POST("/exam-records/:id/submit", examH.SubmitExam)
		auth.GET("/exam-records", examH.ListRecords)
		auth.GET("/exam-records/:id/result", examH.GetResult)

		auth.GET("/stats/overview", statsH.Overview)
		auth.GET("/stats/daily", statsH.Daily)
		auth.GET("/stats/calendar", statsH.Calendar)
		auth.GET("/stats/subject-progress", statsH.SubjectProgress)
		auth.GET("/stats/chapter-progress", statsH.ChapterProgress)

		aiRateLimiter := middleware.RateLimit(10, time.Minute)
		auth.POST("/ai/generate", aiRateLimiter, aiH.GenerateQuestions)
		auth.POST("/ai/analyze", aiRateLimiter, aiH.Analyze)
		auth.POST("/ai/exam/start", aiRateLimiter, aiH.StartExam)
		auth.POST("/ai/essay-score", aiRateLimiter, aiH.EssayScore)
		auth.GET("/ai/essay-score/check", aiH.CheckEssayScore)
	}
}
