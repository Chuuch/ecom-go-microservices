package misstypemanager

import (
	"strings"
	"sync"

	"github.com/chuuch/search-microservice/pkg/logger"
)

type MisstypeManager interface {
	GetMissTypeWord(word string) string
}

type KeyboardMisstypeManager struct {
	log         logger.Logger
	keyMappings map[string]string
	sbPool      *sync.Pool
}

func NewMisstypeManager(log logger.Logger, keyMappings map[string]string) *KeyboardMisstypeManager {
	sbPool := &sync.Pool{New: func() any { return new(strings.Builder) }}
	return &KeyboardMisstypeManager{
		log:         log,
		keyMappings: keyMappings,
		sbPool:      sbPool,
	}
}

func (k *KeyboardMisstypeManager) GetMissTypeWord(originalWord string) string {
	sb := k.sbPool.Get().(*strings.Builder)
	defer k.sbPool.Put(sb)
	sb.Reset()

	for _, c := range []rune(originalWord) {
		lowerCasedChar := strings.ToLower(string(c))
		if mappedChar, ok := k.keyMappings[lowerCasedChar]; ok {
			sb.WriteString(mappedChar)
		} else {
			sb.WriteString(lowerCasedChar)
		}
	}
	return sb.String()
}
