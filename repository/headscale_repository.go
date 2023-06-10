package repository

import (
	"context"
	"github.com/patrickmn/go-cache"
	"google.golang.org/protobuf/types/known/timestamppb"
	pb "headscale-panel/gen/headscale/v1"
	"headscale-panel/log"
	task "headscale-panel/tasks"
	"headscale-panel/vo"
	"strconv"
	"time"
)

// The task.HeadscaleControl Inherited from grpc client
// task.HeadscaleControl can be used to call the grpc client to make a request

// Cache for different interfaces
var apiCache = cache.New(cache.DefaultExpiration, time.Hour)
var preAuthKeyCache = cache.New(cache.DefaultExpiration, time.Hour)
var machineCache = cache.New(time.Minute, time.Minute)
var routeCache = cache.New(time.Minute, time.Minute)
var usersCache = cache.New(24*time.Hour, 48*time.Hour) // usersCache is a cache that stores user information with expiration time of 24 hours.
var syncUserDataChan = make(chan bool, 1)              // syncUserDataChan is a channel used for synchronizing user data.

// userRepo is a repository that handles user information.
var userRepo = NewUserRepository()

// HeadscaleApiKeyRepository is an interface for managing API keys.
type HeadscaleApiKeyRepository interface {
	ListApiKeys() ([]*pb.ApiKey, error)
	CreateApiKey(key *vo.CreateApiKey) (string, error)
	CreateApiKeyWithTimestamp(timestamp *timestamppb.Timestamp) (string, error)
	ExpireApiKey(key *vo.ExpireApiKey) error
	ExpireApiKeyWithString(perfix string) error
}

// HeadscalePreAuthKeyRepository is an interface for managing pre-authorized keys.
type HeadscalePreAuthKeyRepository interface {
	ListPreAuthKey(key *vo.ListPreAuthKey) ([]*pb.PreAuthKey, error)
	ListPreAuthKeyWithString(user string) ([]*pb.PreAuthKey, error)
	CreatePreAuthKey(key *vo.CreatePreAuthKey) (*pb.PreAuthKey, error)
	ExpirePreAuthKey(key *vo.ExpirePreAuthKey) error
	ExpirePreAuthKeyWithString(user, key string) error
}

// HeadscaleUserRepository is an interface for managing user information.
type HeadscaleUserRepository interface {
	ListUser() ([]*pb.User, error)
	GetUserWithString(name string) (*pb.User, error)
	CreateUserWithString(name string) (*pb.User, error)
	DeleteUserWithString(name string) error
	RenameUserWithString(oldname, newname string) (*pb.User, error)
}

// HeadscaleRouteRepository is an interface for managing route information.
type HeadscaleRouteRepository interface {
	GetRoutes() ([]*pb.Route, error)
	GetMachineRoutes(request *vo.GetMachineRoutesRequest) ([]*pb.Route, error)
	GetMachineRoutesWithId(machineId uint64) ([]*pb.Route, error)
	DeleteRoute(request *vo.DeleteRouteRequest) error
	DeleteRouteWithId(routeId uint64) error
	SwitchRoute(request *vo.SwitchRouteRequest) error
	SwitchRouteWithId(routeId uint64, enable bool) error
}

// HeadscaleMachinesRepository is an interface for managing machine information.
type HeadscaleMachinesRepository interface {
	ListMachines(user *vo.ListMachinesRequest) ([]*pb.Machine, error)
	ListMachinesWithUser(user string) ([]*pb.Machine, error)
	GetMachine(getMachine *vo.GetMachineRequest) (*pb.Machine, error)
	GetMachineWithId(machineId uint64) (*pb.Machine, error)
	ExpireMachine(expireMachine *vo.ExpireMachineRequest) (*pb.Machine, error)
	ExpireMachineWithId(machineId uint64) (*pb.Machine, error)
	RenameMachine(machineId uint64, name string) (*pb.Machine, error)
	RenameMachineWithNewName(machineId uint64, name string) (*pb.Machine, error)
	MoveMachine(machine *vo.MoveMachineRequest) (*pb.Machine, error)
	MoveMachineWithUser(machineId uint64, user string) (*pb.Machine, error)
	DeleteMachine(machine *vo.DeleteMachineRequest) error
	DeleteMachineWithId(machineId uint64) error
	RegisterMachine(newMachine *vo.RegisterMachine) (*pb.Machine, error)
	RegisterMachineWithKey(user, key string) (*pb.Machine, error)
	SetTags(tags *vo.SetTagsRequest) (*pb.Machine, error)
	SetTagsWithStringSlice(machineId uint64, tags []string) (*pb.Machine, error)
	SetTagsWithStrings(machineId uint64, tags ...string) (*pb.Machine, error)
}

