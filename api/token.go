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
		ctx.JSON(http.StatusBadRequest, errorResponse(ErrBindToJSON))
		return
	}

	refreshPayload, err := server.tokenMaker.VerifyToken(request.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(ErrVerifyToken))
		return
	}

	session, err := server.store.GetSession(ctx, refreshPayload.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, errorResponse(ErrNoRow))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(ErrSession))
		return
	}

	if session.IsBlocked {
		ctx.JSON(http.StatusUnauthorized, errorResponse(ErrBlockedSession))
		return
	}

	if session.Username != refreshPayload.Username {
		ctx.JSON(http.StatusUnauthorized, errorResponse(ErrWrongUserSession))
		return
	}

	if session.RefreshToken != request.RefreshToken {
		ctx.JSON(http.StatusUnauthorized, errorResponse(ErrWrongSessionToken))
		return
	}

	if session.ExpiredAt.Before(time.Now()) {
		ctx.JSON(http.StatusUnauthorized, errorResponse(ErrExpiredSession))
		return
	}

	accessToken, accessPayload, err := server.tokenMaker.CreateToken(
		refreshPayload.Username,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(ErrTokenCreation))
		return
	}

	response := renewAccessTokenResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessPayload.ExpiredAt,
	}
	ctx.JSON(http.StatusOK, response)
}
