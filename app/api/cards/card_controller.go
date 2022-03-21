package cards

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/ignavan39/ucrm-go/app/core/triggers"
	"github.com/ignavan39/ucrm-go/app/repository"
	"github.com/ignavan39/ucrm-go/pkg/httpext"

	blogger "github.com/sirupsen/logrus"
)

type Controller struct {
	repo            repository.CardRepository
	cardWebhookRepo repository.CardWebhookRepository
}

func NewController(repo repository.CardRepository, cardWebhookRepo repository.CardWebhookRepository) *Controller {
	return &Controller{
		repo:            repo,
		cardWebhookRepo: cardWebhookRepo,
	}
}

func (c *Controller) CreateOne(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var payload CreateOnePayload

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		httpext.JSON(w, httpext.CommonError{
			Error: "[CreateOne] failed decode payload",
			Code:  http.StatusBadRequest,
		}, http.StatusBadRequest)
		return
	}

	card, err := c.repo.AddCard(ctx, payload.Name, payload.Order, payload.PipelineId)
	if err != nil {
		blogger.Errorf("[card/createOne] CTX: [%v], ERROR:[%s]", ctx, err.Error())
		httpext.JSON(w, httpext.CommonError{
			Error: fmt.Sprintf("[CreateOne]:%s", err.Error()),
			Code:  http.StatusInternalServerError,
		}, http.StatusInternalServerError)
		return
	}

	webhook, err := c.cardWebhookRepo.GetCardWebhookByPipelineId(payload.PipelineId)
	if err != nil {
		blogger.Errorf("[card/createOne] CTX: [%v], ERROR:[%s]", ctx, err.Error())
		httpext.JSON(w, httpext.CommonError{
			Error: "[CreateOne] failed to get webhook",
			Code:  http.StatusBadRequest,
		}, http.StatusBadRequest)
		return
	}

	go triggers.SendCardUpdatesToSubscriber(webhook.Url, card, nil)
	httpext.JSON(w, card, http.StatusOK)
}

func (c *Controller) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "cardId")
	card, err := c.repo.GetOneCardWithoutRelations(ctx, id)

	if err != nil {
		httpext.JSON(w, httpext.CommonError{
			Error: fmt.Sprintf("[Delete]:%s", err.Error()),
			Code:  http.StatusInternalServerError,
		}, http.StatusInternalServerError)
		return
	}

	if card == nil {
		httpext.JSON(w, httpext.CommonError{
			Error: "card not found",
			Code:  http.StatusNotFound,
		}, http.StatusNotFound)
		return
	}

	err = c.repo.DeleteOneCard(ctx, id)
	if err != nil {
		httpext.JSON(w, httpext.CommonError{
			Error: fmt.Sprintf("[Delete]:%s", err.Error()),
			Code:  http.StatusInternalServerError,
		}, http.StatusInternalServerError)
		return
	}

	webhook, err := c.cardWebhookRepo.GetCardWebhookByPipelineId(card.PipelineId)
	if err != nil {
		httpext.JSON(w, httpext.CommonError{
			Error: "[Delete] failed to get webhook",
			Code:  http.StatusBadRequest,
		}, http.StatusBadRequest)
		return
	}

	go triggers.SendCardUpdatesToSubscriber(webhook.Url, nil, card)
	httpext.JSON(w, card, http.StatusOK)
}

