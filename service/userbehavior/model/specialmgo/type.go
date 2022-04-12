package specialmgo

// 专题
type Special struct {
	ID       string `bson:"_id"`
	AuthorId string `bson:"authorId"`  //作者id
	AuthorNick int    `bson:"authorNick"`
}
