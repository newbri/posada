package api

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/newbri/posadamissportia/db"
	"github.com/rs/zerolog/log"
	"net/http"
	"strings"
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

func (server *Server) getAllProperty(ctx *gin.Context) {
	var request struct {
		Limit  int32 `json:"limit" binding:"required,gte=1"`
		Offset int32 `json:"offset" binding:"min=0"`
	}
	if err := ctx.ShouldBindJSON(&request); err != nil {
		log.Info().Msg(ctx.Error(err).Error())
		return
	}

	arg := db.LimitOffset{
		Limit:  request.Limit,
		Offset: request.Offset,
	}

	allProperty, err := server.store.GetAllProperty(ctx, arg)
	if err != nil {
		log.Info().Msg(ctx.Error(err).Error())
		return
	}

	ctx.JSON(http.StatusOK, allProperty)
}

type propertyID struct {
	ID string `uri:"id" binding:"required,alphanum"`
}

func (server *Server) getProperty(ctx *gin.Context) {
	var request propertyID
	if err := ctx.ShouldBindUri(&request); err != nil {
		log.Info().Msg(ctx.Error(ErrShouldBindUri).Error())
		return
	}

	property, err := server.store.GetProperty(ctx, request.ID)
	if err != nil {
		log.Info().Msg(ctx.Error(err).Error())
		return
	}

	ctx.JSON(http.StatusOK, property)
}

func (server *Server) updateProperty(ctx *gin.Context) {
	var request struct {
		ExternalID string `json:"external_id" binding:"required"`
		Name       string `json:"name" binding:"omitempty"`
		Address    string `json:"address" binding:"omitempty"`
		State      string `json:"state" binding:"omitempty"`
		City       string `json:"city" binding:"omitempty"`
		Country    string `json:"country" binding:"omitempty"`
		PostalCode string `json:"postal_code" binding:"omitempty"`
		Phone      string `json:"phone" binding:"omitempty,e164"`
		Email      string `json:"email" binding:"omitempty,email"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		log.Info().Msg(ctx.Error(err).Error())
		return
	}

	args := db.UpdatePropertyParams{
		ExternalID: request.ExternalID,
		Name: sql.NullString{
			String: request.Name,
			Valid:  len(strings.TrimSpace(request.Name)) > 0,
		},
		Address: sql.NullString{
			String: request.Address,
			Valid:  len(strings.TrimSpace(request.Address)) > 0,
		},
		State: sql.NullString{
			String: request.State,
			Valid:  len(strings.TrimSpace(request.State)) > 0,
		},
		City: sql.NullString{
			String: request.City,
			Valid:  len(strings.TrimSpace(request.City)) > 0,
		},
		Country: sql.NullString{
			String: request.Country,
			Valid:  len(strings.TrimSpace(request.Country)) > 0,
		},
		PostalCode: sql.NullString{
			String: request.PostalCode,
			Valid:  len(strings.TrimSpace(request.PostalCode)) > 0,
		},
		Phone: sql.NullString{
			String: request.Phone,
			Valid:  len(strings.TrimSpace(request.Phone)) > 0,
		},
		Email: sql.NullString{
			String: request.Email,
			Valid:  len(strings.TrimSpace(request.Email)) > 0,
		},
	}

	property, err := server.store.UpdateProperty(ctx, args)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info().Msg(ctx.Error(ErrNoRow).Error())
			return
		}
		log.Info().Msg(ctx.Error(ErrInternalServer).Error())
		return
	}

	ctx.JSON(http.StatusOK, property)
}

func (server *Server) deleteProperty(ctx *gin.Context) {
	var request propertyID
	if err := ctx.ShouldBindUri(&request); err != nil {
		log.Info().Msg(ctx.Error(ErrShouldBindUri).Error())
		return
	}

	property, err := server.store.DeleteProperty(ctx, request.ID)
	if err != nil {
		log.Info().Msg(ctx.Error(err).Error())
		return
	}

	ctx.JSON(http.StatusOK, property)
}
