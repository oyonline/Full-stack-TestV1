package authctx

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"
	"github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth/user"
)

const (
	claimPrimaryRoleID   = "primaryRoleId"
	claimPrimaryRoleKey  = "primaryRoleKey"
	claimPrimaryRoleName = "primaryRoleName"
	claimRoleIDs         = "roleIds"
	claimRoleKeys        = "roleKeys"
	claimRoleNames       = "roleNames"
)

func GetPrimaryRoleID(c *gin.Context) int {
	if value, ok := claimInt(c, claimPrimaryRoleID); ok && value > 0 {
		return value
	}
	return user.GetRoleId(c)
}

func GetPrimaryRoleName(c *gin.Context) string {
	if value, ok := claimString(c, claimPrimaryRoleName); ok && value != "" {
		return value
	}
	if value, ok := claimString(c, jwt.RoleNameKey); ok && value != "" {
		return value
	}
	return user.GetRoleName(c)
}

func GetPrimaryRoleKey(c *gin.Context) string {
	if value, ok := claimString(c, claimPrimaryRoleKey); ok && value != "" {
		return value
	}
	if value, ok := claimString(c, jwt.RoleKey); ok && value != "" {
		return value
	}
	return ""
}

func GetRoleIDs(c *gin.Context) []int {
	roleIDs := claimInts(c, claimRoleIDs)
	if len(roleIDs) > 0 {
		return roleIDs
	}
	primaryRoleID := user.GetRoleId(c)
	if primaryRoleID > 0 {
		return []int{primaryRoleID}
	}
	return []int{}
}

func GetRoleNames(c *gin.Context) []string {
	roleNames := claimStrings(c, claimRoleNames)
	if len(roleNames) > 0 {
		return roleNames
	}
	primaryRoleName, _ := claimString(c, jwt.RoleNameKey)
	if primaryRoleName == "" {
		primaryRoleName = user.GetRoleName(c)
	}
	if primaryRoleName != "" {
		return []string{primaryRoleName}
	}
	return []string{}
}

func GetRoleKeys(c *gin.Context) []string {
	roleKeys := claimStrings(c, claimRoleKeys)
	if len(roleKeys) > 0 {
		return roleKeys
	}
	primaryRoleKey, _ := claimString(c, jwt.RoleKey)
	if primaryRoleKey != "" {
		return []string{primaryRoleKey}
	}
	return []string{}
}

func claimInt(c *gin.Context, key string) (int, bool) {
	if value, ok := claimValue(c, key); ok {
		return toInt(value)
	}
	return 0, false
}

func claimString(c *gin.Context, key string) (string, bool) {
	if value, ok := claimValue(c, key); ok {
		return toString(value)
	}
	return "", false
}

func claimInts(c *gin.Context, key string) []int {
	if value, ok := claimValue(c, key); ok {
		return toIntSlice(value)
	}
	return []int{}
}

func claimStrings(c *gin.Context, key string) []string {
	if value, ok := claimValue(c, key); ok {
		return toStringSlice(value)
	}
	return []string{}
}

func claimValue(c *gin.Context, key string) (interface{}, bool) {
	if value, ok := c.Get(key); ok {
		return value, true
	}
	claims := jwt.ExtractClaims(c)
	value, ok := claims[key]
	return value, ok
}

func toInt(value interface{}) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case float32:
		return int(v), true
	case float64:
		return int(v), true
	case string:
		if v == "" {
			return 0, false
		}
		n, err := strconv.Atoi(v)
		if err != nil {
			return 0, false
		}
		return n, true
	default:
		return 0, false
	}
}

func toString(value interface{}) (string, bool) {
	switch v := value.(type) {
	case string:
		return v, true
	default:
		text := strings.TrimSpace(fmt.Sprint(v))
		if text == "" || text == "<nil>" {
			return "", false
		}
		return text, true
	}
}

func toIntSlice(value interface{}) []int {
	switch v := value.(type) {
	case []int:
		return uniqueInts(v)
	case []int32:
		items := make([]int, 0, len(v))
		for _, item := range v {
			items = append(items, int(item))
		}
		return uniqueInts(items)
	case []int64:
		items := make([]int, 0, len(v))
		for _, item := range v {
			items = append(items, int(item))
		}
		return uniqueInts(items)
	case []float64:
		items := make([]int, 0, len(v))
		for _, item := range v {
			items = append(items, int(item))
		}
		return uniqueInts(items)
	case []interface{}:
		items := make([]int, 0, len(v))
		for _, item := range v {
			if number, ok := toInt(item); ok {
				items = append(items, number)
			}
		}
		return uniqueInts(items)
	case string:
		if strings.TrimSpace(v) == "" {
			return []int{}
		}
		parts := strings.Split(v, ",")
		items := make([]int, 0, len(parts))
		for _, part := range parts {
			if number, ok := toInt(strings.TrimSpace(part)); ok {
				items = append(items, number)
			}
		}
		return uniqueInts(items)
	default:
		if number, ok := toInt(value); ok {
			return []int{number}
		}
		return []int{}
	}
}

func toStringSlice(value interface{}) []string {
	switch v := value.(type) {
	case []string:
		return uniqueStrings(v)
	case []interface{}:
		items := make([]string, 0, len(v))
		for _, item := range v {
			if text, ok := toString(item); ok && text != "" {
				items = append(items, text)
			}
		}
		return uniqueStrings(items)
	case string:
		if strings.TrimSpace(v) == "" {
			return []string{}
		}
		parts := strings.Split(v, ",")
		items := make([]string, 0, len(parts))
		for _, part := range parts {
			text := strings.TrimSpace(part)
			if text != "" {
				items = append(items, text)
			}
		}
		return uniqueStrings(items)
	default:
		if text, ok := toString(value); ok && text != "" {
			return []string{text}
		}
		return []string{}
	}
}

func uniqueInts(values []int) []int {
	if len(values) == 0 {
		return []int{}
	}
	seen := make(map[int]struct{}, len(values))
	result := make([]int, 0, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func uniqueStrings(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		text := strings.TrimSpace(value)
		if text == "" {
			continue
		}
		if _, ok := seen[text]; ok {
			continue
		}
		seen[text] = struct{}{}
		result = append(result, text)
	}
	return result
}
