package web

// VO view object，对标前端的

type ArticleVo struct {
	Id         int64  `json:"id,omitempty"`
	Title      string `json:"title,omitempty"`
	Abstract   string `json:"abstract,omitempty"`
	Content    string `json:"content,omitempty"`
	AuthorId   int64  `json:"authorId,omitempty"`
	AuthorName string `json:"authorName,omitempty"`
	Status     uint8  `json:"status,omitempty"`
	//计数
	ReadCnt    int64 `json:"readCnt,omitempty"`
	LikeCnt    int64 `json:"likeCnt,omitempty"`
	CollectCnt int64 `json:"collectCnt,omitempty"`
	//我个人有没有收藏，有没有点赞
	Liked     bool `json:"liked"`
	Collected bool `json:"collected"`

	Ctime string `json:"ctime,omitempty"`
	Utime string `json:"utime,omitempty"`
}
