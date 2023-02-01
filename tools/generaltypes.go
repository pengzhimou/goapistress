package tools

import (
	"fmt"
	"strings"
)

// Slice 自定义数组参数
type FlagSlice []string

// String string
func (f *FlagSlice) String() string {
	return fmt.Sprint(*f)
}

// Set set
func (f *FlagSlice) Set(s string) error {
	*f = append(*f, s)
	return nil
}

// Map 自定义数组参数
type FlagMap map[string]string

// String string
func (f FlagMap) String() string {
	return fmt.Sprintf("%v", map[string]string(f))
}

// Set set
func (f FlagMap) Set(value string) error {
	split := strings.SplitN(value, "=", 2)
	f[split[0]] = split[1]
	return nil
}