// headscaleRepository is a struct that implements all the repository interfaces.
type headscaleRepository struct{}

// NewApiKeyRepo returns a new instance of HeadscaleApiKeyRepository.
func NewApiKeyRepo() HeadscaleApiKeyRepository {
	return &headscaleRepository{}
}

// NewPreAuthkeyRepo returns a new instance of HeadscalePreAuthKeyRepository.
func NewPreAuthkeyRepo() HeadscalePreAuthKeyRepository {
	return &headscaleRepository{}
}

// NewUserRepo returns a new instance of HeadscaleUserRepository.
func NewUserRepo() HeadscaleUserRepository {
	return &headscaleRepository{}
}

// NewRouteRepo returns a new instance of HeadscaleRouteRepository.
func NewRouteRepo() HeadscaleRouteRepository {
	return &headscaleRepository{}
}

// NewMachinesRepo returns a new instance of HeadscaleMachinesRepository.
func NewMachinesRepo() HeadscaleMachinesRepository {
	return &headscaleRepository{}
}

// ApiKeys start

// ListApiKeys retrieves a list of API keys from the cache or from the HeadscaleControl task.
// If the list is retrieved from the cache, it returns the cached list. Otherwise, it retrieves the list from the task.
// It also adds the retrieved list to the cache for future use.
func (h *headscaleRepository) ListApiKeys() ([]*pb.ApiKey, error) {
	if list, ok := apiCache.Get("apikey"); ok {
		return list.([]*pb.ApiKey), nil
	}
	list, err := task.HeadscaleControl.ListApiKeys(context.Background(), &pb.ListApiKeysRequest{})
	if err != nil {
		return nil, err
	}
	if err = apiCache.Add("apikey", list.ApiKeys, cache.DefaultExpiration); err != nil {
		log.Log.Errorf("add api keys to cache error:%v", err)
	}
	return list.ApiKeys, nil
}

// CreateApiKey creates a new API key using the specified request and returns the new API key.
func (h *headscaleRepository) CreateApiKey(key *vo.CreateApiKey) (string, error) {
	newkey, err := task.HeadscaleControl.CreateApiKey(context.Background(), &key.CreateApiKeyRequest)
	if err != nil {
		return "", err
	}
	return newkey.ApiKey, nil
}

// CreateApiKeyWithTimestamp creates a new API key with the specified expiration timestamp and returns the new API key.
// It also deletes the cached list of API keys.
func (h *headscaleRepository) CreateApiKeyWithTimestamp(timestamp *timestamppb.Timestamp) (string, error) {
	newkey, err := task.HeadscaleControl.CreateApiKey(context.Background(), &pb.CreateApiKeyRequest{Expiration: timestamp})
	if err != nil {
		return "", err
	}
	apiCache.Delete("apikey")
	return newkey.ApiKey, nil
}

// ExpireApiKey expires an API key using the specified request.
// It also deletes the cached list of API keys.
func (h *headscaleRepository) ExpireApiKey(key *vo.ExpireApiKey) (err error) {
	_, err = task.HeadscaleControl.ExpireApiKey(context.Background(), &key.ExpireApiKeyRequest)
	apiCache.Delete("apikey")
	return
}

// ExpireApiKeyWithString expires all API keys with the specified prefix.
// It also deletes the cached list of API keys.
func (h *headscaleRepository) ExpireApiKeyWithString(prefix string) (err error) {
	_, err = task.HeadscaleControl.ExpireApiKey(context.Background(), &pb.ExpireApiKeyRequest{Prefix: prefix})
	apiCache.Delete("apikey")
	return
}

// ApiKeys end

// PreAuthKey start

// ListPreAuthKey
// This method takes a pointer to a vo.ListPreAuthKey struct
// and returns a slice of pb.PreAuthKey and an error.
// It first checks if the preAuthKeyCache has a key "preAuthKey".
// If it does, it returns the value as a slice of pb.PreAuthKey.
// If not, it calls task.HeadscaleControl.ListPreAuthKeys to get the list of pre-authenticated keys
// and adds the list to the cache with a default expiration time.
// It then returns the list and any error that occurred.
func (h *headscaleRepository) ListPreAuthKey(key *vo.ListPreAuthKey) ([]*pb.PreAuthKey, error) {
	if list, ok := preAuthKeyCache.Get("preAuthKey"); ok {
		return list.([]*pb.PreAuthKey), nil
	}
	list, err := task.HeadscaleControl.ListPreAuthKeys(context.Background(), &key.ListPreAuthKeysRequest)
	if err != nil {
		return nil, err
	}
	if err = apiCache.Add("preAuthKey", list.PreAuthKeys, cache.DefaultExpiration); err != nil {
		log.Log.Errorf("add preAuthKeys to cache error:%v", err)
	}
	return list.PreAuthKeys, err
}

