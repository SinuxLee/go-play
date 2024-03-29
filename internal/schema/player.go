package schema

type Player struct {
	UserId         uint64 `json:"userId"`
	Avatar         uint16 `json:"avatar"`         // 头像
	AvatarURL      string `json:"avatarURL"`      // 头像链接
	AvatarFrame    uint16 `json:"avatarFrame"`    // 头像框
	Level          uint16 `json:"level"`          // 等级
	GuildId        uint32 `json:"guildId"`        // 工会编号
	GuildName      string `json:"guildName"`      // 工会名称
	GuildIcon      uint16 `json:"guildIcon"`      // 工会图标
	GuildIconFrame uint16 `json:"guildIconFrame"` // 工会图标框
	GuildIconBG    uint32 `json:"guildIconBG"`    // 工会图标背景
	Nick           string `json:"nick"`           // 昵称
	Likes          uint32 `json:"likes"`          // 点赞数
	GoldBadge      uint32 `json:"goldBadge"`      // 金牌数
	SilverBadge    uint32 `json:"silverBadge"`    // 银牌数
	BronzeBadge    uint32 `json:"bronzeBadge"`    // 铜牌数
	CollectLevel   uint32 `json:"collLvl"`        // 收集等级
}
