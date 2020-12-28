package output

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestGetUserPasswordAndHost(t *testing.T) {
	var url string
	url = "http://admin:pw@127.0.0.1:9200/"
	user, password, host := getUserPasswordAndHost(url)
	assert.Equal(t, user, "admin")
	assert.Equal(t, password, "pw")
	assert.Equal(t, host, "127.0.0.1:9200")

	url = "http://127.0.0.1:9200/"
	user, password, host = getUserPasswordAndHost(url)
	assert.Equal(t, user, "")
	assert.Equal(t, password, "")
	assert.Equal(t, host, "127.0.0.1:9200")
}
