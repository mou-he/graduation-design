package suggestion

import (
	"context"
	"errors"
	"fmt"

	"github.com/mou-he/graduation-design/common/mysql"
	dao "github.com/mou-he/graduation-design/dao/suggestion"
	"github.com/mou-he/graduation-design/model"
)

// 创建主建议
func Publish(ctx context.Context, req *model.PublishRequest, userId, username string) (*model.Suggestion, error) {
	// 业务规则校验
	if userId == "" || username == "" {
		return nil, errors.New("请先登录")
	}
	if len(req.Content) == 0 || len(req.Content) > 500 {
		return nil, errors.New("内容不能为空且长度≤500字")
	}

	// 构建模型（删除LikeCount字段）
	suggestion := &model.Suggestion{
		RootID:     0,
		ParentID:   0,
		UserID:     userId,
		Username:   username,
		Content:    req.Content,
		ReplyCount: 0, // 初始化回复数为0
		IsDeleted:  0,
	}

	// 调用DAO入库
	if err := dao.CreateSuggestion(ctx, suggestion); err != nil {
		return nil, fmt.Errorf("发布失败：%w", err)
	}

	// 更新RootID
	if err := dao.UpdateSuggestionRootID(ctx, suggestion.ID, suggestion.ID); err != nil {
		_ = dao.SoftDeleteSuggestion(ctx, suggestion.ID) // 补偿
		return nil, fmt.Errorf("发布失败：%w", err)
	}

	return suggestion, nil
}

// 删除留言（移除所有点赞相关Redis清理逻辑）
func Delete(ctx context.Context, id uint64, userId string) error {
	// 校验
	if userId == "" {
		return errors.New("请先登录")
	}

	// 权限校验
	suggestion, err := dao.GetSuggestionByID(ctx, id)
	if err != nil {
		return fmt.Errorf("内容不存在：%w", err)
	}
	if suggestion.UserID != userId {
		return errors.New("无权限删除该内容")
	}

	// 开启事务：删除评论 + 更新回复数（如果是回复）
	tx := mysql.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var rootID uint64
	if suggestion.ParentID == 0 {
		// 主留言：级联删除
		rootID = suggestion.RootID
		if err := dao.SoftDeleteSuggestionsByRootIDWithTx(tx, rootID); err != nil {
			tx.Rollback()
			return fmt.Errorf("删除失败：%w", err)
		}
	} else {
		// 回复：单删 + 更新根评论回复数（-1）
		rootID = suggestion.RootID
		if err := dao.SoftDeleteSuggestionWithTx(tx, id); err != nil {
			tx.Rollback()
			return fmt.Errorf("删除失败：%w", err)
		}

		// 更新根评论回复数（减1）
		if err := dao.UpdateSuggestionReplyCountWithTx(tx, rootID, -1); err != nil {
			tx.Rollback()
			return fmt.Errorf("更新回复数失败：%w", err)
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("事务提交失败：%w", err)
	}

	return nil
}

// 更新留言
func Update(ctx context.Context, req *model.UpdateRequest, userId string) error {
	// 校验
	if userId == "" {
		return errors.New("请先登录")
	}
	if req.ID == 0 {
		return errors.New("内容ID不能为空")
	}
	if len(req.Content) == 0 || len(req.Content) > 500 {
		return errors.New("内容不能为空且长度≤500字")
	}

	// 权限校验
	suggestion, err := dao.GetSuggestionByID(ctx, req.ID)
	if err != nil {
		return fmt.Errorf("内容不存在：%w", err)
	}
	if suggestion.UserID != userId {
		return errors.New("无权限更新该内容")
	}

	// 更新内容
	if err := dao.UpdateSuggestionContent(ctx, req.ID, req.Content); err != nil {
		return fmt.Errorf("更新失败：%w", err)
	}

	return nil
}

// 回复留言（删除点赞相关逻辑 + 新增回复数更新）
func Reply(ctx context.Context, req *model.ReplyRequest, userId, username string) (*model.Suggestion, error) {
	// 校验
	if userId == "" || username == "" {
		return nil, errors.New("请先登录")
	}
	if req.ParentID == 0 {
		return nil, errors.New("请指定要回复的评论ID")
	}
	if len(req.Content) == 0 || len(req.Content) > 500 {
		return nil, errors.New("回复内容不能为空且长度≤500字")
	}

	// 校验父级存在
	parentSuggestion, err := dao.GetSuggestionByID(ctx, req.ParentID)
	if err != nil {
		return nil, fmt.Errorf("要回复的评论不存在或已删除：%w", err)
	}

	// 开启事务：新增回复 + 更新根评论回复数
	tx := mysql.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 构建回复模型（删除LikeCount字段）
	reply := &model.Suggestion{
		RootID:          parentSuggestion.RootID,
		ParentID:        req.ParentID,
		UserID:          userId,
		Username:        username,
		ReplyToUserID:   req.ReplyToUserID,
		ReplyToUsername: req.ReplyToUsername,
		Content:         req.Content,
		IsDeleted:       0,
	}

	// 入库（使用事务）
	if err := dao.CreateSuggestionWithTx(tx, reply); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("回复失败：%w", err)
	}

	// 更新根评论回复数（+1）
	if err := dao.UpdateSuggestionReplyCountWithTx(tx, parentSuggestion.RootID, 1); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("更新回复数失败：%w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("事务提交失败：%w", err)
	}

	return reply, nil
}

