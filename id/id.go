package libid

import "github.com/satori/go.uuid"

type UniqueIDGenerator func() string

func GenerateUniqueID() string {
	return uuid.NewV4().String()
}