func (c *Controller) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var payload UpdateOnePayload
	id := chi.URLParam(r, "cardId")

	if len(id) == 0 {
		httpext.JSON(w, httpext.CommonError{
			Error: "[Update] wrong id",
			Code:  http.StatusBadRequest,
		}, http.StatusBadRequest)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		httpext.JSON(w, httpext.CommonError{
			Error: "[Update] failed decode payload",
			Code:  http.StatusBadRequest,
		}, http.StatusBadRequest)
		return
	}

	card, err := c.repo.GetOneCard(ctx, id)
	if err != nil {
		httpext.JSON(w, httpext.CommonError{
			Error: fmt.Sprintf("[GetOne]:%s", err.Error()),
			Code:  http.StatusInternalServerError,
		}, http.StatusInternalServerError)
		return
	}

	if card == nil {
		httpext.JSON(w, httpext.CommonError{
			Error: "card not found",
			Code:  http.StatusNotFound,
		}, http.StatusNotFound)
		return
	}

	updatedCard, err := c.repo.UpdateCard(ctx, id, payload.Name, payload.Fields)
	if err != nil {
		httpext.JSON(w, httpext.CommonError{
			Error: fmt.Sprintf("[Update]:%s", err.Error()),
			Code:  http.StatusInternalServerError,
		}, http.StatusInternalServerError)
		return
	}

	webhook, err := c.cardWebhookRepo.GetCardWebhookByPipelineId(card.PipelineId)
	if err != nil {
		httpext.JSON(w, httpext.CommonError{
			Error: "[Update] failed to get pipeline",
			Code:  http.StatusBadRequest,
		}, http.StatusBadRequest)
		return
	}

	if webhook != nil {
		go triggers.SendCardUpdatesToSubscriber(webhook.Url, card, updatedCard)
	}

	httpext.JSON(w, updatedCard, http.StatusOK)
}

func (c *Controller) GetOne(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "cardId")

	if len(id) == 0 {
		httpext.JSON(w, httpext.CommonError{
			Error: "[Update] wrong id",
			Code:  http.StatusBadRequest,
		}, http.StatusBadRequest)
		return
	}

	card, err := c.repo.GetOneCard(ctx, id)
	if err != nil {
		httpext.JSON(w, httpext.CommonError{
			Error: fmt.Sprintf("[GetOne]:%s", err.Error()),
			Code:  http.StatusInternalServerError,
		}, http.StatusInternalServerError)
		return
	}

	if card == nil {
		httpext.JSON(w, httpext.CommonError{
			Error: "card not found",
			Code:  http.StatusNotFound,
		}, http.StatusNotFound)
		return
	}

	httpext.JSON(w, card, http.StatusOK)
}

func (c *Controller) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cardId := chi.URLParam(r, "cardId")
	pipelineId := chi.URLParam(r, "pipelineId")
	orderQuery := chi.URLParam(r, "order")

	if len(cardId) == 0 {
		httpext.JSON(w, httpext.CommonError{
			Error: "missing cardId: cards/updateOrder",
			Code:  http.StatusBadRequest,
		}, http.StatusBadRequest)
		return
	}

	if len(pipelineId) == 0 {
		httpext.JSON(w, httpext.CommonError{
			Error: "missing pipelineId: cards/updateOrder",
			Code:  http.StatusBadRequest,
		}, http.StatusBadRequest)
		return
	}

	if len(orderQuery) == 0 {
		httpext.JSON(w, httpext.CommonError{
			Error: "missing order: cards/updateOrder",
			Code:  http.StatusBadRequest,
		}, http.StatusBadRequest)
		return
	}

	newOrder, err := strconv.Atoi(orderQuery)
	if err != nil {
		httpext.JSON(w, httpext.CommonError{
			Error: "missing order: cards/updateOrder",
			Code:  http.StatusBadRequest,
		}, http.StatusBadRequest)
		return
	}

	var payload UpdateOrder
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		httpext.JSON(w, httpext.CommonError{
			Error: "failed decode payload: pipelines/createOne",
			Code:  http.StatusBadRequest,
		}, http.StatusBadRequest)
		return
	}

	err = c.repo.UpdateOrderForCard(ctx, cardId, pipelineId, payload.OldOrder, newOrder)
	if err != nil {
		httpext.JSON(w, httpext.CommonError{
			Error: err.Error(),
			Code:  http.StatusInternalServerError,
		}, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
