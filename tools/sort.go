package tools

import "time"

type MyUint64List []uint64

func (my64 MyUint64List) Len() int           { return len(my64) }
func (my64 MyUint64List) Swap(i, j int)      { my64[i], my64[j] = my64[j], my64[i] }
func (my64 MyUint64List) Less(i, j int) bool { return my64[i] < my64[j] }

// DiffNano 时间差，纳秒
func DiffNano(startTime time.Time) (diff int64) {
	diff = int64(time.Since(startTime))
	return
}

// InArrayStr 判断字符串是否在数组内
func InArrayStr(str string, arr []string) (inArray bool) {
	for _, s := range arr {
		if s == str {
			inArray = true
			break
		}
	}
	return
}
