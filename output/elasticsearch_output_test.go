package output

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestGetUserPasswordAndHost(t *testing.T) {
	var url string
	url = "http://admin:pw@127.0.0.1:9200/"
	scheme, user, password, host := getUserPasswordAndHost(url)
	assert.Equal(t, scheme, "http")
	assert.Equal(t, user, "admin")
	assert.Equal(t, password, "pw")
	assert.Equal(t, host, "127.0.0.1:9200")

	url = "https://admin:pw@127.0.0.1:9200/"
	scheme, user, password, host = getUserPasswordAndHost(url)
	assert.Equal(t, scheme, "https")
	assert.Equal(t, user, "admin")
	assert.Equal(t, password, "pw")
	assert.Equal(t, host, "127.0.0.1:9200")

	url = "http://127.0.0.1:9200/"
	scheme, user, password, host = getUserPasswordAndHost(url)
	assert.Equal(t, scheme, "http")
	assert.Equal(t, user, "")
	assert.Equal(t, password, "")
	assert.Equal(t, host, "127.0.0.1:9200")

	url = "https://127.0.0.1:9200/"
	scheme, user, password, host = getUserPasswordAndHost(url)
	assert.Equal(t, scheme, "https")
	assert.Equal(t, user, "")
	assert.Equal(t, password, "")
	assert.Equal(t, host, "127.0.0.1:9200")
}

func TestFilterNodesIPList(t *testing.T) {
	var resp string = `{
  "_nodes" : {
    "total" : 1,
    "successful" : 1,
    "failed" : 0
  },
  "cluster_name" : "docker-cluster",
  "nodes" : {
    "-uN_iMTNRtOP8vcTeJU2cw" : {
      "name" : "6741267939d8",
      "transport_address" : "10.0.108.2:9300",
      "host" : "10.0.108.2",
      "ip" : "10.0.108.2",
      "version" : "7.10.1",
      "build_flavor" : "default",
      "build_type" : "docker",
      "build_hash" : "1c34507e66d7db1211f66f3513706fdf548736aa",
      "roles" : [
        "data",
        "data_cold",
        "data_content",
        "data_hot",
        "data_warm",
        "ingest",
        "master",
        "ml",
        "remote_cluster_client",
        "transform"
      ],
      "attributes" : {
        "ml.machine_memory" : "24670289920",
        "xpack.installed" : "true",
        "transform.node" : "true",
        "ml.max_open_jobs" : "20"
      },
      "http" : {
        "bound_address" : [
          "0.0.0.0:9200"
        ],
        "publish_address" : "10.0.108.2:9200",
        "max_content_length_in_bytes" : 104857600
      }
    }
  }
}`
	var v map[string]interface{}
	json.Unmarshal([]byte(resp), &v)

	for _, c := range []struct {
		match string
		want  []string
	}{
		{
			match: "",
			want:  []string{"10.0.108.2:9200"},
		},
		{
			match: `EQ($.roles[0],"data")`,
			want:  []string{"10.0.108.2:9200"},
		},
		{
			match: `EQ($.roles[0],"master")`,
			want:  []string{},
		},
	} {
		got, err := filterNodesIPList(v, c.match)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(c.want, got) {
			t.Errorf("match: %s, want %v, got %v", c.match, c.want, got)
		}
	}

	resp = `{
  "_nodes" : {
    "total" : 1,
    "successful" : 1,
    "failed" : 0
  },
  "cluster_name" : "docker-cluster",
  "nodes" : {
    "-uN_iMTNRtOP8vcTeJU2cw" : {
      "name" : "6741267939d8",
      "transport_address" : "10.0.108.2:9300",
      "host" : "10.0.108.2",
      "ip" : "10.0.108.2",
      "version" : "7.10.1",
      "build_flavor" : "default",
      "build_type" : "docker",
      "build_hash" : "1c34507e66d7db1211f66f3513706fdf548736aa",
      "roles" : [
        "data",
        "data_cold",
        "data_content",
        "data_hot",
        "data_warm",
        "ingest",
        "master",
        "ml",
        "remote_cluster_client",
        "transform"
      ],
      "attributes" : {
        "ml.machine_memory" : "24670289920",
        "xpack.installed" : "true",
        "transform.node" : "true",
        "ml.max_open_jobs" : "20"
      },
      "http" : {
        "bound_address" : [
          "0.0.0.0:9200"
        ],
        "max_content_length_in_bytes" : 104857600
      }
    }
  }
}`
	json.Unmarshal([]byte(resp), &v)

	for _, c := range []struct {
		match string
		want  []string
	}{
		{
			match: "",
			want:  nil,
		},
		{
			match: `EQ($.roles[0],"data")`,
			want:  nil,
		},
	} {
		got, err := filterNodesIPList(v, c.match)
		if err == nil {
			t.Errorf("publish_address not exists, should return error")
		}
		if !reflect.DeepEqual(c.want, got) {
			t.Errorf("match: %s, want %v, got %v", c.match, c.want, got)
		}
	}

	resp = `{
  "_nodes" : {
    "total" : 1,
    "successful" : 1,
    "failed" : 0
  },
  "cluster_name" : "docker-cluster",
  "nodes" : {
    "-uN_iMTNRtOP8vcTeJU2cw" : {
      "name" : "6741267939d8",
      "transport_address" : "10.0.108.2:9300",
      "host" : "10.0.108.2",
      "ip" : "10.0.108.2",
      "version" : "7.10.1",
      "build_flavor" : "default",
      "build_type" : "docker",
      "build_hash" : "1c34507e66d7db1211f66f3513706fdf548736aa",
      "attributes" : {
        "ml.machine_memory" : "24670289920",
        "xpack.installed" : "true",
        "transform.node" : "true",
        "ml.max_open_jobs" : "20"
      },
      "http" : {
        "bound_address" : [
          "0.0.0.0:9200"
        ],
        "publish_address" : "10.0.108.2:9200",
        "max_content_length_in_bytes" : 104857600
      }
    }
  }
}`
	json.Unmarshal([]byte(resp), &v)

	for _, c := range []struct {
		match string
		want  []string
	}{
		{
			match: "",
			want:  []string{"10.0.108.2:9200"},
		},
		{
			match: `EQ($.roles[0],"data")`,
			want:  []string{},
		},
	} {
		got, err := filterNodesIPList(v, c.match)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(c.want, got) {
			t.Errorf("match: %s, want %v, got %v", c.match, c.want, got)
		}
	}
}
