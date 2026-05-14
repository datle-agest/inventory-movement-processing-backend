package http

import (
	"inventory-movement-processing/common"
	"inventory-movement-processing/pkg/core"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ImportBatch godoc
// @Summary Import inventory movements from CSV
// @Description Upload CSV file to process inventory movements in batch
// @Tags Movements
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "CSV file"
// @Success 200 {object} core.APIResponse{result=object} "Import result summary"
// @Failure 400 {object} core.APIResponse "Invalid file or request"
// @Failure 409 {object} core.APIResponse "Duplicate movement"
// @Failure 500 {object} core.APIResponse "Internal server error"
// @Router /v1/inventory-movements/import [post]
func (h *Handler) ImportBatch() gin.HandlerFunc {

	return func(c *gin.Context) {

		file, err := c.FormFile("file")

		if err != nil {

			core.WriteError(
				c,
				common.ErrBadRequest("file is required"),
			)

			return
		}

		result, err := h.service.ImportBatch(
			c.Request.Context(),
			file,
		)

		if err != nil {
			core.WriteError(c, err)
			return
		}

		c.JSON(
			http.StatusOK,
			core.Success(result),
		)
	}
}
