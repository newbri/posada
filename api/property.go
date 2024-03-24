package api

import (
	"github.com/gin-gonic/gin"
	"github.com/newbri/posadamissportia/db"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

type propertyResponse struct {
	ExternalID string    `json:"external_id"`
	Name       string    `json:"name"`
	Address    string    `json:"address"`
	State      string    `json:"state"`
	City       string    `json:"city"`
	Country    string    `json:"country"`
	PostalCode string    `json:"postal_code"`
	Phone      string    `json:"phone"`
	Email      string    `json:"email"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
}

func (server *Server) createProperty(ctx *gin.Context) {
	var request struct {
		Name       string `json:"name" binding:"required"`
		Address    string `json:"address" binding:"required"`
		State      string `json:"state" binding:"required"`
		City       string `json:"city" binding:"required"`
		Country    string `json:"country" binding:"required"`
		PostalCode string `json:"postal_code" binding:"required"`
		Phone      string `json:"phone" binding:"required,e164"`
		Email      string `json:"email" binding:"required,email"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		log.Info().Msg(ctx.Error(err).Error())
		return
	}

	arg := db.CreatePropertyParams{
		Name:       request.Name,
		Address:    request.Address,
		State:      request.State,
		City:       request.City,
		Country:    request.Country,
		PostalCode: request.PostalCode,
		Phone:      request.Phone,
		Email:      request.Email,
		ExpiredAt:  time.Time{},
		CreatedAt:  time.Now(),
	}

	createdProperty, err := server.store.CreateProperty(ctx, arg)
	if err != nil {
		log.Info().Msg(ctx.Error(err).Error())
		return
	}

	response := propertyResponse{
		ExternalID: createdProperty.ExternalID,
		Name:       createdProperty.Name,
		Address:    createdProperty.Address,
		State:      createdProperty.State,
		City:       createdProperty.City,
		Country:    createdProperty.Country,
		PostalCode: createdProperty.PostalCode,
		Phone:      createdProperty.Phone,
		Email:      createdProperty.Email,
		IsActive:   createdProperty.IsActive,
		CreatedAt:  createdProperty.CreatedAt,
	}

	ctx.JSON(http.StatusOK, response)
}

func (server *Server) activateDeactivateProperty(ctx *gin.Context) {
	var request struct {
		Active     *bool  `json:"active" binding:"required"`
		ExternalID string `json:"external_id" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		log.Info().Msg(ctx.Error(err).Error())
		return
	}

	activeProperty, err := server.store.ActivateDeactivateProperty(ctx, *request.Active, request.ExternalID)
	if err != nil {
		log.Info().Msg(ctx.Error(err).Error())
		return
	}

	response := propertyResponse{
		ExternalID: activeProperty.ExternalID,
		Name:       activeProperty.Name,
		Address:    activeProperty.Address,
		State:      activeProperty.State,
		City:       activeProperty.City,
		Country:    activeProperty.Country,
		PostalCode: activeProperty.PostalCode,
		Phone:      activeProperty.Phone,
		Email:      activeProperty.Email,
		IsActive:   activeProperty.IsActive,
		CreatedAt:  activeProperty.CreatedAt,
	}

	ctx.JSON(http.StatusOK, response)
}
