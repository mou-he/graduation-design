package router

import (
	"github.com/gin-gonic/gin"
	suggestion "github.com/mou-he/graduation-design/controller/suggestion"
)

func SuggestionRouter(r *gin.RouterGroup) {
	{
		r.POST("/publish", suggestion.PublishSuggestion)       // 发布主留言
		r.POST("/reply", suggestion.ReplySuggestion)           // 回复留言
		r.POST("/update", suggestion.UpdateSuggestion)         // 更新留言
		r.POST("/delete", suggestion.DeleteSuggestion)         // 删除留言
		r.POST("/root/page", suggestion.GetRootSuggestionPage) // 根留言分页
		r.POST("/reply/page", suggestion.GetReplyByRootIDPage) // 回复分页
	}
}
