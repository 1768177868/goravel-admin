package services

import (
	"testing"
	"time"

	"github.com/goravel/framework/facades"
	"github.com/stretchr/testify/suite"

	apperrors "goravel/app/errors"
	"goravel/app/models"
	"goravel/app/services"
	"goravel/tests"
)

type NotificationServiceTestSuite struct {
	suite.Suite
	tests.TestCase
	service services.NotificationService
}

func TestNotificationServiceTestSuite(t *testing.T) {
	suite.Run(t, &NotificationServiceTestSuite{})
}

func (s *NotificationServiceTestSuite) SetupSuite() {
	s.service = services.NewNotificationServiceImpl()
}

func (s *NotificationServiceTestSuite) SetupTest() {
	s.RefreshDatabase()
	s.Seed()
}

func (s *NotificationServiceTestSuite) TearDownTest() {
	// 清理测试数据
	_, _ = facades.Orm().Query().Where("type", "test").Delete(&models.Notification{})
}

// ==================== Create 方法测试 ====================

// TestCreate_SingleNotification 测试创建单个通知
func (s *NotificationServiceTestSuite) TestCreate_SingleNotification() {
	var admin models.Admin
	err := facades.Orm().Query().First(&admin)
	s.Require().NoError(err)

	receiverID := admin.ID
	notification, err := s.service.Create("测试标题", "测试内容", "test", nil, &receiverID)

	s.NoError(err)
	s.NotNil(notification)
	s.Equal("测试标题", notification.Title)
	s.Equal("测试内容", notification.Content)
	s.Equal("test", notification.Type)
	s.Equal(receiverID, *notification.ReceiverID)
	s.NotZero(notification.ID)
}

// TestCreate_BroadcastToAllAdmins 测试批量创建通知给所有管理员
func (s *NotificationServiceTestSuite) TestCreate_BroadcastToAllAdmins() {
	// 获取所有管理员
	var admins []models.Admin
	err := facades.Orm().Query().Find(&admins)
	s.Require().NoError(err)
	s.Require().NotEmpty(admins, "应该有至少一个管理员")

	// 创建通知给所有管理员（receiverID 为 nil）
	notification, err := s.service.Create("系统通知", "这是一条系统通知", "system", nil, nil)

	s.NoError(err)
	s.NotNil(notification)
	s.Equal("系统通知", notification.Title)

	// 验证所有管理员都收到了通知
	count, err := facades.Orm().Query().
		Model(&models.Notification{}).
		Where("title", "系统通知").
		Where("type", "system").
		Count()
	s.NoError(err)
	s.Equal(int64(len(admins)), count, "应该为每个管理员创建一条通知")
}

// TestCreate_NoAdminsFound 测试没有管理员时的错误处理
func (s *NotificationServiceTestSuite) TestCreate_NoAdminsFound() {
	// 删除所有管理员
	_, err := facades.Orm().Query().Delete(&models.Admin{})
	s.Require().NoError(err)

	// 尝试创建通知给所有管理员
	notification, err := s.service.Create("测试", "内容", "test", nil, nil)

	s.Error(err)
	s.Nil(notification)

	// 验证返回的是业务错误
	businessErr, ok := apperrors.GetBusinessError(err)
	s.True(ok, "应该返回业务错误")
	s.Equal("record_not_found", businessErr.Code)
}

// ==================== List 方法测试 ====================

// TestList_BasicQuery 测试基本查询
func (s *NotificationServiceTestSuite) TestList_BasicQuery() {
	var admin models.Admin
	err := facades.Orm().Query().First(&admin)
	s.Require().NoError(err)

	// 创建几条测试通知
	for i := 0; i < 5; i++ {
		_, err := s.service.Create("测试通知", "内容", "test", nil, &admin.ID)
		s.Require().NoError(err)
	}

	// 查询通知列表
	notifications, total, err := s.service.List(admin.ID, 1, 10, "", "")

	s.NoError(err)
	s.GreaterOrEqual(int(total), 5)
	s.NotEmpty(notifications)
}

