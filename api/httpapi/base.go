package httpapi

import (
	"github.com/allentom/haruka"
	"github.com/jinzhu/copier"
	"github.com/mitchellh/mapstructure"
	"github.com/projectxpolaris/youcomic/api/httpapi/serializer"
	"github.com/projectxpolaris/youcomic/auth"
	ApiError "github.com/projectxpolaris/youcomic/error"
	"github.com/projectxpolaris/youcomic/model"
	"github.com/projectxpolaris/youcomic/permission"
	"github.com/projectxpolaris/youcomic/services"
	"github.com/projectxpolaris/youcomic/utils"
	"github.com/projectxpolaris/youcomic/validate"
	"github.com/sirupsen/logrus"
	"net/http"
)

// decode json body(with error abort)
//
// response body => interface
func DecodeJsonBody(context *haruka.Context, requestBody interface{}) error {
	err := context.ParseJson(&requestBody)
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.JsonParseError, nil)
	}
	return err
}

func RenderTemplate(context *haruka.Context, template serializer.TemplateSerializer, model interface{}) {
	err := serializer.DefaultSerializerModelByTemplate(model, template)
	if err != nil {
		logrus.Error(err)
		ApiError.RaiseApiError(context, err, nil)
	}
}

func AssignUpdateModel(requestBody interface{}, model interface{}) error {
	return copier.Copy(model, requestBody)
}

func AssignRequestBodyToModel(requestBody interface{}, model interface{}) error {
	return copier.Copy(model, requestBody)
}
func GetLookUpId(ctx *haruka.Context, lookup string) (int, error) {
	id, err := ctx.GetPathParameterAsInt(lookup)
	return id, err
}

type PageReader interface {
	Read(ctx *haruka.Context) (int, int)
}
type DefaultPagination struct {
	Pagination
}

func (r *DefaultPagination) Read(ctx *haruka.Context) (int, int) {
	var err error
	r.Page, err = ctx.GetQueryInt("page")
	if err != nil {
		r.Page = 1
	}
	r.PageSize, err = ctx.GetQueryInt("page_size")
	if err != nil {
		r.PageSize = 20
	}
	return r.Page, r.PageSize
}

type Pagination struct {
	Page     int
	PageSize int
}

type FilterMapping struct {
	Lookup string
	Method string
	Many   bool
}

type SuccessResponse struct {
	Success bool `json:"success"`
}

func ServerSuccessResponse(ctx *haruka.Context) {
	ctx.JSONWithStatus(SuccessResponse{Success: true}, http.StatusOK)
}

type ModelsBatchView struct {
	Context          *haruka.Context
	CreateModel      func() interface{}
	AllowUpdateField []string
	AllowOperations  []BatchOperation
	OperationFunc    map[BatchOperation]func(v *ModelsBatchView) error
	RequestBody      ModelBatchRequestBody
	Permissions      map[BatchOperation]func(v *ModelsBatchView) []permission.PermissionChecker
	Validators       map[BatchOperation]func(v *ModelsBatchView) []validate.Validator
	Claims           auth.JwtClaims
}

type BatchOperation string

const (
	Create BatchOperation = "Create"
	Update BatchOperation = "Update"
	Delete BatchOperation = "Delete"
)

type ModelBatchRequestBody struct {
	Create []interface{} `json:"create"`
	Update []interface{} `json:"update"`
	Delete []int         `json:"delete"`
}

var DefaultBatchFunctionMap = map[BatchOperation]func(v *ModelsBatchView) error{
	Create: func(v *ModelsBatchView) error {
		var err error
		var models []interface{}
		for _, rawRequestModelToCreate := range v.RequestBody.Create {
			md := v.CreateModel()
			err = mapstructure.Decode(rawRequestModelToCreate, md)
			if err != nil {
				return err
			}
			models = append(models, md)
		}
		err = services.CreateModels(models)
		if err != nil {
			return err
		}
		return nil
	},
	Update: func(v *ModelsBatchView) error {
		var err error
		modelToUpdate := v.CreateModel()
		err = services.UpdateModels(modelToUpdate, v.RequestBody.Update, v.AllowUpdateField...)
		if err != nil {
			return err
		}
		return nil
	},
	Delete: func(v *ModelsBatchView) error {
		var err error
		err = services.DeleteModels(v.CreateModel(), v.RequestBody.Delete...)
		if err != nil {
			ApiError.RaiseApiError(v.Context, err, nil)
			return err
		}
		return nil
	},
}

func (v *ModelsBatchView) Run() {
	var err error
	if claims, ok := v.Context.Param["claim"]; ok {
		v.Claims = claims.(*model.User)
	}
	var requestBody ModelBatchRequestBody
	err = DecodeJsonBody(v.Context, &requestBody)
	if err != nil {
		return
	}
	v.RequestBody = requestBody
	for _, operationKey := range v.AllowOperations {
		//check permission
		if v.Permissions != nil {
			if permissionCheckersFunc, isExist := v.Permissions[operationKey]; isExist {
				if hasPermission := permission.CheckPermissionAndServerError(v.Context, permissionCheckersFunc(v)...); !hasPermission {
					return
				}
			}
		}

		//validate
		if v.Validators != nil {
			if validatorFunc, isExist := v.Validators[operationKey]; isExist {
				if isValidate := validate.RunValidatorsAndRaiseApiError(v.Context, validatorFunc(v)...); !isValidate {
					return
				}
			}
		}
		operation, exist := v.OperationFunc[operationKey]
		if exist {
			err = operation(v)
			if err != nil {
				ApiError.RaiseApiError(v.Context, err, nil)
				return
			}
		} else {
			operation, exist := DefaultBatchFunctionMap[operationKey]
			if exist {
				err = operation(v)
				if err != nil {
					ApiError.RaiseApiError(v.Context, err, nil)
					return
				}
			}
		}
	}
	ServerSuccessResponse(v.Context)
}

