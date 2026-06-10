package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"orderflow/api/internal/http/handler"
	"orderflow/api/internal/http/middleware"
	"orderflow/api/internal/metrics"
)

func SetupRouter(
	authController *handler.AuthController,
	orderController *handler.OrderController,
	menuController *handler.MenuController,
	statsController *handler.StatsController,
	healthController *handler.HealthController,
	versionController *handler.VersionController,
	jwtSecret string,
) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logging())
	r.Use(metrics.Middleware())
	r.Use(cors.Default())

	// endpoints operacionais
	r.GET("/healthz", healthController.Healthz)
	r.GET("/readyz", healthController.Readyz)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	//PUBLIC
	public := r.Group("/api/v1")
	public.POST("/auth/register", authController.Register)
	public.POST("/auth/login", authController.Login)
	public.GET("/menu", menuController.List)
	public.GET("/orders", orderController.List)
	public.GET("/orders/:id", orderController.Get)
	public.GET("/stats", statsController.Today)
	public.GET("/version", versionController.Version)

	//PROTECTED
	protected := public.Group("/orders", middleware.AuthMiddleware(jwtSecret))
	protected.POST("", orderController.Create)

	return r
}
