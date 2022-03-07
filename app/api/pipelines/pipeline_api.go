package pipelines

import (
	"github.com/go-chi/chi"
	"github.com/ignavan39/ucrm-go/app/config"
	"github.com/ignavan39/ucrm-go/app/middlewares"
	"github.com/ignavan39/ucrm-go/app/repository"
)

func RegisterRouter(r chi.Router, controller *Controller, repo repository.DashboardRepository, pipelineRepo repository.PipelineRepository, config config.JWTConfig) {
	r.Group(func(r chi.Router) {
		r.Use(middlewares.AuthGuard(config))
		r.Route("/pipelines", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(middlewares.DashboardAccessGuard(repo, "rw"))
				r.Patch("/order/{dashboardId}/{id}/{order}", controller.UpdateOrder)
				r.Post("/create", controller.CreateOne)
			})
			r.Group(func(r chi.Router) {
				r.Use(middlewares.PipelineAccessGuard(pipelineRepo, "rw"))
				r.Patch("/{id}", controller.UpdateName)
				r.Delete("/{id}", controller.DeleteById)
			})
		})
	})
}
