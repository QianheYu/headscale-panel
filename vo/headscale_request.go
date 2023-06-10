package vo

import pb "headscale-panel/gen/headscale/v1"

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

// GetMachineRoutesRequest struct represents a request to get all routes for a machine.
type GetMachineRoutesRequest struct {
	pb.GetMachineRoutesRequest
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

// Machine start
// ListMachinesRequest struct represents a request to list all machines.
type ListMachinesRequest struct {
	pb.ListMachinesRequest
}

// RegisterMachine struct represents a request to register a new machine.
type RegisterMachine struct {
	pb.RegisterMachineRequest
}

// DeleteMachineRequest struct represents a request to delete a machine.
type DeleteMachineRequest struct {
	pb.DeleteMachineRequest
}

// EditMachineRequest struct represents a request to edit a machine. It contains fields for MachineId, Name, State, and Nodekey.
type EditMachineRequest struct {
	MachineId uint64 `json:"machine_id" validate:"required_unless=State register"`
	Name      string `json:"name" validate:"min=0,max=63,lowercase,required_if=State rename"`
	State     string `json:"state" validate:"required,oneof=register rename expire"`
	Nodekey   string `json:"nodekey" validate:"required_if=State register"`
}

// GetMachineRequest struct represents a request to get details of a specific machine.
type GetMachineRequest struct {
	pb.GetMachineRequest
}

// ExpireMachineRequest struct represents a request to expire a machine.
type ExpireMachineRequest struct {
	pb.ExpireMachineRequest
}

// MoveMachineRequest struct represents a request to move a machine to a different group.
type MoveMachineRequest struct {
	pb.MoveMachineRequest
}

// SetTagsRequest struct represents a request to set tags for a machine.
type SetTagsRequest struct {
	pb.SetTagsRequest
}

// SetAccessControlRequest struct represents a request to set access control for a machine.
type SetAccessControlRequest struct {
	Content string `json:"content" validate:"required"`
}