// ListPreAuthKeyWithString This method takes a string user
// and returns a slice of pb.PreAuthKey and an error. It calls
// task.HeadscaleControl.ListPreAuthKeys with a ListPreAuthKeysRequest
// that has the user field set to the given user. It then returns the list and any error that occurred.
func (h *headscaleRepository) ListPreAuthKeyWithString(user string) ([]*pb.PreAuthKey, error) {

	list, err := task.HeadscaleControl.ListPreAuthKeys(context.Background(), &pb.ListPreAuthKeysRequest{User: user})
	if err != nil {
		return nil, err
	}
	return list.PreAuthKeys, err
}

// CreatePreAuthKey This method takes a pointer to a vo.CreatePreAuthKey struct
// and returns a pointer to a pb.PreAuthKey and an error. It calls
// task.HeadscaleControl.CreatePreAuthKey with a CreatePreAuthKeyRequest based on the given struct.
// If the call is successful, it deletes the "preAuthKey" key from the cache
// and returns the pre-authenticated key and any error that occurred.
func (h *headscaleRepository) CreatePreAuthKey(key *vo.CreatePreAuthKey) (*pb.PreAuthKey, error) {
	req, err := task.HeadscaleControl.CreatePreAuthKey(context.Background(), &key.CreatePreAuthKeyRequest)
	if err != nil {
		return nil, err
	}
	preAuthKeyCache.Delete("preAuthKey")
	return req.PreAuthKey, nil
}

// ExpirePreAuthKey This method takes a pointer to a vo.ExpirePreAuthKey struct
// and returns an error. It calls task.HeadscaleControl.ExpirePreAuthKey with
// an ExpirePreAuthKeyRequest based on the given struct. It then deletes
// the "preAuthKey" key from the cache and returns any error that occurred.
func (h *headscaleRepository) ExpirePreAuthKey(key *vo.ExpirePreAuthKey) (err error) {
	_, err = task.HeadscaleControl.ExpirePreAuthKey(context.Background(), &key.ExpirePreAuthKeyRequest)
	preAuthKeyCache.Delete("preAuthKey")
	return
}

// ExpirePreAuthKeyWithString This method takes two strings, user and key, and returns an error. It calls task.HeadscaleControl.ExpirePreAuthKey with an ExpirePreAuthKeyRequest that has the user and key fields set to the given values. It then deletes the "preAuthKey" key from the cache and returns any error that occurred.
func (h *headscaleRepository) ExpirePreAuthKeyWithString(user, key string) (err error) {
	_, err = task.HeadscaleControl.ExpirePreAuthKey(context.Background(), &pb.ExpirePreAuthKeyRequest{User: user, Key: key})
	preAuthKeyCache.Delete("preAuthKey")
	return
}

// PreAuthKey end

// User start

// ListUser This method returns a slice of pb.User and an error.
// It calls task.HeadscaleControl.ListUsers to get the list of users.
// If the call is successful, it adds the list to the usersCache
// with a default expiration time and returns the list and any error
// that occurred. It also sends a true value to the syncUserDataChan channel
// to trigger a data consistency check and synchronization.
func (h *headscaleRepository) ListUser() ([]*pb.User, error) {
	list, err := task.HeadscaleControl.ListUsers(context.Background(), &pb.ListUsersRequest{})
	if err != nil {
		return nil, err
	}
	select {
	case syncUserDataChan <- true:
		go func() {
			if _, found := usersCache.Get("src"); found {
				usersCache.Delete("src")
			}
			usersCache.Set("src", list, cache.DefaultExpiration)

			// Data consistency checks and synchronisation
		}()
	}

	return list.Users, nil
}

// GetUserWithString This method takes a string name and returns a pointer
// to a pb.User and an error. It calls task.HeadscaleControl.GetUser
// with a GetUserRequest that has the name field set to the given name.
// If the call is successful, it returns the user and any error that occurred.
func (h *headscaleRepository) GetUserWithString(name string) (*pb.User, error) {
	req, err := task.HeadscaleControl.GetUser(context.Background(), &pb.GetUserRequest{Name: name})
	if err != nil {
		return nil, err
	}
	return req.User, nil
}

// CreateUserWithString This method takes a string name and returns a pointer
// to a pb.User and an error. It calls task.HeadscaleControl.CreateUser
// with a CreateUserRequest that has the name field set to the given name.
// If the call is successful, it returns the created user and any error that occurred.
func (h *headscaleRepository) CreateUserWithString(name string) (*pb.User, error) {
	req, err := task.HeadscaleControl.CreateUser(context.Background(), &pb.CreateUserRequest{Name: name})
	if err != nil {
		return nil, err
	}
	return req.User, nil
}

