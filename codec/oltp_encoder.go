package codec

import (
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"strings"
	"time"
)

const (
	resourcePrefix       = "oltp_resource_"
	scopePrefix          = "oltp_scope_"
	scopeName            = "oltp_scope_name"
	scopeVersion         = "oltp_scope_version"
	scopeSchemaUrl       = "oltp_scope_schemaurl"
	recordPrefix         = "oltp_logrecords_"
	recordTime           = "oltp_logrecords_time"
	recordSeverity       = "oltp_logrecords_severity"
	recordSeverityNumber = "oltp_logrecords_severitynumber"
	recordFlags          = "oltp_logrecords_flags"
	recordTraceId        = "oltp_logrecords_traceid"
	recordSpanId         = "oltp_logrecords_spanid"
	recordBody           = "oltp_logrecords_body"
)

type OltpEncoder struct{}

func (o *OltpEncoder) Encode(event map[string]interface{}) (plogotlp.ExportRequest, error) {
	logs := plog.NewLogs()
	rsLogs := plog.NewResourceLogs()
	scopeLog := rsLogs.ScopeLogs().AppendEmpty()
	logRecord := scopeLog.LogRecords().AppendEmpty()
	logRecord.SetObservedTimestamp(pcommon.Timestamp(time.Now().UnixNano()))

	for k, v := range event {
		// TODO: just support string, do no not use reflect
		switch k {
		case scopeName:
			if _, ok := v.(string); ok {
				scopeLog.Scope().SetName(v.(string))
			}
		case scopeVersion:
			if _, ok := v.(string); ok {
				scopeLog.Scope().SetVersion(v.(string))
			}
		case scopeSchemaUrl:
			if _, ok := v.(string); ok {
				scopeLog.SetSchemaUrl(v.(string))
			}
		case recordTime:
			if _, ok := v.(uint64); ok {
				logRecord.SetObservedTimestamp(pcommon.Timestamp(v.(uint64)))
			}
		case recordSeverity:
			if _, ok := v.(string); ok {
				logRecord.SetSeverityText(v.(string))
			}
		case recordSeverityNumber:
			if _, ok := v.(int32); ok {
				logRecord.SetSeverityNumber(plog.SeverityNumber(v.(int32)))
			}
		case recordFlags:
			if _, ok := v.(uint32); ok {
				logRecord.SetFlags(plog.LogRecordFlags(v.(uint32)))
			}
		case recordTraceId:
			if _, ok := v.(string); ok {
				if len(v.(string)) == 16 {
					var traceId [16]byte
					copy(traceId[:], v.(string))
					logRecord.SetTraceID(traceId)
				}
			}
		case recordSpanId:
			if _, ok := v.(string); ok {
				if len(v.(string)) == 8 {
					var spanId [8]byte
					copy(spanId[:], v.(string))
					logRecord.SetSpanID(spanId)
				}
			}
		case recordBody:
			if _, ok := v.(string); ok {
				logRecord.Body().SetStr(v.(string))
			}
		default:
			if strings.HasPrefix(k, resourcePrefix) {
				if _, ok := v.(string); ok {
					rsLogs.Resource().Attributes().PutStr(k[len(resourcePrefix):], v.(string))
				}
			} else if strings.HasPrefix(k, scopePrefix) {
				if _, ok := v.(string); ok {
					scopeLog.Scope().Attributes().PutStr(k[len(resourcePrefix):], v.(string))
				}
			} else if strings.HasPrefix(k, recordPrefix) {
				if _, ok := v.(string); ok {
					logRecord.Attributes().PutStr(k[len(recordPrefix):], v.(string))
				}
			}
		}
	}
	return plogotlp.NewExportRequestFromLogs(logs), nil
}
