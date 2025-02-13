package utils

import (
	"fortis/entity/constants"
	"fortis/internal/instrumentation"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Validate and bind the request
func BindRequest(ctx *gin.Context, request interface{}, handler string, startTime time.Time) bool {
	if err := ctx.ShouldBindJSON(request); err != nil {
		HandleError(ctx, http.StatusBadRequest, constants.InvalidRequestParser, constants.InvalidRequest, err, handler, startTime)
		return false
	}
	return true
}

// Handle errors consistently
func HandleError(ctx *gin.Context, statusCode int, logMessage, userMessage string, err error, handler string, startTime time.Time) {
	log.Printf("%s: %v\n", logMessage, err)
	instrumentation.FailureRequestCounter.WithLabelValues(handler, http.StatusText(statusCode)).Inc()
	instrumentation.FailureLatency.WithLabelValues(handler, http.StatusText(statusCode)).Observe(time.Since(startTime).Seconds())
	ctx.JSON(statusCode, gin.H{"error": userMessage})
}
