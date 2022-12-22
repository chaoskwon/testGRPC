package entity

type Member struct {
	MemberID     int64  `xorm:"id"`
	MemberName   string `xorm:"name"`
	Mobile       string `xorm:"mobile"`
	MaskedMobile string `xorm:"-"`
	Email        string `xorm:"email"`
	OrgID        int64  `xorm:"org_id"`
}

func (Member) TableName() string {
	return "members"
}
