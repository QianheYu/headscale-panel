package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"headscale-panel/common"
	"headscale-panel/log"
	"headscale-panel/model"
	"headscale-panel/repository"
	"headscale-panel/response"
	"headscale-panel/vo"
)

type INodesController interface {
	GetNodes(c *gin.Context)   // method: get
	StateNodes(c *gin.Context) // method: post Refactoring the request structure
	DeleteNode(c *gin.Context) // method: delete
	MoveNode(c *gin.Context)
	SetTags(c *gin.Context)
}

type NodesController struct {
	userRepo  repository.IUserRepository
	nodesRepo repository.HeadscaleNodesRepository
}

func NewNodesController() INodesController {
	return &NodesController{userRepo: repository.NewUserRepository(), nodesRepo: repository.NewNodesRepo()}
}

// GetNodes get nodes by user name
func (m *NodesController) GetNodes(c *gin.Context) {
	user, err := m.userRepo.GetCurrentUser(c)
	if err != nil {
		response.Fail(c, nil, "Failed to get Node")
		log.Log.Errorf("get current user error: %v", err)
		return
	}

	if mflag, ok := c.Get("machineFlag"); ok && mflag.(bool) {
		user.Name = ""
	}

	Nodes, err := m.nodesRepo.ListNodesWithUser(user.Name)
	if err != nil && err.Error() != "rpc error: code = Unknown desc = User not found" {
		response.Fail(c, nil, "Failed to get Nodes")
		log.Log.Errorf("get Node error: %v", err)
		return
	}
	response.Success(c, Nodes, "success")
}

// StateNodes Register, expire, and rename devices.
func (m *NodesController) StateNodes(c *gin.Context) {
	req := &vo.EditNodeRequest{}
	// Bind parameters
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, nil, "param error")
		return
	}

	// Validate parameters
	if err := common.Validate.Struct(req); err != nil {
		errStr := err.(validator.ValidationErrors)[0].Translate(common.Trans)
		response.Fail(c, nil, errStr)
		return
	}

	var err error
	var data interface{}
	switch req.State {
	case "rename":
		// rename node
		data, err = m.nodesRepo.RenameNodeWithNewName(req.NodeId, req.Name)
	case "expire":
		// expire node
		data, err = m.nodesRepo.ExpireNodeWithId(req.NodeId)
	case "register":
		// register node
		var user model.User
		user, err = m.userRepo.GetCurrentUser(c)
		if err != nil {
			break
		}
		data, err = m.nodesRepo.RegisterNodeWithKey(user.Name, req.Nodekey)
	default:
		response.Fail(c, nil, "params error")
		return
	}
	if err != nil {
		response.Fail(c, nil, "Failed to operate")
		log.Log.Errorf("operate node error: %v", err)
		return
	}
	response.Success(c, data, "success")
}

// MoveNode move node to another user
func (m *NodesController) MoveNode(c *gin.Context) {
	req := &vo.MoveNodeRequest{}

	// Bind parameters
	if err := c.ShouldBindJSON(req); err != nil {
		response.Fail(c, nil, "param error")
		return
	}

	// validate parameters
	if err := common.Validate.Struct(req); err != nil {
		errStr := err.(validator.ValidationErrors)[0].Translate(common.Trans)
		response.Fail(c, nil, errStr)
		return
	}

	Node, err := m.nodesRepo.MoveNode(req)
	if err != nil {
		response.Fail(c, nil, "Failed to move Node")
		log.Log.Errorf("move Node error: %v", err)
		return
	}
	response.Success(c, Node, "move success")
}

// DeleteNode delete node
func (m *NodesController) DeleteNode(c *gin.Context) {
	req := &vo.DeleteNodeRequest{}

	// Bind parameters
	if err := c.ShouldBindJSON(req); err != nil {
		response.Fail(c, nil, "param error")
		return
	}

	// Validate parameters
	if err := common.Validate.Struct(req); err != nil {
		errStr := err.(validator.ValidationErrors)[0].Translate(common.Trans)
		response.Fail(c, nil, errStr)
		return
	}

	if err := m.nodesRepo.DeleteNode(req); err != nil {
		response.Fail(c, nil, "Failed to delete node")
		log.Log.Errorf("delete node error: %v", err)
		return
	}
	response.Success(c, nil, "success")
}

// SetTags set tag on node
func (m *NodesController) SetTags(c *gin.Context) {
	req := &vo.SetTagsRequest{}

	// Bind parameters
	if err := c.ShouldBindJSON(req); err != nil {
		response.Fail(c, nil, "param error")
		return
	}

	// Validate parameters
	if err := common.Validate.Struct(req); err != nil {
		errStr := err.(validator.ValidationErrors)[0].Translate(common.Trans)
		response.Fail(c, nil, errStr)
		return
	}

	data, err := m.nodesRepo.SetTagsWithStringSlice(req.NodeId, req.Tags)
	if err != nil {
		response.Fail(c, nil, "Failed to set tag")
		log.Log.Errorf("set tag error: %v", err)
		return
	}
	response.Success(c, data, "set tags success")
}
