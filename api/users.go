package api

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/newbri/posadamissportia/db"
	"github.com/newbri/posadamissportia/db/util"
	"net/http"
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
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	arg := db.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hashedPassword,
		FullName:       req.FullName,
		Email:          req.Email,
	}

	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code.Name() {
			case "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	response := newUserResponse(user)
	ctx.JSON(http.StatusOK, response)
}

type getUserRequest struct {
	Username string `uri:"username" binding:"required,alphanum"`
}

func (server *Server) getUser(ctx *gin.Context) {
	var request getUserRequest
	if err := ctx.ShouldBindUri(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUser(ctx, request.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
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
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	args := db.UpdateUserParams{
		Username: request.Username,
		FullName: sql.NullString{
			String: request.FullName,
			Valid:  request.FullName != "",
		},
		Email: sql.NullString{
			String: request.Email,
			Valid:  request.Email != "",
		},
	}

	if request.Password != "" {
		hashedPassword, err := util.HashPassword(request.Password)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
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
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
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
}

func newUserResponse(user db.User) userResponse {
	return userResponse{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
	}
}