//create model view
type CreateModelView struct {
	Context          *haruka.Context
	onAuthUser       func(v *CreateModelView) error
	CreateModel      func() interface{}
	ResponseTemplate serializer.TemplateSerializer
	RequestBody      interface{}
	Claims           auth.JwtClaims
	OnBeforeCreate   func(v *CreateModelView, modelToCreate interface{})
	GetPermissions   func(v *CreateModelView) []permission.PermissionChecker
	GetValidators    func(v *CreateModelView) []validate.Validator
}

func (v *CreateModelView) Run() {
	var err error
	err = DecodeJsonBody(v.Context, v.RequestBody)
	if err != nil {
		return
	}
	if claims, ok := v.Context.Param["claim"]; ok {
		v.Claims = claims.(*model.User)
	}

	//check permission
	if v.GetPermissions != nil {
		permissions := v.GetPermissions(v)
		if hasPermission := permission.CheckPermissionAndServerError(v.Context, permissions...); !hasPermission {
			return
		}
	}

	//validate check
	if v.GetValidators != nil {
		validators := v.GetValidators(v)
		if isValidate := validate.RunValidatorsAndRaiseApiError(v.Context, validators...); !isValidate {
			return
		}
	}

	createModel := v.CreateModel()
	err = copier.Copy(createModel, v.RequestBody)
	if err != nil {
		ApiError.RaiseApiError(v.Context, err, nil)
		return
	}
	if v.OnBeforeCreate != nil {
		v.OnBeforeCreate(v, createModel)
	}
	err = services.CreateModel(createModel)
	if err != nil {
		ApiError.RaiseApiError(v.Context, err, nil)
		return
	}

	//serializer response
	RenderTemplate(v.Context, v.ResponseTemplate, createModel)
	v.Context.JSONWithStatus(v.ResponseTemplate, http.StatusCreated)
}

type RequestBodyReader interface {
	Deserializer(source interface{}) error
}

type DefaultRequestBodyReader struct {
}

func (r *DefaultRequestBodyReader) Deserializer(source interface{}) error {
	return copier.Copy(r, source)
}

type ListView struct {
	Context              *haruka.Context
	Pagination           PageReader
	QueryBuilder         interface{}
	FilterMapping        []FilterMapping
	GetSerializerContext func(v *ListView, result interface{}) map[string]interface{}
	GetTemplate          func() serializer.TemplateSerializer
	GetContainer         func() serializer.ListContainerSerializer
	GetPermissions       func(v *ListView) []permission.PermissionChecker
	OnApplyQuery         func()
	Claims               auth.JwtClaims
}

func (v *ListView) Run() {

	if claims, ok := v.Context.Param["claim"]; ok {
		v.Claims = claims.(*model.User)
	}

	//check permission
	if v.GetPermissions != nil {
		permissions := v.GetPermissions(v)
		if hasPermission := permission.CheckPermissionAndServerError(v.Context, permissions...); !hasPermission {
			return
		}
	}

	if v.Pagination == nil {
		v.Pagination = &DefaultPagination{}
	}
	page, PageSize := v.Pagination.Read(v.Context)
	//get filter
	//allowFilterParam := []string{"id",""}
	pageFilter := (v.QueryBuilder).(services.PageFilter)
	pageFilter.SetPageFilter(page, PageSize)

	for _, filter := range v.FilterMapping {
		utils.FilterByParam(v.Context, filter.Lookup, v.QueryBuilder, filter.Method, filter.Many)
	}
	if v.OnApplyQuery != nil {
		v.OnApplyQuery()
	}
	modelsReader := (v.QueryBuilder).(services.ModelsReader)
	count, modelList, err := modelsReader.ReadModels()
	if err != nil {
		ApiError.RaiseApiError(v.Context, err, nil)
		return
	}
	serializerContext := map[string]interface{}{}
	if v.GetSerializerContext != nil {
		serializerContext = v.GetSerializerContext(v, modelList)
	}

	result := serializer.SerializeMultipleTemplate(modelList, v.GetTemplate(), serializerContext)
	responseBody := v.GetContainer()
	responseBody.SerializeList(result, map[string]interface{}{
		"page":     page,
		"pageSize": PageSize,
		"count":    count,
		"url":      v.Context.Request.URL,
	})
	v.Context.JSONWithStatus(responseBody, http.StatusOK)
}

type ModelView struct {
	Context     *haruka.Context
	GetModels   func() interface{}
	GetTemplate func() serializer.TemplateSerializer
	LookUpKey   string
	SetFilter   func(v *ModelView, lookupValue interface{})
}

func (v *ModelView) Run() {

}
