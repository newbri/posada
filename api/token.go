package api

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

func (server *Server) renewAccessToken(ctx *gin.Context) {
	var request struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		log.Info().Msg(ctx.Error(err).Error())
		return
	}

	refreshPayload, err := server.tokenMaker.VerifyToken(request.RefreshToken)
	if err != nil {
		log.Info().Msg(ctx.Error(ErrVerifyToken).Error())
		return
	}

	session, err := server.store.GetSession(ctx, refreshPayload.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info().Msg(ctx.Error(ErrNoRow).Error())
			return
		}
		log.Info().Msg(ctx.Error(ErrSession).Error())
		return
	}

	if session.IsBlocked {
		log.Info().Msg(ctx.Error(ErrBlockedSession).Error())
		return
	}

	if session.Username != refreshPayload.Username {
		log.Info().Msg(ctx.Error(ErrWrongUserSession).Error())
		return
	}

	if session.RefreshToken != request.RefreshToken {
		log.Info().Msg(ctx.Error(ErrWrongSessionToken).Error())
		return
	}

	if session.ExpiredAt.Before(time.Now()) {
		log.Info().Msg(ctx.Error(ErrExpiredSession).Error())
		return
	}

	accessToken, accessPayload, err := server.tokenMaker.CreateToken(
		refreshPayload.Username,
		refreshPayload.Role,
		server.config.GetConfig().AccessTokenDuration,
	)
	if err != nil {
		log.Info().Msg(ctx.Error(ErrTokenCreation).Error())
		return
	}

	var response struct {
		AccessToken          string    `json:"access_token"`
		AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
	}

	response.AccessToken = accessToken
	response.AccessTokenExpiresAt = accessPayload.ExpiredAt

	ctx.JSON(http.StatusOK, response)
}
