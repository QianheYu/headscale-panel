package controller

import (
	"headscale-panel/common"
	"headscale-panel/log"
	"headscale-panel/repository"
	"headscale-panel/response"
	"headscale-panel/vo"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type IOperationLogController interface {
	GetOperationLogs(c *gin.Context)             // GetOperationLogs
	BatchDeleteOperationLogByIds(c *gin.Context) // BatchDeleteOperationLogByIds
}

type OperationLogController struct {
	operationLogRepository repository.IOperationLogRepository
}

func NewOperationLogController() IOperationLogController {
	operationLogRepository := repository.NewOperationLogRepository()
	operationLogController := OperationLogController{operationLogRepository: operationLogRepository}
	return operationLogController
}

// GetOperationLogs
func (oc OperationLogController) GetOperationLogs(c *gin.Context) {
	var req vo.OperationLogListRequest
	// Bind parameters
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, nil, "param error")
		return
	}
	// Validate parameters
	if err := common.Validate.Struct(&req); err != nil {
		errStr := err.(validator.ValidationErrors)[0].Translate(common.Trans)
		response.Fail(c, nil, errStr)
		return
	}
	// Get logs
	logs, total, err := oc.operationLogRepository.GetOperationLogs(&req)
	if err != nil {
		response.Fail(c, nil, "Failed to get operation logs")
		log.Log.Errorf("get operation logs: %v", err)
		return
	}
	response.Success(c, gin.H{"logs": logs, "total": total}, "Successfully to get operation logs")
}

// BatchDeleteOperationLogByIds
func (oc OperationLogController) BatchDeleteOperationLogByIds(c *gin.Context) {
	var req vo.DeleteOperationLogRequest
	// Bind parameters
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, nil, "param error")
		return
	}
	// Validate parameters
	if err := common.Validate.Struct(&req); err != nil {
		errStr := err.(validator.ValidationErrors)[0].Translate(common.Trans)
		response.Fail(c, nil, errStr)
		return
	}

	if len(req.OperationLogIds) == 0 {
		// Delete all
		err := oc.operationLogRepository.DeleteAllOperationLog()
		if err != nil {
			response.Fail(c, nil, "Failed to delete logs")
			log.Log.Errorf("delete logs error: %v", err)
			return
		}
	} else {
		// Delete logs
		err := oc.operationLogRepository.BatchDeleteOperationLogByIds(req.OperationLogIds)
		if err != nil {
			response.Fail(c, nil, "Failed to delete logs")
			log.Log.Errorf("delete logs error: %v", err)
			return
		}
	}

	response.Success(c, nil, "Successfully delete logs")
}
