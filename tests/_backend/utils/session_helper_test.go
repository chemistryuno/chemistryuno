package utils

import (
	"testing"
)

func TestGenerateSessionID_Uniqueness(t *testing.T) {
	// 生成10000个Session ID，检查是否有重复
	ids := make(map[string]bool)
	count := 10000

	for i := 0; i < count; i++ {
		sid := GenerateSessionID()
		if sid == "" {
			t.Errorf("Session ID生成失败，返回空字符串")
			continue
		}
		if ids[sid] {
			t.Errorf("发现重复的Session ID: %s", sid)
		}
		ids[sid] = true
	}

	if len(ids) != count {
		t.Errorf("期望生成 %d 个唯一ID，实际生成 %d 个", count, len(ids))
	}
}

func TestGenerateSessionID_Length(t *testing.T) {
	sid := GenerateSessionID()
	if len(sid) != 32 {
		t.Errorf("Session ID长度应为32个字符（16字节的hex编码），实际: %d", len(sid))
	}
}

func TestGenerateSessionID_HexFormat(t *testing.T) {
	sid := GenerateSessionID()
	for _, c := range sid {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("Session ID应只包含十六进制字符，发现非法字符: %c", c)
		}
	}
}

func TestGenerateSessionID_NeverEmpty(t *testing.T) {
	// 测试100次，确保没有返回空字符串
	for i := 0; i < 100; i++ {
		sid := GenerateSessionID()
		if sid == "" {
			t.Errorf("第 %d 次生成Session ID失败，返回空字符串", i+1)
		}
		if len(sid) == 0 {
			t.Errorf("第 %d 次生成的Session ID长度为0", i+1)
		}
	}
}

func TestIsDuplicateKeyError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "SQLite unique constraint",
			err:      &mockError{"UNIQUE constraint failed: user_sessions.id"},
			expected: true,
		},
		{
			name:     "MySQL duplicate entry",
			err:      &mockError{"Duplicate entry 'abc123' for key 'PRIMARY'"},
			expected: true,
		},
		{
			name:     "Generic duplicate key",
			err:      &mockError{"duplicate key value violates unique constraint"},
			expected: true,
		},
		{
			name:     "Other error",
			err:      &mockError{"connection timeout"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isDuplicateKeyError(tt.err)
			if result != tt.expected {
				t.Errorf("isDuplicateKeyError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// mockError 用于测试的模拟错误类型
type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}
