package utils

import (
	"chemistryuno/backend/repository"
	"fmt"
	"log"
	"strings"
)

const nicknameGenerationAttempts = 8

// GenerateUniqueRandomNickname returns a valid nickname that does not collide
// with existing users. fallbackKey should be stable, such as a UID or username.
func GenerateUniqueRandomNickname(base, fallbackKey string, exists func(string) (bool, error)) (string, error) {
	base = strings.TrimSpace(base)
	if base == "" {
		base = "研究员"
	}

	for i := 0; i < nicknameGenerationAttempts; i++ {
		candidate := GenerateRandomNickname(base)
		ok, err := nicknameAvailable(candidate, exists)
		if err != nil {
			return "", err
		}
		if ok {
			return candidate, nil
		}
	}

	fallbackKey = strings.TrimSpace(fallbackKey)
	if fallbackKey == "" {
		fallbackKey = "000000"
	}
	candidate := sanitizeNicknameFallback(base + fallbackKey)
	if len([]rune(candidate)) > 20 {
		runes := []rune(candidate)
		candidate = string(runes[:20])
	}
	ok, err := nicknameAvailable(candidate, exists)
	if err != nil {
		return "", err
	}
	if ok {
		return candidate, nil
	}

	for i := 0; i < nicknameGenerationAttempts; i++ {
		candidate = sanitizeNicknameFallback(fmt.Sprintf("%s%s%d", base, fallbackKey, i+1))
		if len([]rune(candidate)) > 20 {
			runes := []rune(candidate)
			candidate = string(runes[:20])
		}
		ok, err = nicknameAvailable(candidate, exists)
		if err != nil {
			return "", err
		}
		if ok {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("unable to generate unique nickname for %q", fallbackKey)
}

func nicknameAvailable(candidate string, exists func(string) (bool, error)) (bool, error) {
	if strings.TrimSpace(candidate) == "" {
		return false, nil
	}
	taken, err := exists(candidate)
	if err != nil {
		return false, err
	}
	return !taken, nil
}

func sanitizeNicknameFallback(value string) string {
	var b strings.Builder
	for _, r := range value {
		switch {
		case r == '_':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= '\u4e00' && r <= '\u9fa5':
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "研究员000000"
	}
	return b.String()
}

// RepairMissingNicknames assigns random nicknames to users with empty nicknames.
func RepairMissingNicknames(userRepo *repository.UserRepository) (int, error) {
	users, err := userRepo.FindUsersWithBlankNickname()
	if err != nil {
		return 0, err
	}
	if len(users) == 0 {
		log.Println("✅ 昵称启动检查完成：未发现缺失昵称的玩家")
		return 0, nil
	}

	repaired := 0
	for _, user := range users {
		nickname, err := GenerateUniqueRandomNickname("研究员", fmt.Sprintf("%d", user.UID), userRepo.ExistsByNickname)
		if err != nil {
			return repaired, fmt.Errorf("generate nickname for uid %d: %w", user.UID, err)
		}
		if err := userRepo.UpdateNickname(user.UID, nickname); err != nil {
			return repaired, fmt.Errorf("update nickname for uid %d: %w", user.UID, err)
		}
		repaired++
	}

	log.Printf("✅ 昵称启动检查完成：已为 %d/%d 位缺失昵称的玩家分配随机昵称", repaired, len(users))
	return repaired, nil
}