// GetRootSuggestionPage 根评论分页查询
func GetRootSuggestionPage(ctx context.Context, page, pageSize int) (int64, []model.RootSuggestionVO, error) {
	if page < 1 {
		return 0, nil, errors.New("页码不能小于1")
	}
	if pageSize < 1 || pageSize > 50 {
		return 0, nil, errors.New("每页条数需在1-50之间")
	}
	return dao.GetRootSuggestionPage(ctx, page, pageSize)
}

// GetReplyByRootIDPage 回复分页查询
func GetReplyByRootIDPage(ctx context.Context, rootID uint64, page, pageSize int) (int64, []model.Suggestion, error) {
	if rootID == 0 {
		return 0, nil, errors.New("根评论ID不能为空")
	}
	if page < 1 {
		return 0, nil, errors.New("页码不能小于1")
	}
	if pageSize < 1 || pageSize > 50 {
		return 0, nil, errors.New("每页条数需在1-50之间")
	}
	return dao.GetReplyByRootIDPage(ctx, rootID, page, pageSize)
}

// -------------------------- 为事务新增的DAO封装方法（需要在dao层补充） --------------------------
// 注意：需要在dao/suggestion包中补充以下方法，用于事务操作

// GetDB 获取DB实例（dao层补充）
// func (d *SuggestionDAO) GetDB() *gorm.DB {
//     return mysql.DB
// }

// CreateSuggestionWithTx 事务内新增（dao层补充）
// func CreateSuggestionWithTx(tx *gorm.DB, s *model.Suggestion) error {
//     return tx.Create(s).Error
// }

// SoftDeleteSuggestionWithTx 事务内软删除单条（dao层补充）
// func SoftDeleteSuggestionWithTx(tx *gorm.DB, id uint64) error {
//     return tx.Model(&model.Suggestion{}).
//         Where("id = ? AND is_deleted = ?", id, 0).
//         Update("is_deleted", 1).Error
// }

// SoftDeleteSuggestionsByRootIDWithTx 事务内级联删除（dao层补充）
// func SoftDeleteSuggestionsByRootIDWithTx(tx *gorm.DB, rootID uint64) error {
//     if err := tx.Model(&model.Suggestion{}).Where("id = ?", rootID).Update("is_deleted", 1).Error; err != nil {
//         return err
//     }
//     if err := tx.Model(&model.Suggestion{}).Where("root_id = ? AND parent_id != ?", rootID, 0).Update("is_deleted", 1).Error; err != nil {
//         return err
//     }
//     return nil
// }

// UpdateSuggestionReplyCountWithTx 事务内更新回复数（dao层补充）
// func UpdateSuggestionReplyCountWithTx(tx *gorm.DB, rootID uint64, num int) error {
//     return tx.Model(&model.Suggestion{}).
//         Where("id = ? AND is_deleted = ?", rootID, 0).
//         Update("reply_count", gorm.Expr("reply_count + ?", num)).Error
// }
