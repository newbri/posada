package api

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type renewAccessTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type renewAccessTokenResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}

func (server *Server) renewAccessToken(ctx *gin.Context) {
	var request renewAccessTokenRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.Error(ErrBindToJSON)
		//ctx.JSON(http.StatusBadRequest, errorResponse(ErrBindToJSON))
		//return
	}

	refreshPayload, err := server.tokenMaker.VerifyToken(request.RefreshToken)
	if err != nil {
		ctx.Error(ErrVerifyToken)
		//ctx.JSON(http.StatusUnauthorized, errorResponse(ErrVerifyToken))
		//return
	}

	session, err := server.store.GetSession(ctx, refreshPayload.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.Error(ErrNoRow)
			//ctx.JSON(http.StatusNotFound, errorResponse(ErrNoRow))
			//return
		}
		ctx.Error(ErrSession)
		//ctx.JSON(http.StatusInternalServerError, errorResponse(ErrSession))
		//return
	}

	if session.IsBlocked {
		ctx.Error(ErrBlockedSession)
		//ctx.JSON(http.StatusUnauthorized, errorResponse(ErrBlockedSession))
		//return
	}

	if session.Username != refreshPayload.Username {
		ctx.Error(ErrWrongUserSession)
		//ctx.JSON(http.StatusUnauthorized, errorResponse(ErrWrongUserSession))
		//return
	}

	if session.RefreshToken != request.RefreshToken {
		ctx.Error(ErrWrongSessionToken)
		//ctx.JSON(http.StatusUnauthorized, errorResponse(ErrWrongSessionToken))
		//return
	}

	if session.ExpiredAt.Before(time.Now()) {
		ctx.Error(ErrExpiredSession)
		//ctx.JSON(http.StatusUnauthorized, errorResponse(ErrExpiredSession))
		//return
	}

	accessToken, accessPayload, err := server.tokenMaker.CreateToken(
		refreshPayload.Username,
		refreshPayload.Role,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		ctx.Error(ErrTokenCreation)
		//ctx.JSON(http.StatusInternalServerError, errorResponse(ErrTokenCreation))
		//return
	}

	response := renewAccessTokenResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessPayload.ExpiredAt,
	}
	ctx.JSON(http.StatusOK, response)
}
