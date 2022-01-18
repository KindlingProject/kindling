package mysql

import (
	"strings"

	"github.com/Kindling-project/kindling/collector/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/analyzer/network/protocol/mysql/tools"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
)

/*
int<3>	payload_length
int<1>	sequence_id
payload
*/
func fastfailMysqlRequest() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return len(message.Data) < 5
	}
}

func parseMysqlRequest() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		return true, false
	}
}

/*
===== PayLoad =====
1              COM_STMT_PREPARE<0x16>
string[EOF]    the query to prepare
*/
func fastfailMysqlPrepare() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return message.Data[4] != 22
	}
}

func parseMysqlPrepare() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		sql := string(message.Data[5:])
		if !isSql(sql) {
			return false, true
		}
		message.AddStringAttribute(constlabels.Sql, sql)
		message.AddStringAttribute(constlabels.ContentKey, tools.SQL_MERGER.ParseStatement(sql))
		return true, true
	}
}

/*
===== PayLoad =====
1              COM_QUERY<03>
string[EOF]    the query the server shall execute
*/
func fastfailMysqlQuery() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return message.Data[4] != 3
	}
}

func parseMysqlQuery() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		sql := string(message.Data[5:])
		if !isSql(sql) {
			return false, true
		}

		message.AddStringAttribute(constlabels.Sql, sql)
		message.AddStringAttribute(constlabels.ContentKey, tools.SQL_MERGER.ParseStatement(sql))
		return true, true
	}
}

var sqlPrefixs = []string{
	"select",
	"insert",
	"update",
	"delete",
	"drop",
	"create",
	"alter",
}

func isSql(sql string) bool {
	lowerSql := strings.ToLower(sql)

	for _, prefix := range sqlPrefixs {
		if strings.HasPrefix(lowerSql, prefix) {
			return true
		}
	}
	return false
}
