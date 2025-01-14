package mutations

import (
	"sync"

	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/redis"
)

type ListItemAction string

const (
	ListItemActionAdd    ListItemAction = "ADD"
	ListItemActionUpdate ListItemAction = "UPDATE"
	ListItemActionRemove ListItemAction = "REMOVE"
)

type Mutate struct {
	mongo mongo.Instance
	redis redis.Instance
	mx    map[string]*sync.Mutex
}

func New(mongoInst mongo.Instance, redisInst redis.Instance) *Mutate {
	return &Mutate{
		mongo: mongoInst,
		redis: redisInst,
		mx:    map[string]*sync.Mutex{},
	}
}
