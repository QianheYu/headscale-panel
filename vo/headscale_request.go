package vo

import pb "github.com/juanfont/headscale/gen/go/headscale/v1"

// ApiKey start
type CreateApiKey struct {
	pb.CreateApiKeyRequest
}

// ExpireApiKey struct represents a request to expire an existing API key.
type ExpireApiKey struct {
	pb.ExpireApiKeyRequest
}

// ApiKey end

// ListPreAuthKey struct represents a request to list all pre-authorized keys.
type ListPreAuthKey struct {
	pb.ListPreAuthKeysRequest
}

// CreatePreAuthKey struct represents a request to create a new pre-authorized key. It contains an additional field 'Expire' for setting the expiration time.
type CreatePreAuthKey struct {
	pb.CreatePreAuthKeyRequest
	Expire string `json:"expire" form:"expire"`
}

// ExpirePreAuthKey struct represents a request to expire an existing pre-authorized key.
type ExpirePreAuthKey struct {
	pb.ExpirePreAuthKeyRequest
}

// PreAuthKey end

// Route start
// DeleteRouteRequest struct represents a request to delete a route.
type DeleteRouteRequest struct {
	pb.DeleteRouteRequest
}

// GetNodeRoutesRequest struct represents a request to get all routes for a node.
type GetNodeRoutesRequest struct {
	pb.GetNodeRoutesRequest
}

// RouteRequest struct represents a request for a specific route.
type RouteRequest struct {
	Id uint64 `json:"id" form:"id"`
}

// SwitchRouteRequest struct represents a request to enable or disable a route.
type SwitchRouteRequest struct {
	RouteId uint64 `json:"route_id" form:"route_id"`
	Enable  bool   `json:"enable" form:"enable"`
}

// Node start
// ListNodesRequest struct represents a request to list all nodes.
type ListNodesRequest struct {
	pb.ListNodesRequest
}

// RegisterNode struct represents a request to register a new node.
type RegisterNode struct {
	pb.RegisterNodeRequest
}

// DeleteNodeRequest struct represents a request to delete a node.
type DeleteNodeRequest struct {
	pb.DeleteNodeRequest
}

// EditNodeRequest struct represents a request to edit a node. It contains fields for NodeId, Name, State, and Nodekey.
type EditNodeRequest struct {
	NodeId  uint64 `json:"node_id" validate:"required_unless=State register"`
	Name    string `json:"name" validate:"required_if=State rename,omitempty,min=0,max=63,lowercase"`
	State   string `json:"state" validate:"required,oneof=register rename expire"`
	Nodekey string `json:"nodekey" validate:"required_if=State register"`
}

// GetNodeRequest struct represents a request to get details of a specific node.
type GetNodeRequest struct {
	pb.GetNodeRequest
}

// ExpireNodeRequest struct represents a request to expire a node.
type ExpireNodeRequest struct {
	pb.ExpireNodeRequest
}

// MoveNodeRequest struct represents a request to move a node to a different group.
type MoveNodeRequest struct {
	pb.MoveNodeRequest
}

// SetTagsRequest struct represents a request to set tags for a node.
type SetTagsRequest struct {
	pb.SetTagsRequest
}

// SetAccessControlRequest struct represents a request to set access control for a node.
type SetAccessControlRequest struct {
	Content string `json:"content" validate:"required"`
}
