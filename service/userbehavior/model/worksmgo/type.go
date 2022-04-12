package worksmgo

// 作品
type Works struct {
    ID         string `bson:"_id"`
    AuthorId   string `bson:"author_id"` // 作品id
    AuthorNick int    `bson:"author_nick"`
    LikeNum    int    `bson:"like_num"`
}
