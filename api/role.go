package api

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/newbri/posadamissportia/db"
	"net/http"
)

type createRoleRequest struct {
	Name        string `json:"name" binding:"required,alphanum"`
	Description string `json:"description"`
}

func (server *Server) createRole(ctx *gin.Context) {
	var request createRoleRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CreateRoleParams{
		ID:          uuid.New(),
		Name:        request.Name,
		Description: request.Description,
	}

	role, err := server.store.CreateRole(ctx, arg)
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

	ctx.JSON(http.StatusOK, role)
}

func (server *Server) getAllRole(ctx *gin.Context) {
	var request struct {
		Limit  int32 `json:"limit" binding:"required,gte=1"`
		Offset int32 `json:"offset" binding:"min=0"`
	}
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListRoleParams{
		Limit:  request.Limit,
		Offset: request.Offset,
	}

	role, err := server.store.GetAllRole(ctx, arg)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, role)
}
