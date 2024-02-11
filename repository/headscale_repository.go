package repository

import (
	"context"
	"fmt"
	pb "github.com/juanfont/headscale/gen/go/headscale/v1"
	"github.com/patrickmn/go-cache"
	"google.golang.org/protobuf/types/known/timestamppb"
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
var NodeCache = cache.New(time.Minute, time.Minute)
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
	GetNodeRoutes(request *vo.GetNodeRoutesRequest) ([]*pb.Route, error)
	GetNodeRoutesWithId(NodeId uint64) ([]*pb.Route, error)
	DeleteRoute(request *vo.DeleteRouteRequest) error
	DeleteRouteWithId(routeId uint64) error
	SwitchRoute(request *vo.SwitchRouteRequest) error
	SwitchRouteWithId(routeId uint64, enable bool) error
}

// HeadscaleNodesRepository is an interface for managing Node information.
type HeadscaleNodesRepository interface {
	ListNodes(user *vo.ListNodesRequest) ([]*pb.Node, error)
	ListNodesWithUser(user string) ([]*pb.Node, error)
	GetNode(getNode *vo.GetNodeRequest) (*pb.Node, error)
	GetNodeWithId(NodeId uint64) (*pb.Node, error)
	ExpireNode(expireNode *vo.ExpireNodeRequest) (*pb.Node, error)
	ExpireNodeWithId(NodeId uint64) (*pb.Node, error)
	RenameNode(NodeId uint64, name string) (*pb.Node, error)
	RenameNodeWithNewName(NodeId uint64, name string) (*pb.Node, error)
	MoveNode(Node *vo.MoveNodeRequest) (*pb.Node, error)
	MoveNodeWithUser(NodeId uint64, user string) (*pb.Node, error)
	DeleteNode(Node *vo.DeleteNodeRequest) error
	DeleteNodeWithId(NodeId uint64) error
	RegisterNode(newNode *vo.RegisterNode) (*pb.Node, error)
	RegisterNodeWithKey(user, key string) (*pb.Node, error)
	SetTags(tags *vo.SetTagsRequest) (*pb.Node, error)
	SetTagsWithStringSlice(NodeId uint64, tags []string) (*pb.Node, error)
	SetTagsWithStrings(NodeId uint64, tags ...string) (*pb.Node, error)
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

// NewNodesRepo returns a new instance of HeadscaleNodesRepository.
func NewNodesRepo() HeadscaleNodesRepository {
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

// GetNodeRoutes This method takes a GetNodeRoutesRequest
// and returns a slice of pb.Route and an error. If the routes
// for the given nodeId are present in the routeCache, it returns them.
// Otherwise, it calls task.HeadscaleControl.GetNodeRoutes to get the routes.
// If the call is successful, it adds the routes to the routeCache with an expiration time
// of one minute and returns the routes and any error that occurred.
func (h *headscaleRepository) GetNodeRoutes(request *vo.GetNodeRoutesRequest) ([]*pb.Route, error) {
	if routes, ok := routeCache.Get(strconv.FormatUint(request.NodeId, 10)); ok {
		return routes.([]*pb.Route), nil
	}
	routes, err := task.HeadscaleControl.GetNodeRoutes(context.Background(), &request.GetNodeRoutesRequest)
	if err != nil {
		return nil, err
	}
	if len(routes.Routes) > 0 && routes.Routes[0].Node != nil {
		if err = routeCache.Add(strconv.FormatUint(routes.Routes[0].Node.Id, 10), routes.Routes, time.Minute); err != nil {
			log.Log.Errorf("add routes by nodeid to cache error:%v", err)
		}
	}
	return routes.Routes, err
}

// GetNodeRoutesWithId This method takes a nodeId and returns a slice of pb.Route and an error.
// If the routes for the given nodeId are present in the routeCache, it returns them.
// Otherwise, it calls task.HeadscaleControl.GetNodeRoutes to get the routes. If the call is successful,
// it adds the routes to the routeCache with an expiration time of one minute and returns the routes and any error that occurred.
func (h *headscaleRepository) GetNodeRoutesWithId(NodeId uint64) ([]*pb.Route, error) {
	if routes, ok := routeCache.Get(strconv.FormatUint(NodeId, 10)); ok {
		return routes.([]*pb.Route), nil
	}
	routes, err := task.HeadscaleControl.GetNodeRoutes(context.Background(), &pb.GetNodeRoutesRequest{NodeId: NodeId})
	if err != nil {
		return nil, err
	}
	if len(routes.Routes) > 0 && routes.Routes[0].Node != nil {
		if err = routeCache.Add(strconv.FormatUint(routes.Routes[0].Node.Id, 10), routes.Routes, time.Minute); err != nil {
			log.Log.Errorf("add routes by nodeid to cache error:%v", err)
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

// Node start

// ListNodes retrieves a list of nodes for a given user, either from cache or by calling the HeadscaleControl API.
// If the nodes are retrieved from cache, they are returned immediately.
// Otherwise, the HeadscaleControl API is called with the user's ListNodesRequest, and the resulting nodes are added to cache before being returned.
// If there is an error retrieving the nodes, an error is returned.
func (h *headscaleRepository) ListNodes(user *vo.ListNodesRequest) ([]*pb.Node, error) {
	if Nodes, ok := NodeCache.Get(user.User); ok {
		return Nodes.([]*pb.Node), nil
	}
	Nodes, err := task.HeadscaleControl.ListNodes(context.Background(), &user.ListNodesRequest)
	if err != nil {
		return nil, err
	}
	if err = NodeCache.Add(user.User, Nodes.Nodes, cache.DefaultExpiration); err != nil {
		log.Log.Errorf("add Nodes to cache error:%v", err)
	}
	return Nodes.Nodes, nil
}

// ListNodesWithUser retrieves a list of nodes for a given user, either from cache or by calling the HeadscaleControl API.
// If the Nodes are retrieved from cache, they are returned immediately.
// Otherwise, the HeadscaleControl API is called with a ListNodesRequest containing the user's name, and the resulting Nodes are added to cache before being returned.
// If there is an error retrieving the nodes, an error is returned.
func (h *headscaleRepository) ListNodesWithUser(user string) ([]*pb.Node, error) {
	if Nodes, ok := NodeCache.Get(user); ok {
		return Nodes.([]*pb.Node), nil
	}
	Nodes, err := task.HeadscaleControl.ListNodes(context.Background(), &pb.ListNodesRequest{User: user})
	if err != nil {
		return nil, err
	}
	if err = NodeCache.Add(user, Nodes.Nodes, cache.DefaultExpiration); err != nil {
		log.Log.Errorf("add Nodes to cache error:%v", err)
	}
	return Nodes.Nodes, nil
}

// GetNode retrieves a node with a given ID, either from cache or by calling the HeadscaleControl API.
// If the node is retrieved from cache, it is returned immediately.
// Otherwise, the HeadscaleControl API is called with the GetNodeRequest, and the resulting node is added to cache before being returned.
// If there is an error retrieving the node, an error is returned.
func (h *headscaleRepository) GetNode(getNode *vo.GetNodeRequest) (*pb.Node, error) {
	if Node, ok := NodeCache.Get(strconv.FormatUint(getNode.NodeId, 10)); ok {
		return Node.(*pb.Node), nil
	}
	Nodes, err := task.HeadscaleControl.GetNode(context.Background(), &getNode.GetNodeRequest)
	if err != nil {
		return nil, err
	}
	if err = NodeCache.Add(strconv.FormatUint(Nodes.Node.Id, 10), Nodes.Node, cache.DefaultExpiration); err != nil {
		log.Log.Errorf("add Node to cache error:%v", err)
	}
	return Nodes.Node, err
}

// GetNodeWithId retrieves a node with a given ID, either from cache or by calling the HeadscaleControl API.
// If the node is retrieved from cache, it is returned immediately.
// Otherwise, the HeadscaleControl API is called with a GetNodeRequest containing the node ID, and the resulting Node is added to cache before being returned.
// If there is an error retrieving the node, an error is returned.
func (h *headscaleRepository) GetNodeWithId(NodeId uint64) (*pb.Node, error) {
	if Node, ok := NodeCache.Get(strconv.FormatUint(NodeId, 10)); ok {
		return Node.(*pb.Node), nil
	}
	Node, err := task.HeadscaleControl.GetNode(context.Background(), &pb.GetNodeRequest{NodeId: NodeId})
	if err != nil {
		return nil, err
	}
	if err = NodeCache.Add(strconv.FormatUint(Node.Node.Id, 10), Node.Node, cache.DefaultExpiration); err != nil {
		log.Log.Errorf("add Node to cache error:%v", err)
	}
	return Node.Node, err
}

// ExpireNode sets a node's expiration date to the current time, effectively expiring the node.
// The HeadscaleControl API is called with the ExpireNodeRequest, and the resulting node is returned.
// If there is an error expiring the node, an error is returned.
func (h *headscaleRepository) ExpireNode(expireNode *vo.ExpireNodeRequest) (*pb.Node, error) {
	Node, err := task.HeadscaleControl.ExpireNode(context.Background(), &expireNode.ExpireNodeRequest)
	if err != nil {
		return nil, err
	}
	NodeCache.Delete("")
	NodeCache.Delete(Node.Node.User.Name)
	if err = NodeCache.Add(strconv.FormatUint(Node.Node.Id, 10), Node.Node, cache.DefaultExpiration); err != nil {
		NodeCache.Delete(strconv.FormatUint(Node.Node.Id, 10))
	}
	return Node.Node, err
}

// ExpireNodeWithId sets a node's expiration date to the current time, effectively expiring the node.
// The HeadscaleControl API is called with an ExpireNodeRequest containing the node ID, and the resulting node is returned.
// If there is an error expiring the node, an error is returned.
func (h *headscaleRepository) ExpireNodeWithId(NodeId uint64) (*pb.Node, error) {
	Node, err := task.HeadscaleControl.ExpireNode(context.Background(), &pb.ExpireNodeRequest{NodeId: NodeId})
	if err != nil {
		return nil, err
	}
	NodeCache.Delete("")
	NodeCache.Delete(Node.Node.User.Name)
	if err = NodeCache.Add(strconv.FormatUint(Node.Node.Id, 10), Node.Node, cache.DefaultExpiration); err != nil {
		NodeCache.Delete(strconv.FormatUint(Node.Node.Id, 10))
	}
	return Node.Node, err
}

// RenameNode renames a node with a given ID to a new name.
// The HeadscaleControl API is called with a RenameNodeRequest containing the node ID and new name, and the resulting Node is returned.
// If there is an error renaming the node, an error is returned.
func (h *headscaleRepository) RenameNode(NodeId uint64, name string) (*pb.Node, error) {
	Node, err := task.HeadscaleControl.RenameNode(context.Background(), &pb.RenameNodeRequest{
		NodeId:  NodeId,
		NewName: name,
	})
	if err != nil {
		return nil, err
	}
	NodeCache.Delete("")
	NodeCache.Delete(Node.Node.User.Name)
	if err = NodeCache.Add(strconv.FormatUint(Node.Node.Id, 10), Node.Node, cache.DefaultExpiration); err != nil {
		NodeCache.Delete(strconv.FormatUint(Node.Node.Id, 10))
	}
	return Node.Node, nil
}

// RenameNodeWithNewName renames a node with a given ID to a new name.
// The HeadscaleControl API is called with a RenameNodeRequest containing the node ID and new name, and the resulting Node is returned.
// If there is an error renaming the node, an error is returned.
func (h *headscaleRepository) RenameNodeWithNewName(NodeId uint64, name string) (*pb.Node, error) {
	Node, err := task.HeadscaleControl.RenameNode(context.Background(), &pb.RenameNodeRequest{
		NodeId:  NodeId,
		NewName: name,
	})
	if err != nil {
		return nil, err
	}
	NodeCache.Delete("")
	NodeCache.Delete(Node.Node.User.Name)
	if err = NodeCache.Add(strconv.FormatUint(Node.Node.Id, 10), Node.Node, cache.DefaultExpiration); err != nil {
		NodeCache.Delete(strconv.FormatUint(Node.Node.Id, 10))
	}
	return Node.Node, nil
}

// MoveNode moves a node to a new user.
// The HeadscaleControl API is called with the MoveNodeRequest, and the resulting node is returned.
// If there is an error moving the node, an error is returned.
func (h *headscaleRepository) MoveNode(Node *vo.MoveNodeRequest) (*pb.Node, error) {
	var name string
	var node interface{}
	var ok bool
	if node, ok = NodeCache.Get(strconv.FormatUint(Node.NodeId, 10)); !ok {
		if nodes, err := h.ListNodesWithUser(""); err != nil {
			log.Log.Error(err)
		} else {
			node = searchNode(nodes, Node.NodeId)
		}
	}

	if node == nil {
		return nil, fmt.Errorf("not find node")
	}
	name = node.(*pb.Node).User.Name
	movedNode, err := task.HeadscaleControl.MoveNode(context.Background(), &Node.MoveNodeRequest)
	if err != nil {
		return nil, err
	}
	NodeCache.Flush()
	NodeCache.Delete("")
	// cannot delete old user's cache
	NodeCache.Delete(name)
	NodeCache.Delete(movedNode.Node.User.Name)
	return movedNode.Node, nil
}

// MoveNodeWithUser moves a node to a new user.
// The HeadscaleControl API is called with a MoveNodeRequest containing the node ID and new user, and the resulting Node is returned.
// If there is an error moving the node, an error is returned.
func (h *headscaleRepository) MoveNodeWithUser(NodeId uint64, user string) (*pb.Node, error) {
	var name string
	var node interface{}
	var ok bool
	if node, ok = NodeCache.Get(strconv.FormatUint(NodeId, 10)); !ok {
		if nodes, err := h.ListNodesWithUser(""); err != nil {
			log.Log.Error(err)
		} else {
			node = searchNode(nodes, NodeId)
		}
	}

	if node == nil {
		return nil, fmt.Errorf("not find node")
	}
	name = node.(*pb.Node).User.Name

	Node, err := task.HeadscaleControl.MoveNode(context.Background(), &pb.MoveNodeRequest{
		NodeId: NodeId,
		User:   user,
	})
	if err != nil {
		return nil, err
	}
	NodeCache.Flush()
	NodeCache.Delete("")
	// cannot delete old user's cache
	NodeCache.Delete(name)
	NodeCache.Delete(Node.Node.User.Name)
	return Node.Node, nil
}

// DeleteNode deletes a Node with a given ID.
// The HeadscaleControl API is called with the DeleteNodeRequest, and any cached Nodes are deleted.
// If there is an error deleting the Node, an error is returned.
func (h *headscaleRepository) DeleteNode(Node *vo.DeleteNodeRequest) (err error) {
	var name string
	var node interface{}
	var ok bool
	if node, ok = NodeCache.Get(strconv.FormatUint(Node.NodeId, 10)); !ok {
		if nodes, err := h.ListNodesWithUser(""); err != nil {
			log.Log.Error(err)
		} else {
			node = searchNode(nodes, Node.NodeId)
		}
	}

	if node == nil {
		return fmt.Errorf("not find node")
	}
	name = node.(*pb.Node).User.Name
	_, err = task.HeadscaleControl.DeleteNode(context.Background(), &Node.DeleteNodeRequest)
	NodeCache.Delete("")
	// cannot delete user's cache by ID
	NodeCache.Delete(name)
	NodeCache.Delete(strconv.FormatUint(Node.NodeId, 10))
	return
}

// DeleteNodeWithId deletes a node with a given ID.
// The HeadscaleControl API is called with a DeleteNodeRequest containing the node ID, and any cached nodes are deleted.
// If there is an error deleting the node, an error is returned.
func (h *headscaleRepository) DeleteNodeWithId(NodeId uint64) (err error) {
	var name string
	var node interface{}
	var ok bool
	if node, ok = NodeCache.Get(strconv.FormatUint(NodeId, 10)); !ok {
		if nodes, err := h.ListNodesWithUser(""); err != nil {
			log.Log.Error(err)
		} else {
			node = searchNode(nodes, NodeId)
		}
	}

	if node == nil {
		return fmt.Errorf("not find node")
	}
	name = node.(*pb.Node).User.Name
	_, err = task.HeadscaleControl.DeleteNode(context.Background(), &pb.DeleteNodeRequest{NodeId: NodeId})
	NodeCache.Delete("")
	// cannot delete user's cache by nodeId
	NodeCache.Delete(name)
	NodeCache.Delete(strconv.FormatUint(NodeId, 10))
	return
}

// RegisterNode registers a new node with the given details.
// The HeadscaleControl API is called with the RegisterNodeRequest, and any cached nodes are deleted.
// If there is an error registering the node, an error is returned.
func (h *headscaleRepository) RegisterNode(newNode *vo.RegisterNode) (*pb.Node, error) {
	Node, err := task.HeadscaleControl.RegisterNode(context.Background(), &newNode.RegisterNodeRequest)
	if err != nil {
		return nil, err
	}
	NodeCache.Delete("")
	NodeCache.Delete(Node.Node.User.Name)
	NodeCache.Add(strconv.FormatUint(Node.Node.Id, 10), Node.Node, cache.DefaultExpiration)
	return Node.Node, nil
}

// RegisterNodeWithKey registers a new node with the given user and key.
// The HeadscaleControl API is called with a RegisterNodeRequest containing the user and key, and any cached Nodes are deleted.
// If there is an error registering the node, an error is returned.
func (h *headscaleRepository) RegisterNodeWithKey(user, key string) (*pb.Node, error) {
	Node, err := task.HeadscaleControl.RegisterNode(context.Background(), &pb.RegisterNodeRequest{
		User: user,
		Key:  key,
	})
	if err != nil {
		return nil, err
	}
	NodeCache.Delete("")
	NodeCache.Delete(Node.Node.User.Name)
	return Node.Node, nil
}

// SetTags updates the tags of a node with the given SetTagsRequest.
// It returns the updated node and an error if any.
func (h *headscaleRepository) SetTags(tags *vo.SetTagsRequest) (*pb.Node, error) {
	// Call the HeadscaleControl's SetTags method with the given context and request.
	Node, err := task.HeadscaleControl.SetTags(context.Background(), &tags.SetTagsRequest)
	if err != nil {
		return nil, err
	}
	// Delete the cached node information for all nodes and the specific user.
	NodeCache.Delete("")
	NodeCache.Delete(Node.Node.User.Name)
	if err = NodeCache.Add(strconv.FormatUint(Node.Node.Id, 10), Node.Node, cache.DefaultExpiration); err != nil {
		NodeCache.Delete(strconv.FormatUint(Node.Node.Id, 10))
	}
	return Node.Node, err
}

// SetTagsWithStringSlice updates the tags of a node with the given nodeId and tags
// as a slice of strings. It returns the updated node and an error if any.
func (h *headscaleRepository) SetTagsWithStringSlice(NodeId uint64, tags []string) (*pb.Node, error) {
	// Call the HeadscaleControl's SetTags method with the given context and request.
	Node, err := task.HeadscaleControl.SetTags(context.Background(), &pb.SetTagsRequest{
		NodeId: NodeId,
		Tags:   tags,
	})
	if err != nil {
		return nil, err
	}
	// Delete the cached node information for all nodes and the specific user.
	NodeCache.Delete("")
	NodeCache.Delete(Node.Node.User.Name)
	if err = NodeCache.Add(strconv.FormatUint(Node.Node.Id, 10), Node.Node, cache.DefaultExpiration); err != nil {
		NodeCache.Delete(strconv.FormatUint(Node.Node.Id, 10))
	}
	return Node.Node, err
}

// SetTagsWithStrings updates the tags of a node with the given nodeId and tags
// as variadic strings. It returns the updated node and an error if any.
func (h *headscaleRepository) SetTagsWithStrings(NodeId uint64, tags ...string) (*pb.Node, error) {
	// Call the HeadscaleControl's SetTags method with the given context and request.
	Node, err := task.HeadscaleControl.SetTags(context.Background(), &pb.SetTagsRequest{
		NodeId: NodeId,
		Tags:   tags,
	})
	if err != nil {
		return nil, err
	}
	// Delete the cached node information for all nodes and the
	NodeCache.Delete("")
	NodeCache.Delete(Node.Node.User.Name)
	if err = NodeCache.Add(strconv.FormatUint(Node.Node.Id, 10), Node.Node, cache.DefaultExpiration); err != nil {
		NodeCache.Delete(strconv.FormatUint(Node.Node.Id, 10))
	}
	return Node.Node, err
}

// Node end

// GetDevice wait headscale
//func (h *headscaleRepository) GetDevice(Id uint64) (*pb.GetDeviceResponse, error) {
//	Device, err := task.HeadscaleControl.GetDevice(context.Background(), &pb.GetDeviceRoutesRequest{Id: strconv.FormatUint(Id, 10)})
//	if err != nil {
//		return nil, err
//	}
//	DeviceCache.Delete("")
//	DeviceCache.Delete(Device.Device.User)
//	return Device.Device, err
//}

//func (h *headscaleRepository) GetDeviceRoutes(Id uint64) (*pb.GetDeviceRoutesResponse, error) {
//	Routes, err := task.HeadscaleControl.GetDeviceRoutes(context.Background(), &pb.GetDeviceRoutesRequest{Id: strconv.FormatUint(Id, 10)})
//	if err != nil {
//		return nil, err
//	}
//	return Routes.Routes, err
//}

//func (h *headscaleRepository) EnableDeviceRoutes(Id uint64, Routes []string) (*pb.EnableDeviceRoutesResponse, error) {
//	routes, err := task.HeadscaleControl.EnableDeviceRoutes(context.Background(), &pb.EnableDeviceRoutesRequest{Id: strconv.FormatUint(Id, 10), Routes: Routes})
//	if err != nil {
//		return nil, err
//	}
//	return routes, err
//}

//func (h *headscaleRepository) DeleteDevice(Id uint64) error {
//	_, err := task.HeadscaleControl.DeleteDevice(context.Background(), &pb.DeleteDeviceRequest{Id: strconv.FormatUint(Id, 10)})
//	return err
//}

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

func searchNode(nodes []*pb.Node, id uint64) *pb.Node {
	min := 0
	max := len(nodes)

	for min <= max {
		mid := min + (max-min)>>2
		if nodes[mid].Id == id {
			return nodes[mid]
		} else if nodes[mid].Id > id {
			max = mid - 1
		} else {
			min = mid + 1
		}
	}
	return nil
}