// DeleteUserWithString This method takes a string name and returns an error.
// It calls task.HeadscaleControl.DeleteUser with a DeleteUserRequest
// that has the name field set to the given name. It then returns any error that occurred.
func (h *headscaleRepository) DeleteUserWithString(name string) (err error) {
	_, err = task.HeadscaleControl.DeleteUser(context.Background(), &pb.DeleteUserRequest{Name: name})
	return
}

// RenameUserWithString This method takes two strings, oldname and newname,
// and returns a pointer to a pb.User and an error. It calls
// task.HeadscaleControl.RenameUser with a RenameUserRequest
// that has the oldname and newname fields set to the given values.
// If the call is successful, it returns the renamed user and any error that occurred.
func (h *headscaleRepository) RenameUserWithString(oldname, newname string) (*pb.User, error) {
	req, err := task.HeadscaleControl.RenameUser(context.Background(), &pb.RenameUserRequest{
		OldName: oldname,
		NewName: newname,
	})
	if err != nil {
		return nil, err
	}
	return req.User, nil
}

// User end

// Route start

// GetRoutes This method returns a slice of pb.Route and an error.
// If the routes are present in the routeCache, it returns them.
// Otherwise, it calls task.HeadscaleControl.GetRoutes to get the routes.
// If the call is successful, it adds the routes to the routeCache
// with a default expiration time and returns the routes and any error that occurred.
func (h *headscaleRepository) GetRoutes() ([]*pb.Route, error) {
	if routes, ok := routeCache.Get("routes"); ok {
		return routes.([]*pb.Route), nil
	}
	routes, err := task.HeadscaleControl.GetRoutes(context.Background(), &pb.GetRoutesRequest{})
	if err != nil {
		return nil, err
	}
	if err = routeCache.Add("routes", routes.Routes, cache.DefaultExpiration); err != nil {
		log.Log.Errorf("add routes to cache error:%v", err)
	}
	return routes.Routes, nil
}

// DeleteRoute This method takes a DeleteRouteRequest and returns an error.
// It calls task.HeadscaleControl.DeleteRoute with the given request.
// If the call is successful, it deletes the routes from the routeCache and
// returns any error that occurred.
func (h *headscaleRepository) DeleteRoute(request *vo.DeleteRouteRequest) (err error) {
	_, err = task.HeadscaleControl.DeleteRoute(context.Background(), &request.DeleteRouteRequest)
	routeCache.Delete("routes")
	return
}

// DeleteRouteWithId This method takes a routeId and returns an error.
// It calls task.HeadscaleControl.DeleteRoute with a DeleteRouteRequest
// that has the routeId field set to the given value. If the call is successful,
// it deletes the routes from the routeCache and returns any error that occurred.
func (h *headscaleRepository) DeleteRouteWithId(routeId uint64) (err error) {
	_, err = task.HeadscaleControl.DeleteRoute(context.Background(), &pb.DeleteRouteRequest{RouteId: routeId})
	routeCache.Delete("routes")
	return
}

// GetMachineRoutes This method takes a GetMachineRoutesRequest
// and returns a slice of pb.Route and an error. If the routes
// for the given machineId are present in the routeCache, it returns them.
// Otherwise, it calls task.HeadscaleControl.GetMachineRoutes to get the routes.
// If the call is successful, it adds the routes to the routeCache with an expiration time
// of one minute and returns the routes and any error that occurred.
func (h *headscaleRepository) GetMachineRoutes(request *vo.GetMachineRoutesRequest) ([]*pb.Route, error) {
	if routes, ok := routeCache.Get(strconv.FormatUint(request.MachineId, 10)); ok {
		return routes.([]*pb.Route), nil
	}
	routes, err := task.HeadscaleControl.GetMachineRoutes(context.Background(), &request.GetMachineRoutesRequest)
	if err != nil {
		return nil, err
	}
	if len(routes.Routes) > 0 && routes.Routes[0].Machine != nil {
		if err = routeCache.Add(strconv.FormatUint(routes.Routes[0].Machine.Id, 10), routes.Routes, time.Minute); err != nil {
			log.Log.Errorf("add routes by machineid to cache error:%v", err)
		}
	}
	return routes.Routes, err
}

