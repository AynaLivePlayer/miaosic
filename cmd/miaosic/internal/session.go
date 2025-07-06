package internal

import (
	"encoding/json"
	"fmt"
	"github.com/AynaLivePlayer/miaosic"
	"os"
	"path/filepath"
)

var (
	sessions = make(map[string]string)
)

func RestoreSessions(sessionFile string) error {
	if sessionFile == "" {
		return nil
	}

	data, err := os.ReadFile(sessionFile)
	if err != nil {
		if !os.IsNotExist(err) {
			// 仅当文件存在且读取错误时打印日志
		}
		return err
	}

	err = json.Unmarshal(data, &sessions)
	if err != nil {
		return err
	}

	for providerName, session := range sessions {
		provider, ok := miaosic.GetProvider(providerName)
		if !ok {
			continue
		}
		if loginable, ok := provider.(miaosic.Loginable); ok {
			err = loginable.RestoreSession(session)
			if err != nil {
				fmt.Printf("failed to restore session for provider %s err: %s", providerName, err)
			}
		}
	}
	return nil
}

func SaveSessions(sessionFile string) error {
	if sessionFile == "" {
		return nil
	}

	for _, providerName := range miaosic.ListAvailableProviders() {
		provider, ok := miaosic.GetProvider(providerName)
		if !ok {
			continue
		}
		if loginable, ok := provider.(miaosic.Loginable); ok {
			SetSession(providerName, loginable.SaveSession())
		}
	}

	data, err := json.MarshalIndent(sessions, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(sessionFile), 0755); err != nil {
		return err
	}

	return os.WriteFile(sessionFile, data, 0600)
}

func GetSession(provider string) (string, bool) {
	val, ok := sessions[provider]
	return val, ok
}

func SetSession(provider, session string) {
	sessions[provider] = session
}
