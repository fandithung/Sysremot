package main

import (
	"fmt"
	"time"

	sigar "github.com/elastic/gosigar"
)

func filesystemJob(roundedTime time.Time) {
	conn := pool.Get()
	defer conn.Close()

	fslist := sigar.FileSystemList{}
	err := fslist.Get()
	if err != nil {
		errLogger.Println("FilesystemJob error: ", err)
		return
	}

	for _, fs := range fslist.List {
		dirname := fs.DirName
		if dirname != "/" && dirname != "/home" {
			continue
		}

		usage := sigar.FileSystemUsage{}
		usage.Get(dirname)

		if usage.Total == 0 {
			continue
		}

		data := fmt.Sprintf(`{"total":"%d","used":"%d","free":"%d","in-percent":"%d"}`,
			usage.Total, usage.Used, usage.Avail, usePercent(usage.Total, usage.Avail))

		currentKey := fmt.Sprintf("%s|fs:%s|current", AppName, dirname)
		historyKey := fmt.Sprintf("%s|fs:%s|%s|%s", AppName, dirname, roundedTime.Format("2006-01-02"), roundedTime.Format("15:04:05"))

		conn.Send("MULTI")
		conn.Send("SETEX", currentKey, TheTicker.Seconds()*2, data)
		conn.Send("SETEX", historyKey, ExpireInterval, data)
		_, err = conn.Do("EXEC")
		if err != nil {
			errLogger.Println(err)
		}
	}
}
