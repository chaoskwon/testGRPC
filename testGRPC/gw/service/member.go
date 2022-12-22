package service

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testGRPC/gw/entity"

	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
)

func ValidateRequesst(param string) ([]int64, error) {
	logrus.Trace("")

	var err error
	arr := strings.Split(param, ",")

	ids := make([]int64, len(arr))
	for i, id := range arr {
		ids[i], err = strconv.ParseInt(id, 10, 64)
		if err != nil {
			logrus.Error(err.Error())
			return nil, err
		}
	}

	return ids, nil
}

func GetList(ctx context.Context, ids []int64) ([]entity.Member, error) {
	logrus.Trace("")

	var members []entity.Member

	v := ctx.Value("xormDB")
	if v == nil {
		msg := "xormDB를 찾을 수가 없습니다"
		logrus.Error(msg)
		return members, errors.New(msg)
	}

	xormDB, ok := v.(*xorm.Engine)
	if !ok {
		msg := "xormDB 변환중 오류가 발생 했습니다"
		logrus.Error(msg)
		return members, errors.New(msg)
	}

	tableNmae := entity.Member{}.TableName()
	if err := xormDB.Table(tableNmae).In("id", ids).Find(&members); err != nil {
		return members, err
	}

	return members, nil
}
