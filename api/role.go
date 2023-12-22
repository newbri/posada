package api

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/newbri/posadamissportia/db"
	"net/http"
	"time"
)

type createRoleRequest struct {
	Name        string `json:"name" binding:"required,alphanum"`
	Description string `json:"description"`
}

func (server *Server) createRole(ctx *gin.Context) {
	var request createRoleRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(ErrBindToJSON))
		return
	}

	id, _ := uuid.NewV7()
	arg := db.CreateRoleParams{
		ID:          id,
		Name:        request.Name,
		Description: request.Description,
	}

	role, err := server.store.CreateRole(ctx, arg)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code.Name() {
			case "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(ErrUniqueViolation))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(ErrInternalServer))
		return
	}

	roleResp := roleResponse{
		ExternalID:  role.ExternalID,
		Name:        role.Name,
		Description: role.Description,
		UpdatedAt:   role.UpdatedAt,
		CreatedAt:   role.CreatedAt,
	}

	ctx.JSON(http.StatusOK, roleResp)
}

func (server *Server) getAllRole(ctx *gin.Context) {
	var request struct {
		Limit  int32 `json:"limit" binding:"required,gte=1"`
		Offset int32 `json:"offset" binding:"min=0"`
	}
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(ErrBindToJSON))
		return
	}

	arg := db.ListRoleParams{
		Limit:  request.Limit,
		Offset: request.Offset,
	}

	role, err := server.store.GetAllRole(ctx, arg)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, errorResponse(ErrNoRow))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(ErrInternalServer))
		return
	}

	var response []roleResponse
	for _, r := range role {
		response = append(response, roleResponse{
			ExternalID:  r.ExternalID,
			Name:        r.Name,
			Description: r.Description,
			UpdatedAt:   r.UpdatedAt,
			CreatedAt:   r.CreatedAt,
		})
	}

	ctx.JSON(http.StatusOK, response)
}

type idURI struct {
	ID string `uri:"id" binding:"required,alphanum"`
}

type roleResponse struct {
	ExternalID  string    `json:"external_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedAt   time.Time `json:"created_at"`
}

func (server *Server) getRole(ctx *gin.Context) {
	var request idURI
	if err := ctx.ShouldBindUri(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(ErrShouldBindUri))
		return
	}

	role, err := server.store.GetRole(ctx, request.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, errorResponse(ErrNoRow))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(ErrInternalServer))
		return
	}

	if role == nil {
		ctx.JSON(http.StatusNotFound, errorResponse(errors.New("role not found")))
		return
	}

	roleResp := roleResponse{
		ExternalID:  role.ExternalID,
		Name:        role.Name,
		Description: role.Description,
		UpdatedAt:   role.UpdatedAt,
		CreatedAt:   role.CreatedAt,
	}

	ctx.JSON(http.StatusOK, roleResp)
}