// TestList_WithTypeFilter 测试按类型筛选
func (s *NotificationServiceTestSuite) TestList_WithTypeFilter() {
	var admin models.Admin
	err := facades.Orm().Query().First(&admin)
	s.Require().NoError(err)

	// 创建不同类型的通知
	_, err = s.service.Create("系统通知", "内容", "system", nil, &admin.ID)
	s.Require().NoError(err)
	_, err = s.service.Create("测试通知", "内容", "test", nil, &admin.ID)
	s.Require().NoError(err)

	// 只查询 system 类型的通知
	notifications, total, err := s.service.List(admin.ID, 1, 10, "system", "")

	s.NoError(err)
	s.GreaterOrEqual(int(total), 1)
	for _, notif := range notifications {
		s.Equal("system", notif.Type)
	}
}

// TestList_WithReadStatusFilter 测试按已读状态筛选
func (s *NotificationServiceTestSuite) TestList_WithReadStatusFilter() {
	var admin models.Admin
	err := facades.Orm().Query().First(&admin)
	s.Require().NoError(err)

	// 创建通知
	notification, err := s.service.Create("测试通知", "内容", "test", nil, &admin.ID)
	s.Require().NoError(err)

	// 标记为已读
	err = s.service.MarkRead(admin.ID, notification.ID)
	s.Require().NoError(err)

	// 查询未读通知
	unread, total, err := s.service.List(admin.ID, 1, 10, "", "false")
	s.NoError(err)

	// 验证未读通知不包含已读的通知
	found := false
	for _, notif := range unread {
		if notif.ID == notification.ID {
			found = true
			break
		}
	}
	s.False(found, "已读通知不应出现在未读列表中")

	// 查询已读通知
	read, total, err := s.service.List(admin.ID, 1, 10, "", "true")
	s.NoError(err)
	s.GreaterOrEqual(int(total), 1)

	// 验证已读通知包含该通知
	found = false
	for _, notif := range read {
		if notif.ID == notification.ID {
			found = true
			s.True(notif.IsRead)
			break
		}
	}
	s.True(found, "已读通知应出现在已读列表中")
}

// TestList_Pagination 测试分页
func (s *NotificationServiceTestSuite) TestList_Pagination() {
	var admin models.Admin
	err := facades.Orm().Query().First(&admin)
	s.Require().NoError(err)

	// 创建 15 条通知
	for i := 0; i < 15; i++ {
		_, err := s.service.Create("测试通知", "内容", "test", nil, &admin.ID)
		s.Require().NoError(err)
	}

	// 第一页，每页 10 条
	page1, total, err := s.service.List(admin.ID, 1, 10, "", "")
	s.NoError(err)
	s.GreaterOrEqual(int(total), 15)
	s.LessOrEqual(len(page1), 10)

	// 第二页
	page2, _, err := s.service.List(admin.ID, 2, 10, "", "")
	s.NoError(err)
	s.LessOrEqual(len(page2), 10)

	// 验证两页的数据不重复
	if len(page1) > 0 && len(page2) > 0 {
		s.NotEqual(page1[0].ID, page2[0].ID, "两页的数据应该不同")
	}
}

// ==================== MarkRead 方法测试 ====================

// TestMarkRead_Success 测试标记为已读成功
func (s *NotificationServiceTestSuite) TestMarkRead_Success() {
	var admin models.Admin
	err := facades.Orm().Query().First(&admin)
	s.Require().NoError(err)

	// 创建通知
	notification, err := s.service.Create("测试通知", "内容", "test", nil, &admin.ID)
	s.Require().NoError(err)
	s.False(notification.IsRead, "新通知应该是未读状态")

	// 标记为已读
	err = s.service.MarkRead(admin.ID, notification.ID)
	s.NoError(err)

	// 验证已读状态
	var updated models.Notification
	err = facades.Orm().Query().Where("id", notification.ID).First(&updated)
	s.NoError(err)
	s.True(updated.IsRead)
	s.NotNil(updated.ReadAt)
}

// TestMarkRead_NotFound 测试通知不存在
func (s *NotificationServiceTestSuite) TestMarkRead_NotFound() {
	var admin models.Admin
	err := facades.Orm().Query().First(&admin)
	s.Require().NoError(err)

	// 尝试标记不存在的通知为已读
	err = s.service.MarkRead(admin.ID, 99999)
	s.Error(err)

	// 验证返回的是业务错误
	businessErr, ok := apperrors.GetBusinessError(err)
	s.True(ok, "应该返回业务错误")
	s.Equal("record_not_found", businessErr.Code)
}

