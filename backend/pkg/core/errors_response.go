package core

import (
	"errors"
	"fmt"
	"inventory-movement-processing/common"
	"net/http"

	"github.com/gin-gonic/gin"
)

func WriteError(c *gin.Context, err error) {
	var appErr *common.AppError
	if errors.As(err, &appErr) {
		if appErr.StatusCode == http.StatusInternalServerError {
			fmt.Printf("[ERROR] %s: %v\n", c.Request.URL.Path, err)
		}
		c.JSON(appErr.StatusCode, Fail(appErr.StatusCode, appErr.Message))
		return
	}
	// err nào không phải AppError log lại xem bug
	fmt.Printf("[ERROR] unhandled error %s: %v\n", c.Request.URL.Path, err)
	c.JSON(http.StatusInternalServerError, Fail(http.StatusInternalServerError, "internal server error"))
}
