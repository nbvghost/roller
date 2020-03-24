package iface

type ILoginBridge interface {
	Online(UserID uint64, cluster ICluster) (IUser, error)
	Offline(UserID uint64, cluster ICluster)
	CheckToke(Token string) string
	/*GetAS(ActionCode ac.KV, args ...interface{}) *pb.ActionStatus
	CreateGameInfoByUserID(DB *gorm.DB, Redis *redis.Client, User *dao.User, DataTableID dao.FactoryDataTableID) (err error, oldUser bool, factoryID uint64)
	LoginBefore(DB *gorm.DB, cluster *Cluster) error
	GetInvitee(DB *gorm.DB, InviteeUserID uint64) (error, dao.Invite)
	AddInvite(DB *gorm.DB, target *dao.Invite) error
	AddFriend(DB *gorm.DB, SUserID, TUserID uint64) (proto.Message, dao.Friend)
	CacheSession(cluster *Cluster)
	PersistenceSession(session *Session)
	SetLastDaySCoinRank(wsUsers *Cluster)
	GetClusterInfos() []map[string]interface{}
	GetOrm() *gorm.DB
	GetRedis() *redis.Client*/
}

/*type IActionStatusService interface {
	GetAS(ActionCode ac.KV, args ...interface{}) *pb.ActionStatus
}
type IUserService interface {
	CreateGameInfoByUserID(DB *gorm.DB, Redis *redis.Client, UserID uint64, DataTableID dao.FactoryDataTableID) (err error, oldUser bool, factoryID uint64)
	LoginBefore(DB *gorm.DB, cluster *Cluster) error
}

type IInviteService interface {
	GetInvitee(DB *gorm.DB, InviteeUserID uint64) (error, dao.Invite)
	AddInvite(DB *gorm.DB, target *dao.Invite) error
}
type IFriendService interface {
	AddFriend(DB *gorm.DB, SUserID, TUserID uint64) (proto.Message, dao.Friend)
}
type ICacheService interface {
	CacheSession(cluster *Cluster)
	PersistenceSession(session *Session)
}
type  IFunctionItem interface {
	SetLastDaySCoinRank(wsUsers *Cluster)
}*/