// GetMachineRoutesWithId This method takes a machineId and returns a slice of pb.Route and an error.
// If the routes for the given machineId are present in the routeCache, it returns them.
// Otherwise, it calls task.HeadscaleControl.GetMachineRoutes to get the routes. If the call is successful,
// it adds the routes to the routeCache with an expiration time of one minute and returns the routes and any error that occurred.
func (h *headscaleRepository) GetMachineRoutesWithId(machineId uint64) ([]*pb.Route, error) {
	if routes, ok := routeCache.Get(strconv.FormatUint(machineId, 10)); ok {
		return routes.([]*pb.Route), nil
	}
	routes, err := task.HeadscaleControl.GetMachineRoutes(context.Background(), &pb.GetMachineRoutesRequest{MachineId: machineId})
	if err != nil {
		return nil, err
	}
	if len(routes.Routes) > 0 && routes.Routes[0].Machine != nil {
		if err = routeCache.Add(strconv.FormatUint(routes.Routes[0].Machine.Id, 10), routes.Routes, time.Minute); err != nil {
			log.Log.Errorf("add routes by machineid to cache error:%v", err)
		}
	}
	return routes.Routes, err
}

// SwitchRoute This method takes a SwitchRouteRequest and returns an error.
// If the enable field is true, it calls task.HeadscaleControl.EnableRoute with the given routeId.
// Otherwise, it calls task.HeadscaleControl.DisableRoute with the given routeId.
// If the call is successful, it flushes the routeCache and returns any error that occurred.
func (h *headscaleRepository) SwitchRoute(request *vo.SwitchRouteRequest) (err error) {
	if request.Enable {
		_, err = task.HeadscaleControl.EnableRoute(context.Background(), &pb.EnableRouteRequest{RouteId: request.RouteId})
	} else {
		_, err = task.HeadscaleControl.DisableRoute(context.Background(), &pb.DisableRouteRequest{RouteId: request.RouteId})
	}
	//routeCache.Delete(strconv.FormatUint(request.RouteId, 10))
	//routeCache.Delete("routes")
	routeCache.Flush()
	return
}

// SwitchRouteWithId This method takes a routeId and a boolean enable and returns an error.
// If the enable field is true, it calls task.HeadscaleControl.EnableRoute with the given routeId.
// Otherwise, it calls task.HeadscaleControl.DisableRoute with the given routeId.
// If the call is successful, it flushes the routeCache and returns any error that occurred.
func (h *headscaleRepository) SwitchRouteWithId(routeId uint64, enable bool) (err error) {
	if enable {
		_, err = task.HeadscaleControl.EnableRoute(context.Background(), &pb.EnableRouteRequest{RouteId: routeId})
	} else {
		_, err = task.HeadscaleControl.DisableRoute(context.Background(), &pb.DisableRouteRequest{RouteId: routeId})
	}
	//routeCache.Delete(strconv.FormatUint(routeId, 10))
	//routeCache.Delete("routes")
	routeCache.Flush()
	return
}

// Route end

// Machine start

// ListMachines retrieves a list of machines for a given user, either from cache or by calling the HeadscaleControl API.
// If the machines are retrieved from cache, they are returned immediately.
// Otherwise, the HeadscaleControl API is called with the user's ListMachinesRequest, and the resulting machines are added to cache before being returned.
// If there is an error retrieving the machines, an error is returned.
func (h *headscaleRepository) ListMachines(user *vo.ListMachinesRequest) ([]*pb.Machine, error) {
	if machines, ok := machineCache.Get(user.User); ok {
		return machines.([]*pb.Machine), nil
	}
	machines, err := task.HeadscaleControl.ListMachines(context.Background(), &user.ListMachinesRequest)
	if err != nil {
		return nil, err
	}
	if err = machineCache.Add(user.User, machines.Machines, cache.DefaultExpiration); err != nil {
		log.Log.Errorf("add machines to cache error:%v", err)
	}
	return machines.Machines, nil
}

// ListMachinesWithUser retrieves a list of machines for a given user, either from cache or by calling the HeadscaleControl API.
// If the machines are retrieved from cache, they are returned immediately.
// Otherwise, the HeadscaleControl API is called with a ListMachinesRequest containing the user's name, and the resulting machines are added to cache before being returned.
// If there is an error retrieving the machines, an error is returned.
func (h *headscaleRepository) ListMachinesWithUser(user string) ([]*pb.Machine, error) {
	if machines, ok := machineCache.Get(user); ok {
		return machines.([]*pb.Machine), nil
	}
	machines, err := task.HeadscaleControl.ListMachines(context.Background(), &pb.ListMachinesRequest{User: user})
	if err != nil {
		return nil, err
	}
	if err = machineCache.Add(user, machines.Machines, cache.DefaultExpiration); err != nil {
		log.Log.Errorf("add machines to cache error:%v", err)
	}
	return machines.Machines, nil
}

