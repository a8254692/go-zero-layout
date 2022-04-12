package usermgo

type User struct {
	ID       string `bson:"_id"`
	NickName string `bson:"nick_name"`
	AvatarId int    `bson:"avatarId"`
}
