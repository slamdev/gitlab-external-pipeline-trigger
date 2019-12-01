package internal

import (
	"errors"
	"fmt"
	"strings"
)

type MapFlags map[string]string

func (i *MapFlags) String() string {
	mapFlag := *i
	return fmt.Sprintf("%+v", mapFlag)
}

func (i *MapFlags) Set(value string) error {
	parsed := strings.Split(value, ":")
	if len(parsed) != 2 {
		return errors.New("key value pair should be split with : ")
	}
	if *i == nil {
		*i = make(map[string]string)
	}
	mapFlag := *i
	mapFlag[parsed[0]] = parsed[1]
	return nil
}