// GetMachine retrieves a machine with a given ID, either from cache or by calling the HeadscaleControl API.
// If the machine is retrieved from cache, it is returned immediately.
// Otherwise, the HeadscaleControl API is called with the GetMachineRequest, and the resulting machine is added to cache before being returned.
// If there is an error retrieving the machine, an error is returned.
func (h *headscaleRepository) GetMachine(getMachine *vo.GetMachineRequest) (*pb.Machine, error) {
	if machine, ok := machineCache.Get(strconv.FormatUint(getMachine.MachineId, 10)); ok {
		return machine.(*pb.Machine), nil
	}
	machines, err := task.HeadscaleControl.GetMachine(context.Background(), &getMachine.GetMachineRequest)
	if err != nil {
		return nil, err
	}
	if err = machineCache.Add(strconv.FormatUint(machines.Machine.Id, 10), machines.Machine, cache.DefaultExpiration); err != nil {
		log.Log.Errorf("add machine to cache error:%v", err)
	}
	return machines.Machine, err
}

// GetMachineWithId retrieves a machine with a given ID, either from cache or by calling the HeadscaleControl API.
// If the machine is retrieved from cache, it is returned immediately.
// Otherwise, the HeadscaleControl API is called with a GetMachineRequest containing the machine ID, and the resulting machine is added to cache before being returned.
// If there is an error retrieving the machine, an error is returned.
func (h *headscaleRepository) GetMachineWithId(machineId uint64) (*pb.Machine, error) {
	if machine, ok := machineCache.Get(strconv.FormatUint(machineId, 10)); ok {
		return machine.(*pb.Machine), nil
	}
	machine, err := task.HeadscaleControl.GetMachine(context.Background(), &pb.GetMachineRequest{MachineId: machineId})
	if err != nil {
		return nil, err
	}
	if err = machineCache.Add(strconv.FormatUint(machine.Machine.Id, 10), machine.Machine, cache.DefaultExpiration); err != nil {
		log.Log.Errorf("add machine to cache error:%v", err)
	}
	return machine.Machine, err
}

// ExpireMachine sets a machine's expiration date to the current time, effectively expiring the machine.
// The HeadscaleControl API is called with the ExpireMachineRequest, and the resulting machine is returned.
// If there is an error expiring the machine, an error is returned.
func (h *headscaleRepository) ExpireMachine(expireMachine *vo.ExpireMachineRequest) (*pb.Machine, error) {
	machine, err := task.HeadscaleControl.ExpireMachine(context.Background(), &expireMachine.ExpireMachineRequest)
	if err != nil {
		return nil, err
	}
	machineCache.Delete("")
	machineCache.Delete(machine.Machine.User.Name)
	return machine.Machine, err
}

// ExpireMachineWithId sets a machine's expiration date to the current time, effectively expiring the machine.
// The HeadscaleControl API is called with an ExpireMachineRequest containing the machine ID, and the resulting machine is returned.
// If there is an error expiring the machine, an error is returned.
func (h *headscaleRepository) ExpireMachineWithId(machineId uint64) (*pb.Machine, error) {
	machine, err := task.HeadscaleControl.ExpireMachine(context.Background(), &pb.ExpireMachineRequest{MachineId: machineId})
	if err != nil {
		return nil, err
	}
	machineCache.Delete("")
	machineCache.Delete(machine.Machine.User.Name)
	return machine.Machine, err
}

// RenameMachine renames a machine with a given ID to a new name.
// The HeadscaleControl API is called with a RenameMachineRequest containing the machine ID and new name, and the resulting machine is returned.
// If there is an error renaming the machine, an error is returned.
func (h *headscaleRepository) RenameMachine(machineId uint64, name string) (*pb.Machine, error) {
	machine, err := task.HeadscaleControl.RenameMachine(context.Background(), &pb.RenameMachineRequest{
		MachineId: machineId,
		NewName:   name,
	})
	if err != nil {
		return nil, err
	}
	machineCache.Delete("")
	machineCache.Delete(machine.Machine.User.Name)
	return machine.Machine, nil
}

// RenameMachineWithNewName renames a machine with a given ID to a new name.
// The HeadscaleControl API is called with a RenameMachineRequest containing the machine ID and new name, and the resulting machine is returned.
// If there is an error renaming the machine, an error is returned.
func (h *headscaleRepository) RenameMachineWithNewName(machineId uint64, name string) (*pb.Machine, error) {
	machine, err := task.HeadscaleControl.RenameMachine(context.Background(), &pb.RenameMachineRequest{
		MachineId: machineId,
		NewName:   name,
	})
	if err != nil {
		return nil, err
	}
	machineCache.Delete("")
	machineCache.Delete(machine.Machine.User.Name)
	return machine.Machine, nil
}

