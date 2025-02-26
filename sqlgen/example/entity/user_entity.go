package entity

import "time"

// User entity a user struct data.
type User struct {
	Id                uint32    `json:"id"`
	Uid               int64     `json:"uid"`
	NickName          string    `json:"nick_name"`
	AvatarUri         string    `json:"avatar_uri"`
	ReadingPreference int8      `json:"reading_preference"`
	CreateTime        time.Time `json:"create_time"`
	UpdateTime        time.Time `json:"update_time"`
	AutoBuy           int8      `json:"auto_buy"`
	IsAutoBuy         int8      `json:"is_auto_buy"`
}
