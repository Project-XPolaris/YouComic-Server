package httpapi

import (
	"net/http"

	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youcomic/model"
	"github.com/projectxpolaris/youcomic/services"
	"github.com/sirupsen/logrus"
)

// TemplateListHandler 获取模板列表
var TemplateListHandler haruka.RequestHandler = func(context *haruka.Context) {
	// 获取查询参数
	templateType := context.GetQueryString("type")
	page, err := context.GetQueryInt("page")
	if err != nil {
		page = 1
	}
	pageSize, err := context.GetQueryInt("pageSize")
	if err != nil {
		pageSize = 20
	}

	// 构建查询
	query := services.GetDB().Model(&model.Template{})

	if templateType != "" {
		query = query.Where("type = ?", templateType)
	}

	// 计算总数
	var total int64
	err = query.Count(&total).Error
	if err != nil {
		logrus.WithError(err).Error("计算模板总数失败")
		context.JSONWithStatus(haruka.JSON{
			"success": false,
			"message": "获取模板列表失败",
		}, http.StatusInternalServerError)
		return
	}

	// 分页查询
	var templates []model.Template
	offset := (page - 1) * pageSize
	err = query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&templates).Error
	if err != nil {
		logrus.WithError(err).Error("获取模板列表失败")
		context.JSONWithStatus(haruka.JSON{
			"success": false,
			"message": "获取模板列表失败",
		}, http.StatusInternalServerError)
		return
	}

	context.JSONWithStatus(haruka.JSON{
		"success": true,
		"data": haruka.JSON{
			"templates": templates,
			"total":     total,
			"page":      page,
			"pageSize":  pageSize,
		},
	}, http.StatusOK)
}

// TemplateDetailHandler 获取模板详情
var TemplateDetailHandler haruka.RequestHandler = func(context *haruka.Context) {
	id, err := GetLookUpId(context, "id")
	if err != nil {
		context.JSONWithStatus(haruka.JSON{
			"success": false,
			"message": "无效的模板ID",
		}, http.StatusBadRequest)
		return
	}

	template, err := services.GetTemplateByID(uint(id))
	if err != nil {
		logrus.WithError(err).Error("获取模板详情失败")
		context.JSONWithStatus(haruka.JSON{
			"success": false,
			"message": "模板不存在",
		}, http.StatusNotFound)
		return
	}

	context.JSONWithStatus(haruka.JSON{
		"success": true,
		"data":    template,
	}, http.StatusOK)
}

// TemplateCreateHandler 创建模板
var TemplateCreateHandler haruka.RequestHandler = func(context *haruka.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Type        string `json:"type" binding:"required"`
		Content     string `json:"content" binding:"required"`
		Description string `json:"description"`
		Version     string `json:"version"`
	}

	err := DecodeJsonBody(context, &req)
	if err != nil {
		context.JSONWithStatus(haruka.JSON{
			"success": false,
			"message": "请求参数错误",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}

	template := &model.Template{
		Name:        req.Name,
		Type:        req.Type,
		Content:     req.Content,
		Description: req.Description,
		Version:     req.Version,
		IsDefault:   false, // 用户创建的模板默认不是系统默认模板
	}

	err = services.CreateTemplate(template)
	if err != nil {
		logrus.WithError(err).Error("创建模板失败")
		context.JSONWithStatus(haruka.JSON{
			"success": false,
			"message": "创建模板失败",
		}, http.StatusInternalServerError)
		return
	}

	context.JSONWithStatus(haruka.JSON{
		"success": true,
		"data":    template,
		"message": "模板创建成功",
	}, http.StatusOK)
}

// TemplateUpdateHandler 更新模板
var TemplateUpdateHandler haruka.RequestHandler = func(context *haruka.Context) {
	id, err := GetLookUpId(context, "id")
	if err != nil {
		context.JSONWithStatus(haruka.JSON{
			"success": false,
			"message": "无效的模板ID",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		Name        string `json:"name"`
		Type        string `json:"type"`
		Content     string `json:"content"`
		Description string `json:"description"`
		Version     string `json:"version"`
	}

	err = DecodeJsonBody(context, &req)
	if err != nil {
		context.JSONWithStatus(haruka.JSON{
			"success": false,
			"message": "请求参数错误",
			"error":   err.Error(),
		}, http.StatusBadRequest)
		return
	}

	// 获取现有模板
	template, err := services.GetTemplateByID(uint(id))
	if err != nil {
		context.JSONWithStatus(haruka.JSON{
			"success": false,
			"message": "模板不存在",
		}, http.StatusNotFound)
		return
	}

	// 更新字段
	if req.Name != "" {
		template.Name = req.Name
	}
	if req.Type != "" {
		template.Type = req.Type
	}
	if req.Content != "" {
		template.Content = req.Content
	}
	if req.Description != "" {
		template.Description = req.Description
	}
	if req.Version != "" {
		template.Version = req.Version
	}

	err = services.UpdateTemplate(template)
	if err != nil {
		logrus.WithError(err).Error("更新模板失败")
		context.JSONWithStatus(haruka.JSON{
			"success": false,
			"message": "更新模板失败",
		}, http.StatusInternalServerError)
		return
	}

	context.JSONWithStatus(haruka.JSON{
		"success": true,
		"data":    template,
		"message": "模板更新成功",
	}, http.StatusOK)
}

// TemplateDeleteHandler 删除模板
var TemplateDeleteHandler haruka.RequestHandler = func(context *haruka.Context) {
	id, err := GetLookUpId(context, "id")
	if err != nil {
		context.JSONWithStatus(haruka.JSON{
			"success": false,
			"message": "无效的模板ID",
		}, http.StatusBadRequest)
		return
	}

	// 检查是否为默认模板
	template, err := services.GetTemplateByID(uint(id))
	if err != nil {
		context.JSONWithStatus(haruka.JSON{
			"success": false,
			"message": "模板不存在",
		}, http.StatusNotFound)
		return
	}

	if template.IsDefault {
		context.JSONWithStatus(haruka.JSON{
			"success": false,
			"message": "不能删除系统默认模板",
		}, http.StatusBadRequest)
		return
	}

	err = services.DeleteTemplate(uint(id))
	if err != nil {
		logrus.WithError(err).Error("删除模板失败")
		context.JSONWithStatus(haruka.JSON{
			"success": false,
			"message": "删除模板失败",
		}, http.StatusInternalServerError)
		return
	}

	context.JSONWithStatus(haruka.JSON{
		"success": true,
		"message": "模板删除成功",
	}, http.StatusOK)
}

// TemplateTypesHandler 获取模板类型列表
var TemplateTypesHandler haruka.RequestHandler = func(context *haruka.Context) {
	types := []haruka.JSON{
		{
			"value":       services.TemplateTypeTagPrompt,
			"label":       "标签提示模板",
			"description": "用于单个文本的标签分析",
		},
		{
			"value":       services.TemplateTypeBatchTagPrompt,
			"label":       "批量标签提示模板",
			"description": "用于多个文本的批量标签分析",
		},
	}

	context.JSONWithStatus(haruka.JSON{
		"success": true,
		"data":    types,
	}, http.StatusOK)
}