// MoveMachine moves a machine to a new user.
// The HeadscaleControl API is called with the MoveMachineRequest, and the resulting machine is returned.
// If there is an error moving the machine, an error is returned.
func (h *headscaleRepository) MoveMachine(machine *vo.MoveMachineRequest) (*pb.Machine, error) {
	movedMachine, err := task.HeadscaleControl.MoveMachine(context.Background(), &machine.MoveMachineRequest)
	if err != nil {
		return nil, err
	}
	machineCache.Flush()
	//machineCache.Delete("")
	// cannot delete old user's cache
	//machineCache.Delete(machine.User)
	//machineCache.Delete(movedMachine.Machine.User.Name)
	return movedMachine.Machine, nil
}

// MoveMachineWithUser moves a machine to a new user.
// The HeadscaleControl API is called with a MoveMachineRequest containing the machine ID and new user, and the resulting machine is returned.
// If there is an error moving the machine, an error is returned.
func (h *headscaleRepository) MoveMachineWithUser(machineId uint64, user string) (*pb.Machine, error) {
	machine, err := task.HeadscaleControl.MoveMachine(context.Background(), &pb.MoveMachineRequest{
		MachineId: machineId,
		User:      user,
	})
	if err != nil {
		return nil, err
	}
	machineCache.Flush()
	//machineCache.Delete("")
	// cannot delete old user's cache
	//machineCache.Delete(user)
	//machineCache.Delete(machine.Machine.User.Name)
	return machine.Machine, nil
}

// DeleteMachine deletes a machine with a given ID.
// The HeadscaleControl API is called with the DeleteMachineRequest, and any cached machines are deleted.
// If there is an error deleting the machine, an error is returned.
func (h *headscaleRepository) DeleteMachine(machine *vo.DeleteMachineRequest) (err error) {
	_, err = task.HeadscaleControl.DeleteMachine(context.Background(), &machine.DeleteMachineRequest)
	machineCache.Delete("")
	// cannot delete user's cache by ID
	//machineCache.Delete(machine.Machine.User.Name)
	return
}

// DeleteMachineWithId deletes a machine with a given ID.
// The HeadscaleControl API is called with a DeleteMachineRequest containing the machine ID, and any cached machines are deleted.
// If there is an error deleting the machine, an error is returned.
func (h *headscaleRepository) DeleteMachineWithId(machineId uint64) (err error) {
	_, err = task.HeadscaleControl.DeleteMachine(context.Background(), &pb.DeleteMachineRequest{MachineId: machineId})
	machineCache.Delete("")
	// cannot delete user's cache by machineId
	//machineCache.Delete(machine.Machine.User.Name)
	return
}

// RegisterMachine registers a new machine with the given details.
// The HeadscaleControl API is called with the RegisterMachineRequest, and any cached machines are deleted.
// If there is an error registering the machine, an error is returned.
func (h *headscaleRepository) RegisterMachine(newMachine *vo.RegisterMachine) (*pb.Machine, error) {
	machine, err := task.HeadscaleControl.RegisterMachine(context.Background(), &newMachine.RegisterMachineRequest)
	if err != nil {
		return nil, err
	}
	machineCache.Delete("")
	machineCache.Delete(machine.Machine.User.Name)
	return machine.Machine, nil
}

// RegisterMachineWithKey registers a new machine with the given user and key.
// The HeadscaleControl API is called with a RegisterMachineRequest containing the user and key, and any cached machines are deleted.
// If there is an error registering the machine, an error is returned.
func (h *headscaleRepository) RegisterMachineWithKey(user, key string) (*pb.Machine, error) {
	machine, err := task.HeadscaleControl.RegisterMachine(context.Background(), &pb.RegisterMachineRequest{
		User: user,
		Key:  key,
	})
	if err != nil {
		return nil, err
	}
	machineCache.Delete("")
	machineCache.Delete(machine.Machine.User.Name)
	return machine.Machine, nil
}

// SetTags updates the tags of a machine with the given SetTagsRequest.
// It returns the updated machine and an error if any.
func (h *headscaleRepository) SetTags(tags *vo.SetTagsRequest) (*pb.Machine, error) {
	// Call the HeadscaleControl's SetTags method with the given context and request.
	machine, err := task.HeadscaleControl.SetTags(context.Background(), &tags.SetTagsRequest)
	if err != nil {
		return nil, err
	}
	// Delete the cached machine information for all machines and the specific user.
	machineCache.Delete("")
	machineCache.Delete(machine.Machine.User.Name)
	return machine.Machine, err
}

