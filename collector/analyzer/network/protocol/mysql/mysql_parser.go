package mysql

import (
	"github.com/dxsup/kindling-collector/analyzer/network/protocol"
)

/*
      Request                                         Response
       /            \                                            /     |    \
prepare   query                                err   ok  eof
*/
func NewMysqlParser() *protocol.ProtocolParser {
	requestParser := protocol.CreatePkgParser(fastfailMysqlRequest(), parseMysqlRequest())
	requestParser.Add(fastfailMysqlPrepare(), parseMysqlPrepare())
	requestParser.Add(fastfailMysqlQuery(), parseMysqlQuery())

	responseParser := protocol.CreatePkgParser(fastfailMysqlResponse(), parseMysqlResponse())
	responseParser.Add(fastfailMysqlErr(), parseMysqlErr())
	responseParser.Add(fastfailMysqlOk(), parseMysqlOk())
	responseParser.Add(fastfailMysqlEof(), parseMysqlEof())
	responseParser.Add(fastfailMysqlResultSet(), parseMysqlResultSet())

	return protocol.NewProtocolParser(protocol.MYSQL, requestParser, responseParser, nil)
}
