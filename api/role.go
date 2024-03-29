package api

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/newbri/posadamissportia/db"
	"github.com/rs/zerolog/log"
	"net/http"
	"strings"
	"time"
)

func (server *Server) createRole(ctx *gin.Context) {
	var request struct {
		Name        string `json:"name" binding:"required,alphanum"`
		Description string `json:"description" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		log.Info().Msg(ctx.Error(err).Error())
		return
	}

	arg := db.CreateRoleParams{
		Name:        request.Name,
		Description: request.Description,
	}

	role, err := server.store.CreateRole(ctx, arg)
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
		log.Info().Msg(ctx.Error(err).Error())
		return
	}

	arg := db.ListRoleParams{
		Limit:  request.Limit,
		Offset: request.Offset,
	}

	roles, err := server.store.GetAllRole(ctx, arg)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info().Msg(ctx.Error(ErrNoRow).Error())
			return
		}
		log.Info().Msg(ctx.Error(ErrInternalServer).Error())
		return
	}

	var response []roleResponse
	for _, r := range roles {
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
		log.Info().Msg(ctx.Error(ErrShouldBindUri).Error())
		return
	}

	role, err := server.store.GetRole(ctx, request.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info().Msg(ctx.Error(ErrNoRow).Error())
			return
		}
		log.Info().Msg(ctx.Error(ErrInternalServer).Error())
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

func (server *Server) updateRole(ctx *gin.Context) {
	type updateRoleRequest struct {
		ExternalID  string `json:"external_id" binding:"required,alphanum"`
		Name        string `json:"name" binding:"omitempty"`
		Description string `json:"description" binding:"omitempty"`
	}

	var request updateRoleRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		log.Info().Msg(ctx.Error(err).Error())
		return
	}

	args := db.UpdateRoleParams{
		ExternalID:  request.ExternalID,
		Name:        sql.NullString{String: request.Name, Valid: len(strings.TrimSpace(request.Name)) > 0},
		Description: sql.NullString{String: request.Description, Valid: len(strings.TrimSpace(request.Description)) > 0},
		UpdateAt:    time.Now(),
	}

	role, err := server.store.UpdateRole(ctx, args)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info().Msg(ctx.Error(ErrNoRow).Error())
			return
		}
		log.Info().Msg(ctx.Error(ErrInternalServer).Error())
		return
	}

	response := roleResponse{
		ExternalID:  role.ExternalID,
		Name:        role.Name,
		Description: role.Description,
		UpdatedAt:   role.UpdatedAt,
		CreatedAt:   role.CreatedAt,
	}
	ctx.JSON(http.StatusOK, response)
}

func (server *Server) deleteRole(ctx *gin.Context) {
	var request idURI
	if err := ctx.ShouldBindUri(&request); err != nil {
		log.Info().Msg(ctx.Error(ErrShouldBindUri).Error())
		return
	}

	role, err := server.store.DeleteRole(ctx, request.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info().Msg(ctx.Error(ErrNoRow).Error())
			return
		}
		log.Info().Msg(ctx.Error(ErrInternalServer).Error())
		return
	}

	response := roleResponse{
		ExternalID:  role.ExternalID,
		Name:        role.Name,
		Description: role.Description,
		UpdatedAt:   role.UpdatedAt,
		CreatedAt:   role.CreatedAt,
	}
	ctx.JSON(http.StatusOK, response)
}