// SetTagsWithStringSlice updates the tags of a machine with the given machineId and tags
// as a slice of strings. It returns the updated machine and an error if any.
func (h *headscaleRepository) SetTagsWithStringSlice(machineId uint64, tags []string) (*pb.Machine, error) {
	// Call the HeadscaleControl's SetTags method with the given context and request.
	machine, err := task.HeadscaleControl.SetTags(context.Background(), &pb.SetTagsRequest{
		MachineId: machineId,
		Tags:      tags,
	})
	if err != nil {
		return nil, err
	}
	// Delete the cached machine information for all machines and the specific user.
	machineCache.Delete("")
	machineCache.Delete(machine.Machine.User.Name)
	return machine.Machine, err
}

// SetTagsWithStrings updates the tags of a machine with the given machineId and tags
// as variadic strings. It returns the updated machine and an error if any.
func (h *headscaleRepository) SetTagsWithStrings(machineId uint64, tags ...string) (*pb.Machine, error) {
	// Call the HeadscaleControl's SetTags method with the given context and request.
	machine, err := task.HeadscaleControl.SetTags(context.Background(), &pb.SetTagsRequest{
		MachineId: machineId,
		Tags:      tags,
	})
	if err != nil {
		return nil, err
	}
	// Delete the cached machine information for all machines and the
	machineCache.Delete("")
	machineCache.Delete(machine.Machine.User.Name)
	return machine.Machine, err
}

// Machine end

//func (h *headscaleRepository) RebuildCache() error {
//	data, err := h.ListUser()
//	if err != nil {
//		return err
//	}
//
//	users, ok := data.([]*pb.User)
//	if !ok {
//		return errors.New("assertion failed")
//	}
//
//	oldData, found := usersCache.Get("users")
//	oldUsers := make([]*pb.User, 0)
//	if found {
//		if oldUsers, ok = oldData.([]*pb.User); ok {
//			return errors.New("assertion failed")
//		}
//	}
//
//	createUsers, deleteUsers, _ := checkSyncUser(users, oldUsers)
//
//	if len(deleteUsers) > 0 {
//		userNames := []string{}
//		for _, user := range deleteUsers {
//			userNames = append(userNames, user.Name)
//		}
//		if err := userRepo.BatchDeleteUserByNames(userNames); err != nil {
//			log.Log.Error(err)
//			// todo 发送到事务处理器
//		}
//	}
//
//	if len(createUsers) > 0 {
//		//if err := userRepo.CreateUsers(grpcUserToModelUser(createUsers)); err != nil {
//		//	log.Log.Error(err)
//		//	// todo 发送到事务管理器
//		//}
//		usersCache.Set("not used", createUsers, cache.DefaultExpiration)
//	}
//
//	usersCache.Set("users", users, cache.DefaultExpiration)
//	return nil
//}

// checkSyncUser compares the users from the namespace with the users from the storage
// and returns the users to be created, deleted, and updated in the storage.
func checkSyncUser(users, storageUser []*pb.User) ([]*pb.User, []*pb.User, []*pb.User) {
	var deleteUsers []*pb.User
	var createUsers []*pb.User
	var srcUsers []*pb.User
	lenNamespase := len(users)
	lenStorageUser := len(storageUser)

	news := lenNamespase - lenStorageUser
	if news <= 0 {
		news = 0
	}

	// i is used as a counter pointer for new users
	// j is used as a counter pointer for the original users

	var i = 0
	var j = 0

	lasttime := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	if lenStorageUser > 0 {
		lasttime = storageUser[lenStorageUser-1].CreatedAt.AsTime()
	out:
		for i < lenNamespase {
			for j < lenStorageUser {
				if users[i].Id != storageUser[j].Id {
					// Add to the deleted list.
					deleteUsers = append(deleteUsers, storageUser[j])
					j++
					break
				}

				if users[i].CreatedAt.AsTime().After(lasttime) {
					break out
				}

				srcUsers = append(srcUsers, users[i])
				j++
				i++
			}
		}

		for ; j < len(storageUser); j++ {
			// Add to the deleted list.
			deleteUsers = append(deleteUsers, storageUser[j])
		}
	}

	for ; i < len(users); i++ {
		// Add to the create list.
		createUsers = append(createUsers, users[i])
	}
	return createUsers, deleteUsers, srcUsers
}

//func grpcUserToModelUser(users []*pb.User) (data []model.User) {
//	for _, user := range users {
//		data = append(data, model.User{
//			ID:   user.Id,
//			Name: user.Name,
//			//CreateAt: user.CreatedAt,
//		})
//	}
//	return
//}

// searchUser searches for a user with the given id in the given list of users.
// It uses binary search algorithm for searching.
func searchUser(users []*pb.User, id string) *pb.User {
	min := 0
	max := len(users) - 1
	for min <= max {
		mid := min + (max-min)>>2
		if users[mid].Id == id {
			return users[mid]
		} else if users[mid].Id > id {
			max = mid - 1
		} else {
			min = mid + 1
		}
	}
	return nil
}