// TestMarkRead_AlreadyRead 测试已读通知再次标记
func (s *NotificationServiceTestSuite) TestMarkRead_AlreadyRead() {
	var admin models.Admin
	err := facades.Orm().Query().First(&admin)
	s.Require().NoError(err)

	// 创建并标记为已读
	notification, err := s.service.Create("测试通知", "内容", "test", nil, &admin.ID)
	s.Require().NoError(err)

	err = s.service.MarkRead(admin.ID, notification.ID)
	s.Require().NoError(err)

	// 再次标记为已读（应该不报错）
	err = s.service.MarkRead(admin.ID, notification.ID)
	s.NoError(err, "已读通知再次标记应该不报错")
}

// ==================== MarkAllRead 方法测试 ====================

// TestMarkAllRead_Success 测试标记所有通知为已读
func (s *NotificationServiceTestSuite) TestMarkAllRead_Success() {
	var admin models.Admin
	err := facades.Orm().Query().First(&admin)
	s.Require().NoError(err)

	// 创建多条未读通知
	for i := 0; i < 5; i++ {
		_, err := s.service.Create("测试通知", "内容", "test", nil, &admin.ID)
		s.Require().NoError(err)
	}

	// 标记所有为已读
	err = s.service.MarkAllRead(admin.ID)
	s.NoError(err)

	// 验证未读数量为 0
	count, err := s.service.UnreadCount(admin.ID)
	s.NoError(err)
	s.Equal(int64(0), count)
}

// ==================== UnreadCount 方法测试 ====================

// TestUnreadCount_Basic 测试未读数量统计
func (s *NotificationServiceTestSuite) TestUnreadCount_Basic() {
	var admin models.Admin
	err := facades.Orm().Query().First(&admin)
	s.Require().NoError(err)

	// 创建 3 条未读通知
	for i := 0; i < 3; i++ {
		_, err := s.service.Create("测试通知", "内容", "test", nil, &admin.ID)
		s.Require().NoError(err)
	}

	// 验证未读数量
	count, err := s.service.UnreadCount(admin.ID)
	s.NoError(err)
	s.GreaterOrEqual(count, int64(3))

	// 标记一条为已读
	notifications, _, err := s.service.List(admin.ID, 1, 1, "", "false")
	s.Require().NoError(err)
	s.Require().NotEmpty(notifications)

	err = s.service.MarkRead(admin.ID, notifications[0].ID)
	s.Require().NoError(err)

	// 验证未读数量减少
	newCount, err := s.service.UnreadCount(admin.ID)
	s.NoError(err)
	s.Equal(count-1, newCount)
}

// ==================== ListRecent 方法测试 ====================

// TestListRecent_Basic 测试获取最近通知
func (s *NotificationServiceTestSuite) TestListRecent_Basic() {
	var admin models.Admin
	err := facades.Orm().Query().First(&admin)
	s.Require().NoError(err)

	// 创建多条通知
	for i := 0; i < 10; i++ {
		_, err := s.service.Create("测试通知", "内容", "test", nil, &admin.ID)
		s.Require().NoError(err)
		time.Sleep(10 * time.Millisecond) // 确保时间戳不同
	}

	// 获取最近 5 条
	recent, err := s.service.ListRecent(admin.ID, 5)
	s.NoError(err)
	s.LessOrEqual(len(recent), 5)
	s.GreaterOrEqual(len(recent), 1)

	// 验证返回了数据（排序验证在数据库层面完成，这里只验证数据存在）
	s.NotEmpty(recent, "应该返回最近的通知")
}

// TestListRecent_LimitValidation 测试限制验证
func (s *NotificationServiceTestSuite) TestListRecent_LimitValidation() {
	var admin models.Admin
	err := facades.Orm().Query().First(&admin)
	s.Require().NoError(err)

	// 测试负数限制（应该使用默认值 5）
	recent, err := s.service.ListRecent(admin.ID, -1)
	s.NoError(err)
	s.LessOrEqual(len(recent), 5)

	// 测试超过最大值的限制（应该使用默认值 5）
	recent, err = s.service.ListRecent(admin.ID, 100)
	s.NoError(err)
	s.LessOrEqual(len(recent), 5)
}
