package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mou-he/graduation-design/common/code"
	"github.com/mou-he/graduation-design/controller"
	"github.com/mou-he/graduation-design/model"
	"github.com/mou-he/graduation-design/service/suggestion"
)

// -------------------------- 请求/响应结构体 --------------------------
type (
	// 所有结构体完全复用，无需修改
	PublishSuggestionRequest struct {
		Content string `json:"content" binding:"required,min=1,max=500"`
	}
	PublishSuggestionResponse struct {
		controller.Response
		Suggestion *model.Suggestion `json:"suggestion,omitempty"`
	}
	ReplySuggestionRequest struct {
		ParentID        uint64 `json:"parent_id" binding:"required"`
		ReplyToUserID   string `json:"reply_to_user_id,omitempty"`
		ReplyToUsername string `json:"reply_to_username,omitempty"`
		Content         string `json:"content" binding:"required,min=1,max=500"`
	}
	ReplySuggestionResponse struct {
		controller.Response
		Reply *model.Suggestion `json:"reply,omitempty"`
	}
	UpdateSuggestionRequest struct {
		ID      uint64 `json:"id" binding:"required"`
		Content string `json:"content" binding:"required,min=1,max=500"`
	}
	UpdateSuggestionResponse struct {
		controller.Response
	}
	DeleteSuggestionRequest struct {
		ID uint64 `json:"id" binding:"required"`
	}
	DeleteSuggestionResponse struct {
		controller.Response
	}
	GetRootSuggestionPageRequest struct {
		Page     int `json:"page" binding:"required,min=1"`
		PageSize int `json:"page_size" binding:"required,min=1,max=50"`
	}
	GetRootSuggestionPageResponse struct {
		controller.Response
		Total    int64                    `json:"total,omitempty"`
		Page     int                      `json:"page,omitempty"`
		PageSize int                      `json:"page_size,omitempty"`
		List     []model.RootSuggestionVO `json:"list,omitempty"`
	}
	GetReplyByRootIDPageRequest struct {
		RootID   uint64 `json:"root_id" binding:"required"`
		Page     int    `json:"page" binding:"required,min=1"`
		PageSize int    `json:"page_size" binding:"required,min=1,max=50"`
	}
	GetReplyByRootIDPageResponse struct {
		controller.Response
		Total    int64              `json:"total,omitempty"`
		Page     int                `json:"page,omitempty"`
		PageSize int                `json:"page_size,omitempty"`
		List     []model.Suggestion `json:"list,omitempty"`
	}
)

// PublishSuggestion 发布主留言（完全复用）
func PublishSuggestion(c *gin.Context) {
	req := new(PublishSuggestionRequest)
	res := new(PublishSuggestionResponse)
	userId := c.Request.Header.Get("x-user-id")
	username := c.Request.Header.Get("x-username")

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidParams))
		return
	}

	modelReq := &model.PublishRequest{Content: req.Content}
	sug, err := suggestion.Publish(c.Request.Context(), modelReq, userId, username)
	if err != nil {
		res.CodeOf(code.CodeServerBusy)
		res.StatusMsg = err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	res.Success()
	res.Suggestion = sug
	c.JSON(http.StatusOK, res)
}

// ReplySuggestion 回复留言（仅自动兼容LikeCount字段删除，无代码修改）
func ReplySuggestion(c *gin.Context) {
	req := new(ReplySuggestionRequest)
	res := new(ReplySuggestionResponse)
	userId := c.Request.Header.Get("x-user-id")
	username := c.Request.Header.Get("x-username")

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidParams))
		return
	}

	modelReq := &model.ReplyRequest{
		ParentID:        req.ParentID,
		ReplyToUserID:   req.ReplyToUserID,
		ReplyToUsername: req.ReplyToUsername,
		Content:         req.Content,
	}
	reply, err := suggestion.Reply(c.Request.Context(), modelReq, userId, username)
	if err != nil {
		res.CodeOf(code.CodeServerBusy)
		res.StatusMsg = err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	res.Success()
	res.Reply = reply
	c.JSON(http.StatusOK, res)
}

// UpdateSuggestion 更新留言（完全复用）
func UpdateSuggestion(c *gin.Context) {
	req := new(UpdateSuggestionRequest)
	res := new(UpdateSuggestionResponse)
	userId := c.Request.Header.Get("x-user-id")

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidParams))
		return
	}

	modelReq := &model.UpdateRequest{ID: req.ID, Content: req.Content}
	err := suggestion.Update(c.Request.Context(), modelReq, userId)
	if err != nil {
		if err.Error() == "无权限更新该内容" {
			res.CodeOf(code.CodeNoAuth)
		} else {
			res.CodeOf(code.CodeServerBusy)
		}
		res.StatusMsg = err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	res.Success()
	c.JSON(http.StatusOK, res)
}

// DeleteSuggestion 删除留言（完全复用）
func DeleteSuggestion(c *gin.Context) {
	req := new(DeleteSuggestionRequest)
	res := new(DeleteSuggestionResponse)
	userId := c.Request.Header.Get("x-user-id")

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidParams))
		return
	}

	err := suggestion.Delete(c.Request.Context(), req.ID, userId)
	if err != nil {
		if err.Error() == "无权限删除该内容" {
			res.CodeOf(code.CodeNoAuth)
		} else {
			res.CodeOf(code.CodeServerBusy)
		}
		res.StatusMsg = err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	res.Success()
	c.JSON(http.StatusOK, res)
}

// GetRootSuggestionPage 根留言分页查询（完全复用，自动返回ReplyCount）
func GetRootSuggestionPage(c *gin.Context) {
	req := new(GetRootSuggestionPageRequest)
	res := new(GetRootSuggestionPageResponse)

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidParams))
		return
	}

	total, list, err := suggestion.GetRootSuggestionPage(c.Request.Context(), req.Page, req.PageSize)
	if err != nil {
		res.CodeOf(code.CodeServerBusy)
		res.StatusMsg = err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	res.Success()
	res.Total = total
	res.Page = req.Page
	res.PageSize = req.PageSize
	res.List = list // list中的RootSuggestionVO已包含ReplyCount、无LikeCount
	c.JSON(http.StatusOK, res)
}

// GetReplyByRootIDPage 回复分页查询（完全复用）
func GetReplyByRootIDPage(c *gin.Context) {
	req := new(GetReplyByRootIDPageRequest)
	res := new(GetReplyByRootIDPageResponse)

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidParams))
		return
	}

	total, list, err := suggestion.GetReplyByRootIDPage(c.Request.Context(), req.RootID, req.Page, req.PageSize)
	if err != nil {
		res.CodeOf(code.CodeServerBusy)
		res.StatusMsg = err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	res.Success()
	res.Total = total
	res.Page = req.Page
	res.PageSize = req.PageSize
	res.List = list
	c.JSON(http.StatusOK, res)
}
