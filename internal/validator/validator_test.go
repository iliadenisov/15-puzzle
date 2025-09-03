package validator_test

import (
	"15-puzzle/internal/validator"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	botToken = "example:token"
	initData = "auth_date=269666017&chat_instance=6039284203686499081&chat_type=sender&hash=40cde8dc7250ee616cd7d7a090749a9a42cc68f018c969fcb48fdf5e62657ad6&signature=FF5oTJSnmxdqgtNozxsLywXyVKdssh_DbvksGUaQuhkMiRfp10HJmf5o88uokPpqF4yhpHbX1c8uLbrKUuUdAA&user=%7B%22allows_write_to_pm%22%3Atrue%2C%22first_name%22%3A%22Ilia%22%2C%22id%22%3A303133707%2C%22is_premium%22%3Atrue%2C%22language_code%22%3A%22en%22%2C%22last_name%22%3A%22Denisov%22%2C%22photo_url%22%3A%22https%3A%2F%2Fyoutu.be%2FdQw4w9WgXcQ%22%7D"
	userId   = 303133707
)

func TestValidUser(t *testing.T) {
	valid, id, err := validator.ValidUser(validator.EncodeHmacSha256([]byte(botToken), []byte("WebAppData")), initData)
	assert.NoError(t, err)
	assert.True(t, valid)
	assert.Equal(t, userId, id)

	valid, _, err = validator.ValidUser(validator.EncodeHmacSha256([]byte(botToken), []byte("Web_AppData")), initData)
	assert.NoError(t, err)
	assert.False(t, valid)

	_, _, err = validator.ValidUser(validator.EncodeHmacSha256([]byte(botToken), []byte("WebAppData")), "param;=value&")
	assert.Error(t, err)
}
