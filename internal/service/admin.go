/*
 * @Author: cloudyi.li
 * @Date: 2023-05-15 13:30:31
 * @LastEditTime: 2023-05-25 16:10:35
 * @LastEditors: cloudyi.li
 * @FilePath: /chatserver-api/internal/service/admin.go
 */
package service

import (
	"chatserver-api/internal/consts"
	"chatserver-api/internal/dao"
	"chatserver-api/internal/model"
	"chatserver-api/internal/model/entity"
	"chatserver-api/pkg/cache"
	"chatserver-api/pkg/logger"
	"chatserver-api/utils/uuid"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

var _ AdminService = (*adminService)(nil)

type AdminService interface {
	AdminVerify(ctx *gin.Context) bool
	CdKeyGenerate(ctx *gin.Context, number int, CardId int64) (res model.CdKeyGenerateRes, err error)
	GiftCardUpdate(ctx *gin.Context, req model.GiftCardUpdate) error
	GiftCardCreate(ctx *gin.Context, req model.GiftCardCreate) error
}

// userService 实现UserService接口
type adminService struct {
	kd   dao.CDkeyDao
	ud   dao.UserDao
	aSrv uuid.SnowNode
	rc   *redis.Client
}

func NewAdminService(_kd dao.CDkeyDao, _ud dao.UserDao) *adminService {
	return &adminService{
		kd:   _kd,
		ud:   _ud,
		aSrv: *uuid.NewNode(5),
		rc:   cache.GetRedisClient(),
	}
}

func (as *adminService) AdminVerify(ctx *gin.Context) bool {
	userId := ctx.GetInt64(consts.UserID)
	role, err := as.ud.UserGetRole(ctx, userId)
	if err != nil || role != consts.Administrator {
		return false
	} else {
		return true
	}
}

func (as *adminService) CdKeyGenerate(ctx *gin.Context, number int, CardId int64) (res model.CdKeyGenerateRes, err error) {
	var cdkey entity.CdKey
	var cdkeylist []entity.CdKey
	var codekey []string
	for i := 0; i < number; i++ {
		keyId := as.aSrv.GenSnowID()
		code := uuid.IdToCode(keyId)
		cdkey.Id = keyId
		cdkey.CodeKey = code
		cdkey.GiftCardId = CardId
		cdkeylist = append(cdkeylist, cdkey)
		codekey = append(codekey, code)
	}
	err = as.kd.CdKeyGenerate(ctx, cdkeylist)
	res.CodeKey = codekey
	return
}

func (as *adminService) GiftCardCreate(ctx *gin.Context, req model.GiftCardCreate) error {
	err := as.rc.Del(ctx, consts.GiftcardPrefix+"0").Err()
	if err != nil {
		logger.Errorf("删除GiftcardList缓存失败:%v", err.Error())
	}
	var giftcard entity.GiftCard
	giftcard.Id = as.aSrv.GenSnowID()
	giftcard.CardAmount = req.CardAmount
	giftcard.CardDiscount = req.CardDiscount
	giftcard.CardName = req.CardName
	giftcard.CardBuyLink = req.CardLink
	giftcard.CardComment = req.CardComment
	return as.kd.GiftCardCreate(ctx, &giftcard)
}

func (as *adminService) GiftCardUpdate(ctx *gin.Context, req model.GiftCardUpdate) error {
	var giftcard entity.GiftCard
	cardId, err := strconv.ParseInt(req.CardId, 10, 64)
	if err != nil {
		return err
	}
	giftcard.Id = cardId
	giftcard.CardAmount = req.CardAmount
	giftcard.CardDiscount = req.CardDiscount
	giftcard.CardName = req.CardName
	giftcard.CardBuyLink = req.CardLink
	giftcard.CardComment = req.CardComment
	return as.kd.GiftCardUpdate(ctx, &giftcard)
}
