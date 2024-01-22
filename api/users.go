package api

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/newbri/posadamissportia/db"
	"github.com/newbri/posadamissportia/db/util"
	"github.com/newbri/posadamissportia/token"
	"github.com/rs/zerolog/log"
	"net/http"
	"strings"
	"time"
)

type createUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

func (server *Server) createUser(ctx *gin.Context) {
	var req createUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Info().Msg(ctx.Error(err).Error())
		return
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		log.Info().Msg(ctx.Error(ErrInternalServer).Error())
		return
	}

	defaultRole := server.config.DefaultRole
	role, err := server.store.GetRoleByName(ctx, defaultRole)
	if err != nil {
		log.Info().Msg(ctx.Error(err).Error())
		return
	}

	arg := db.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hashedPassword,
		FullName:       req.FullName,
		Email:          req.Email,
		RoleID:         role.InternalID,
	}

	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code.Name() {
			case "unique_violation":
				log.Info().Msg(ctx.Error(ErrUniqueViolation).Error())
				return
			}
		}
		log.Info().Msg(ctx.Error(ErrInternalServer).Error())
		return
	}

	response := newUserResponse(user)
	ctx.JSON(http.StatusOK, response)
}

type usernameURI struct {
	Username string `uri:"username" binding:"required,alphanum"`
}

func (server *Server) getUser(ctx *gin.Context) {
	var request usernameURI
	if err := ctx.ShouldBindUri(&request); err != nil {
		log.Info().Msg(ctx.Error(ErrShouldBindUri).Error())
		return
	}

	user, err := server.store.GetUser(ctx, request.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info().Msg(ctx.Error(ErrNoRow).Error())
			return
		}
		log.Info().Msg(ctx.Error(ErrInternalServer).Error())
		return
	}

	response := newUserResponse(user)
	ctx.JSON(http.StatusOK, response)
}

type updateUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"omitempty,min=6"`
	FullName string `json:"full_name" binding:"omitempty"`
	Email    string `json:"email" binding:"omitempty,email"`
}

func (server *Server) updateUser(ctx *gin.Context) {
	var request updateUserRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		log.Info().Msg(ctx.Error(err).Error())
		return
	}

	args := db.UpdateUserParams{
		Username: request.Username,
		FullName: sql.NullString{
			String: request.FullName,
			Valid:  len(strings.TrimSpace(request.FullName)) > 0,
		},
		Email: sql.NullString{
			String: request.Email,
			Valid:  len(strings.TrimSpace(request.Email)) > 0,
		},
	}

	if len(strings.TrimSpace(request.Password)) > 0 {
		hashedPassword, err := util.HashPassword(request.Password)
		if err != nil {
			log.Info().Msg(ctx.Error(ErrInternalServer).Error())
			return
		}

		args.HashedPassword = sql.NullString{
			String: hashedPassword,
			Valid:  true,
		}

		args.PasswordChangedAt = sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		}
	}

	user, err := server.store.UpdateUser(ctx, args)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info().Msg(ctx.Error(ErrNoRow).Error())
			return
		}
		log.Info().Msg(ctx.Error(ErrInternalServer).Error())
		return
	}

	response := newUserResponse(user)
	ctx.JSON(http.StatusOK, response)
}

type userResponse struct {
	Username          string    `json:"username"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
	Role              *db.Role  `json:"role"`
}

func newUserResponse(user *db.User) userResponse {
	return userResponse{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
		Role:              user.Role,
	}
}

func (server *Server) deleteUser(ctx *gin.Context) {
	var request usernameURI
	if err := ctx.ShouldBindUri(&request); err != nil {
		log.Info().Msg(ctx.Error(ErrShouldBindUri).Error())
		return
	}

	user, err := server.store.DeleteUser(ctx, request.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info().Msg(ctx.Error(ErrNoRow).Error())
			return
		}
		log.Info().Msg(ctx.Error(ErrInternalServer).Error())
		return
	}

	response := newUserResponse(user)
	ctx.JSON(http.StatusOK, response)
}

type loginUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginUserResponse struct {
	SessionID             uuid.UUID    `json:"session_id"`
	AccessToken           string       `json:"access_token"`
	AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"`
	RefreshToken          string       `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"`
	User                  userResponse `json:"user"`
}

func (server *Server) loginUser(ctx *gin.Context) {
	var request loginUserRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		log.Info().Msg(ctx.Error(err).Error())
		return
	}

	user, err := server.store.GetUser(ctx, request.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info().Msg(ctx.Error(ErrNoRow).Error())
			return
		}

		log.Info().Msg(ctx.Error(ErrInternalServer).Error())
		return
	}

	err = util.CheckPassword(request.Password, user.HashedPassword)
	if err != nil {
		log.Info().Msg(ctx.Error(ErrPasswordMistMach).Error())
		return
	}

	accessToken, accessPayload, err := server.tokenMaker.CreateToken(
		user.Username,
		user.Role,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		log.Info().Msg(ctx.Error(ErrInternalServer).Error())
		return
	}

	refreshToken, refreshPayload, err := server.tokenMaker.CreateToken(
		user.Username,
		user.Role,
		server.config.RefreshTokenDuration,
	)
	if err != nil {
		log.Info().Msg(ctx.Error(ErrInternalServer).Error())
		return
	}

	session, err := server.store.CreateSession(ctx, db.CreateSessionParams{
		ID:           refreshPayload.ID,
		Username:     user.Username,
		RefreshToken: refreshToken,
		UserAgent:    ctx.Request.UserAgent(),
		ClientIp:     ctx.ClientIP(),
		IsBlocked:    false,
		ExpiredAt:    refreshPayload.ExpiredAt,
		CreatedAt:    time.Now(),
	})
	if err != nil {
		log.Info().Msg(ctx.Error(ErrInternalServer).Error())
		return
	}

	response := loginUserResponse{
		SessionID:             session.ID,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
		User:                  newUserResponse(user),
	}
	ctx.JSON(http.StatusOK, response)
}

func (server *Server) getUserInfo(ctx *gin.Context) {
	data, _ := ctx.Get(authorizationPayloadKey)

	payload, _ := data.(*token.Payload)
	user, err := server.store.GetUser(ctx, payload.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info().Msg(ctx.Error(ErrNoRow).Error())
			return
		}
		log.Info().Msg(ctx.Error(ErrInternalServer).Error())
		return
	}

	response := newUserResponse(user)
	ctx.JSON(http.StatusOK, response)
}

func (server *Server) getAllCustomerUser(ctx *gin.Context) {
	var request struct {
		Limit  int32 `json:"limit" binding:"required,gte=1"`
		Offset int32 `json:"offset" binding:"min=0"`
	}
	if err := ctx.ShouldBindJSON(&request); err != nil {
		log.Info().Msg(ctx.Error(err).Error())
		return
	}

	arg := db.ListUsersParams{
		Limit:  request.Limit,
		Offset: request.Offset,
	}

	users, err := server.store.GetAllCustomerUser(ctx, arg)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info().Msg(ctx.Error(ErrNoRow).Error())
			return
		}
		log.Info().Msg(ctx.Error(ErrInternalServer).Error())
		return
	}

	ctx.JSON(http.StatusOK, users)
}
