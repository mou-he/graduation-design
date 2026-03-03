package suggestion

import (
	"context"

	"github.com/mou-he/graduation-design/common/mysql"
	"github.com/mou-he/graduation-design/model"
	"gorm.io/gorm"
)

// -------------------------- 基础原子操作（仅数据库操作，无业务校验） --------------------------
// CreateSuggestion 新增留言/回复（纯入库）
func CreateSuggestion(ctx context.Context, s *model.Suggestion) error {
	return mysql.DB.WithContext(ctx).Create(s).Error
}

// GetSuggestionByID 查询单条未删除的留言/回复
func GetSuggestionByID(ctx context.Context, id uint64) (*model.Suggestion, error) {
	var s model.Suggestion
	err := mysql.DB.WithContext(ctx).
		Where("id = ? AND is_deleted = ?", id, 0).
		First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// UpdateSuggestionContent 仅更新内容
func UpdateSuggestionContent(ctx context.Context, id uint64, content string) error {
	return mysql.DB.WithContext(ctx).Model(&model.Suggestion{}).
		Where("id = ? AND is_deleted = ?", id, 0).
		Update("content", content).Error
}

// SoftDeleteSuggestion 软删除单条内容
func SoftDeleteSuggestion(ctx context.Context, id uint64) error {
	return mysql.DB.WithContext(ctx).Model(&model.Suggestion{}).
		Where("id = ? AND is_deleted = ?", id, 0).
		Update("is_deleted", 1).Error
}

// SoftDeleteSuggestionsByRootID 级联软删除主留言+所有回复
func SoftDeleteSuggestionsByRootID(ctx context.Context, rootID uint64) error {
	tx := mysql.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除根评论
	if err := tx.Model(&model.Suggestion{}).Where("id = ?", rootID).Update("is_deleted", 1).Error; err != nil {
		tx.Rollback()
		return err
	}
	// 删除所有回复
	if err := tx.Model(&model.Suggestion{}).Where("root_id = ? AND parent_id != ?", rootID, 0).Update("is_deleted", 1).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// UpdateSuggestionRootID 更新留言的RootID
func UpdateSuggestionRootID(ctx context.Context, id uint64, rootID uint64) error {
	return mysql.DB.WithContext(ctx).Model(&model.Suggestion{}).
		Where("id = ? AND is_deleted = ?", id, 0).
		Update("root_id", rootID).Error
}

// UpdateSuggestionReplyCount 更新根评论的回复数（新增方法）
func UpdateSuggestionReplyCount(ctx context.Context, rootID uint64, num int) error {
	return mysql.DB.WithContext(ctx).Model(&model.Suggestion{}).
		Where("id = ? AND is_deleted = ?", rootID, 0).
		Update("reply_count", gorm.Expr("reply_count + ?", num)).Error
}

// -------------------------- 分页核心功能 --------------------------
// GetRootSuggestionPage 根评论分页查询（含下属回复数）
func GetRootSuggestionPage(ctx context.Context, page, pageSize int) (int64, []model.RootSuggestionVO, error) {
	offset := (page - 1) * pageSize

	type rootSuggestionDTO struct {
		model.Suggestion
		// 兼容：如果数据库中reply_count字段已存在，直接取；否则用子查询统计
		ReplyCount int64 `gorm:"column:reply_count"`
	}

	var (
		total       int64
		rootDTOList []rootSuggestionDTO
		rootVOList  []model.RootSuggestionVO
	)

	// 查询根评论总数
	if err := mysql.DB.WithContext(ctx).
		Model(&model.Suggestion{}).
		Where("parent_id = ? AND is_deleted = ?", 0, 0).
		Count(&total).Error; err != nil {
		return 0, nil, err
	}

	// 子查询统计回复数（如果数据库字段未同步，用这个；同步后可直接取reply_count）
	subQuery := mysql.DB.Model(&model.Suggestion{}).
		Select("COUNT(*)").
		Where("root_id = suggestion_board.id AND parent_id != ? AND is_deleted = ?", 0, 0)

	// 查询根评论列表 + 回复数
	if err := mysql.DB.WithContext(ctx).
		Model(&model.Suggestion{}).
		// 优先使用数据库自带的reply_count字段，没有则用子查询
		Select("suggestion_board.*, COALESCE(suggestion_board.reply_count, (?)) as reply_count", subQuery).
		Where("parent_id = ? AND is_deleted = ?", 0, 0).
		Order("create_time DESC").
		Limit(pageSize).
		Offset(offset).
		Scan(&rootDTOList).Error; err != nil {
		return 0, nil, err
	}

	// 转换为VO（删除LikeCount，只保留ReplyCount）
	for _, dto := range rootDTOList {
		rootVOList = append(rootVOList, model.RootSuggestionVO{
			ID:         dto.ID,
			Content:    dto.Content,
			Username:   dto.Username,
			CreateTime: dto.CreateTime,
			ReplyCount: dto.ReplyCount, // 回复数（数据库字段或子查询统计）
		})
	}

	return total, rootVOList, nil
}

// GetReplyByRootIDPage 根评论下的回复分页查询
func GetReplyByRootIDPage(ctx context.Context, rootID uint64, page, pageSize int) (int64, []model.Suggestion, error) {
	offset := (page - 1) * pageSize

	var (
		total     int64
		replyList []model.Suggestion
	)

	// 查询回复总数
	if err := mysql.DB.WithContext(ctx).
		Model(&model.Suggestion{}).
		Where("root_id = ? AND parent_id != ? AND is_deleted = ?", rootID, 0, 0).
		Count(&total).Error; err != nil {
		return 0, nil, err
	}

	// 分页查询回复列表
	if err := mysql.DB.WithContext(ctx).
		Where("root_id = ? AND parent_id != ? AND is_deleted = ?", rootID, 0, 0).
		Order("create_time ASC").
		Limit(pageSize).
		Offset(offset).
		Find(&replyList).Error; err != nil {
		return 0, nil, err
	}

	return total, replyList, nil
}

// GetSuggestionIDsByRootID 根据根ID查询所有相关内容ID（清理Redis用）
func GetSuggestionIDsByRootID(ctx context.Context, rootID uint64) ([]uint64, error) {
	var ids []uint64
	err := mysql.DB.WithContext(ctx).
		Model(&model.Suggestion{}).
		Where("root_id = ? AND is_deleted = ?", rootID, 0).
		Pluck("id", &ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}

// CreateSuggestionWithTx 事务内新增留言/回复
func CreateSuggestionWithTx(tx *gorm.DB, s *model.Suggestion) error {
	return tx.Create(s).Error
}

// SoftDeleteSuggestionWithTx 事务内软删除单条内容
func SoftDeleteSuggestionWithTx(tx *gorm.DB, id uint64) error {
	return tx.Model(&model.Suggestion{}).
		Where("id = ? AND is_deleted = ?", id, 0).
		Update("is_deleted", 1).Error
}

// SoftDeleteSuggestionsByRootIDWithTx 事务内级联软删除主留言+所有回复
func SoftDeleteSuggestionsByRootIDWithTx(tx *gorm.DB, rootID uint64) error {
	if err := tx.Model(&model.Suggestion{}).Where("id = ?", rootID).Update("is_deleted", 1).Error; err != nil {
		return err
	}
	if err := tx.Model(&model.Suggestion{}).Where("root_id = ? AND parent_id != ?", rootID, 0).Update("is_deleted", 1).Error; err != nil {
		return err
	}
	return nil
}

// UpdateSuggestionReplyCountWithTx 事务内更新根评论的回复数
func UpdateSuggestionReplyCountWithTx(tx *gorm.DB, rootID uint64, num int) error {
	return tx.Model(&model.Suggestion{}).
		Where("id = ? AND is_deleted = ?", rootID, 0).
		Update("reply_count", gorm.Expr("reply_count + ?", num)).Error
}
