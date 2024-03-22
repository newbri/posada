package api

import (
	"github.com/gin-gonic/gin"
	"github.com/newbri/posadamissportia/db"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

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

	var response struct {
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

	response.ExternalID = createdProperty.ExternalID
	response.Name = createdProperty.Name
	response.Address = createdProperty.Address
	response.State = createdProperty.State
	response.City = createdProperty.City
	response.Country = createdProperty.Country
	response.PostalCode = createdProperty.PostalCode
	response.Phone = createdProperty.Phone
	response.Email = createdProperty.Email
	response.IsActive = createdProperty.IsActive
	response.CreatedAt = createdProperty.CreatedAt

	ctx.JSON(http.StatusOK, response)
}
