//go:build !windows
// +build !windows

package storage

import (
	"os"
	"os/user"
	"strconv"
	"sync"
	"syscall"

	"github.com/bnb-chain/greenfield-storage-provider/util/log"

	"github.com/pkg/sftp"
)

var (
	uidMap = make(map[int]string)
	gidMap = make(map[int]string)
	mutex  sync.Mutex
)

func getOwnerGroup(info os.FileInfo) (string, string) {
	mutex.Lock()
	defer mutex.Unlock()
	var owner, group string
	switch st := info.Sys().(type) {
	case *syscall.Stat_t:
		owner = userName(int(st.Uid))
		group = groupName(int(st.Gid))
	case *sftp.FileStat:
		owner = userName(int(st.UID))
		group = groupName(int(st.GID))
	}
	return owner, group
}

func userName(uid int) string {
	name, ok := uidMap[uid]
	if !ok {
		u, err := user.LookupId(strconv.Itoa(uid))
		if err != nil {
			log.Warnf("lookup uid %d: %s", uid, err)
			name = strconv.Itoa(uid)
		} else {
			name = u.Username
		}
		uidMap[uid] = name
	}
	return name
}

func groupName(gid int) string {
	name, ok := gidMap[gid]
	if !ok {
		g, err := user.LookupGroupId(strconv.Itoa(gid))
		if err != nil {
			log.Warnf("lookup gid %d: %s", gid, err)
			name = strconv.Itoa(gid)
		} else {
			name = g.Name
		}
		gidMap[gid] = name
	}
	return name
}
