package http

import (
	"inventory-movement-processing/common"
	"inventory-movement-processing/pkg/core"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ImportBatch godoc
// @Summary Import inventory movements from a CSV file
// @Description Uploads a CSV file containing stock movements (IN, OUT, ADJUST) to process in batch.
// @Description The engine validates the CSV format and groups rows by item ID.
// @Description Movements are then processed concurrently via a worker pool, ensuring non-negative stock limits and avoiding duplicate external IDs.
// @Description Returns a summary of the batch import execution including success/fail counts and row-level details.
// @Tags Movements
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "CSV file to import (required, max size: 5MB)"
// @Success 200 {object} core.APIResponse{result=entity.ImportBatchResult}
// @Failure 400 {object} core.ErrResponseBadRequest
// @Failure 409 {object} core.ErrResponseConflict
// @Failure 500 {object} core.ErrResponseInternal
// @Router /v1/inventory-movements/import [post]
// @Security BearerAuth
func (h *Handler) ImportBatch() gin.HandlerFunc {

	return func(c *gin.Context) {

		file, err := c.FormFile("file")

		if err != nil {
			core.WriteError(c, common.ErrBadRequest("file is required"))
			return
		}

		result, err := h.service.ImportBatch(c.Request.Context(), file)

		if err != nil {
			core.WriteError(c, err)
			return
		}

		c.JSON(
			http.StatusOK,
			core.SuccessWithMessage(result, "import batch success"),
		)
	}
}
