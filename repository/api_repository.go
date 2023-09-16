package repository

import (
	"errors"
	"fmt"
	"github.com/thoas/go-funk"
	"headscale-panel/common"
	"headscale-panel/dto"
	"headscale-panel/model"
	"headscale-panel/vo"
	"strings"
)

type IApiRepository interface {
	GetApis(req *vo.ApiListRequest) ([]*model.Api, int64, error) // Get API list
	GetApisById(apiIds []uint) ([]*model.Api, error)             // Get API list by API ID
	GetApiTree() ([]*dto.ApiTreeDto, error)                      // Get API tree (classified by API Category field)
	CreateApi(api *model.Api) error                              // Create API
	UpdateApiById(apiId uint, api *model.Api) error              // Update API
	BatchDeleteApiByIds(apiIds []uint) error                     // Batch delete APIs
	GetApiDescByPath(path string, method string) (string, error) // Get API description based on API path and request method
}

type ApiRepository struct{}

func NewApiRepository() IApiRepository {
	return ApiRepository{}
}

// Get API list
func (a ApiRepository) GetApis(req *vo.ApiListRequest) ([]*model.Api, int64, error) {
	var list []*model.Api
	db := common.DB.Model(&model.Api{}).Order("created_at DESC")

	method := strings.TrimSpace(req.Method)
	if method != "" {
		db = db.Where("method LIKE ?", fmt.Sprintf("%%%s%%", method))
	}
	path := strings.TrimSpace(req.Path)
	if path != "" {
		db = db.Where("path LIKE ?", fmt.Sprintf("%%%s%%", path))
	}
	category := strings.TrimSpace(req.Category)
	if category != "" {
		db = db.Where("category LIKE ?", fmt.Sprintf("%%%s%%", category))
	}
	creator := strings.TrimSpace(req.Creator)
	if creator != "" {
		db = db.Where("creator LIKE ?", fmt.Sprintf("%%%s%%", creator))
	}

	// Only paginate when pageNum > 0 and pageSize > 0
	// Record total number
	var total int64
	err := db.Count(&total).Error
	if err != nil {
		return list, total, err
	}
	pageNum := int(req.PageNum)
	pageSize := int(req.PageSize)
	if pageNum > 0 && pageSize > 0 {
		err = db.Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&list).Error
	} else {
		err = db.Find(&list).Error
	}
	return list, total, err
}

// Get API list by API ID
func (a ApiRepository) GetApisById(apiIds []uint) ([]*model.Api, error) {
	var apis []*model.Api
	err := common.DB.Where("id IN (?)", apiIds).Find(&apis).Error
	return apis, err
}

// Get API tree (classified by API Category field)
func (a ApiRepository) GetApiTree() ([]*dto.ApiTreeDto, error) {
	var apiList []*model.Api
	err := common.DB.Order("category").Order("created_at").Find(&apiList).Error
	// Get all categories
	var categoryList []string
	for _, api := range apiList {
		categoryList = append(categoryList, api.Category)
	}
	// Get deduplicated categories
	categoryUniq := funk.UniqString(categoryList)

	apiTree := make([]*dto.ApiTreeDto, len(categoryUniq))

	for i, category := range categoryUniq {
		apiTree[i] = &dto.ApiTreeDto{
			ID:       -i,
			Desc:     category,
			Category: category,
			Children: nil,
		}
		for _, api := range apiList {
			if category == api.Category {
				apiTree[i].Children = append(apiTree[i].Children, api)
			}
		}
	}

	return apiTree, err
}

// Create API
func (a ApiRepository) CreateApi(api *model.Api) error {
	err := common.DB.Create(api).Error
	return err
}

// Update API
func (a ApiRepository) UpdateApiById(apiId uint, api *model.Api) error {
	// Get API information by ID
	var oldApi model.Api
	err := common.DB.First(&oldApi, apiId).Error
	if err != nil {
		return errors.New("failed to get API information by API ID")
	}
	err = common.DB.Model(api).Where("id = ?", apiId).Updates(api).Error
	if err != nil {
		return err
	}
	// Update the policy in the casbin once the method and path have been updated
	if oldApi.Path != api.Path || oldApi.Method != api.Method {
		policies := common.CasbinEnforcer.GetFilteredPolicy(1, oldApi.Path, oldApi.Method)
		// The interface only operates if it exists in the casbin's policy
		if len(policies) > 0 {
			// Delete
			isRemoved, _ := common.CasbinEnforcer.RemovePolicies(policies)
			if !isRemoved {
				return errors.New("update permission API failed")
			}
			for _, policy := range policies {
				policy[1] = api.Path
				policy[2] = api.Method
			}
			// Add
			isAdded, _ := common.CasbinEnforcer.AddPolicies(policies)
			if !isAdded {
				return errors.New("update permission API failed")
			}
			// Load policy
			err := common.CasbinEnforcer.LoadPolicy()
			if err != nil {
				return errors.New("update permission API succeeded, permission API strategy loading failed")
			} else {
				return err
			}
		}
	}
	return err
}

// Batch delete API
func (a ApiRepository) BatchDeleteApiByIds(apiIds []uint) error {

	apis, err := a.GetApisById(apiIds)
	if err != nil {
		return errors.New("failed to get API list by interface ID")
	}
	if len(apis) == 0 {
		return errors.New("API list not obtained by interface ID")
	}

	err = common.DB.Where("id IN (?)", apiIds).Unscoped().Delete(&model.Api{}).Error
	// If delete success, then delete policy in Casbin
	if err == nil {
		for _, api := range apis {
			policies := common.CasbinEnforcer.GetFilteredPolicy(1, api.Path, api.Method)
			if len(policies) > 0 {
				isRemoved, _ := common.CasbinEnforcer.RemovePolicies(policies)
				if !isRemoved {
					return errors.New("delete permission API failed")
				}
			}
		}
		// Reload policy
		err := common.CasbinEnforcer.LoadPolicy()
		if err != nil {
			return errors.New("delete permission API succeeded, permission API strategy loading failed")
		} else {
			return err
		}
	}
	return err
}

// Get API description based on API path and request method
func (a ApiRepository) GetApiDescByPath(path string, method string) (string, error) {
	var api model.Api
	err := common.DB.Where("path = ?", path).Where("method = ?", method).First(&api).Error
	return api.Desc, err
}
