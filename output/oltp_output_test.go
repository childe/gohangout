package output

import (
	"encoding/hex"
	"github.com/magiconair/properties/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"testing"
)

func TestOltpOutputEncode(t *testing.T) {
	o := newOTLPOutput(make(map[interface{}]interface{}))
	event := make(map[string]interface{})
	event["oltp_resource_service.name"] = "gohangout"
	event["oltp_scope_my.scope.attribute"] = "gohangout scope attribute"
	event["oltp_scope_name"] = "gohangout.scope"
	event["oltp_scope_version"] = "1.0.0"
	event["oltp_scope_schemaurl"] = "https://127.0.0.1/hello"
	event["oltp_logrecords_attribute1"] = "gohangout.log.attribute1"
	event["oltp_logrecords_attribute2"] = "gohangout.log.attribute2"
	event["oltp_logrecords_time"] = uint64(1544712660300000000)
	event["oltp_logrecords_severity"] = "info"
	event["oltp_logrecords_severitynumber"] = int32(9)
	event["oltp_logrecords_flags"] = uint32(0)
	event["oltp_logrecords_traceid"] = "5B8EFFF798038103D269B633813FC60C"
	event["oltp_logrecords_spanid"] = "EEE19B7EC3C1B174"
	event["oltp_logrecords_body"] = "example log body"
	request, err := o.(*OLTPOutput).oltpEncoder.Encode(event)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, request.Logs().LogRecordCount(), 1)
	resource := request.Logs().ResourceLogs().At(0)
	resAttr, _ := resource.Resource().Attributes().Get("service.name")
	assert.Equal(t, resAttr.Str(), "gohangout")
	scope := resource.ScopeLogs().At(0)
	assert.Equal(t, scope.Scope().Name(), "gohangout.scope")
	assert.Equal(t, scope.Scope().Version(), "1.0.0")
	assert.Equal(t, scope.SchemaUrl(), "https://127.0.0.1/hello")
	scopeAttr, _ := scope.Scope().Attributes().Get("my.scope.attribute")
	assert.Equal(t, scopeAttr.Str(), "gohangout scope attribute")
	logRecord := scope.LogRecords().At(0)
	assert.Equal(t, logRecord.Timestamp(), pcommon.Timestamp(1544712660300000000))
	assert.Equal(t, logRecord.SeverityText(), "info")
	assert.Equal(t, logRecord.SeverityNumber(), plog.SeverityNumber(9))
	assert.Equal(t, logRecord.Flags(), plog.LogRecordFlags(0))
	bytes, _ := hex.DecodeString("5B8EFFF798038103D269B633813FC60C")
	var traceId [16]byte
	copy(traceId[:], bytes)
	assert.Equal(t, logRecord.TraceID(), pcommon.TraceID(traceId))
	bytes, _ = hex.DecodeString("EEE19B7EC3C1B174")
	var spanId [8]byte
	copy(spanId[:], bytes)
	assert.Equal(t, logRecord.SpanID(), pcommon.SpanID(spanId))
	assert.Equal(t, logRecord.Body().Str(), "example log body")
	recordAttr1, _ := logRecord.Attributes().Get("attribute1")
	assert.Equal(t, recordAttr1.Str(), "gohangout.log.attribute1")
	recordAttr2, _ := logRecord.Attributes().Get("attribute2")
	assert.Equal(t, recordAttr2.Str(), "gohangout.log.attribute2")
	//o.Emit(event)
}
