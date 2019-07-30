package masala

import ("gopkg.in/mgo.v2"
	"time")

func NewMongoBuilder(config IConfig) *mgo.Database {
	info := &mgo.DialInfo{
		Addrs:    []string{config.GetHost()},
		Timeout:  60 * time.Second,
		Database: config.GetName(),
		Username: config.GetUser(),
		Password: config.GetPassword(),
	}
	session, err := mgo.DialWithInfo(info)
	if err != nil {
		panic(err)
	}
	return session.DB(config.GetName())
}