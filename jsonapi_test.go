// Copyright (c) 2013 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package btcjson_test

import (
	"bytes"
	"github.com/conformal/btcjson"
	"io"
	"io/ioutil"
	"testing"
)

// cmdtests is a table of all the possible commands and a list of inputs,
// some of which should work, some of which should not (indicated by the
// pass variable).  This mainly checks the type and number of the arguments,
// it does not actually check to make sure the values are correct (i.e., that
// addresses are reasonable) as the bitcoin client must be able to deal with
// that.
var cmdtests = []struct {
	cmd  string
	args []interface{}
	pass bool
}{
	{"getinfo", nil, true},
	{"getinfo", []interface{}{1}, false},
	{"listaccounts", nil, true},
	{"listaccounts", []interface{}{1}, true},
	{"listaccounts", []interface{}{"test"}, false},
	{"listaccounts", []interface{}{1, 2}, false},
	{"getblockhash", nil, false},
	{"getblockhash", []interface{}{1}, true},
	{"getblockhash", []interface{}{1, 2}, false},
	{"getblockhash", []interface{}{1.1}, false},
	{"settxfee", nil, false},
	{"settxfee", []interface{}{1.0}, true},
	{"settxfee", []interface{}{1.0, 2.0}, false},
	{"settxfee", []interface{}{1}, false},
	{"getmemorypool", nil, true},
	{"getmemorypool", []interface{}{"test"}, true},
	{"getmemorypool", []interface{}{1}, false},
	{"getmemorypool", []interface{}{"test", 2}, false},
	{"backupwallet", nil, false},
	{"backupwallet", []interface{}{1, 2}, false},
	{"backupwallet", []interface{}{1}, false},
	{"backupwallet", []interface{}{"testpath"}, true},
	{"setaccount", nil, false},
	{"setaccount", []interface{}{1}, false},
	{"setaccount", []interface{}{1, 2, 3}, false},
	{"setaccount", []interface{}{1, "test"}, false},
	{"setaccount", []interface{}{"test", "test"}, true},
	{"verifymessage", nil, false},
	{"verifymessage", []interface{}{1}, false},
	{"verifymessage", []interface{}{1, 2}, false},
	{"verifymessage", []interface{}{1, 2, 3, 4}, false},
	{"verifymessage", []interface{}{"test", "test", "test"}, true},
	{"verifymessage", []interface{}{"test", "test", 1}, false},
	{"getaddednodeinfo", nil, false},
	{"getaddednodeinfo", []interface{}{1}, false},
	{"getaddednodeinfo", []interface{}{true}, true},
	{"getaddednodeinfo", []interface{}{true, 1}, false},
	{"getaddednodeinfo", []interface{}{true, "test"}, true},
	{"setgenerate", nil, false},
	{"setgenerate", []interface{}{1, 2, 3}, false},
	{"setgenerate", []interface{}{true}, true},
	{"setgenerate", []interface{}{true, 1}, true},
	{"setgenerate", []interface{}{true, 1.1}, false},
	{"setgenerate", []interface{}{"true", 1}, false},
	{"getbalance", nil, true},
	{"getbalance", []interface{}{"test"}, true},
	{"getbalance", []interface{}{"test", 1}, true},
	{"getbalance", []interface{}{"test", 1.0}, false},
	{"getbalance", []interface{}{1, 1}, false},
	{"getbalance", []interface{}{"test", 1, 2}, false},
	{"getbalance", []interface{}{1}, false},
	{"addnode", nil, false},
	{"addnode", []interface{}{1, 2, 3}, false},
	{"addnode", []interface{}{"test"}, true},
	{"addnode", []interface{}{1}, false},
	{"addnode", []interface{}{"test", 1}, true},
	{"addnode", []interface{}{"test", 1.0}, false},
	{"listreceivedbyaccount", nil, true},
	{"listreceivedbyaccount", []interface{}{1, 2, 3}, false},
	{"listreceivedbyaccount", []interface{}{1}, true},
	{"listreceivedbyaccount", []interface{}{1.0}, false},
	{"listreceivedbyaccount", []interface{}{1, false}, true},
	{"listreceivedbyaccount", []interface{}{1, "false"}, false},
	{"listtransactions", nil, true},
	{"listtransactions", []interface{}{"test"}, true},
	{"listtransactions", []interface{}{"test", 1}, true},
	{"listtransactions", []interface{}{"test", 1, 2}, true},
	{"listtransactions", []interface{}{"test", 1, 2, 3}, false},
	{"listtransactions", []interface{}{1}, false},
	{"listtransactions", []interface{}{"test", 1.0}, false},
	{"listtransactions", []interface{}{"test", 1, "test"}, false},
	{"importprivkey", nil, false},
	{"importprivkey", []interface{}{"test"}, true},
	{"importprivkey", []interface{}{1}, false},
	{"importprivkey", []interface{}{"test", "test"}, true},
	{"importprivkey", []interface{}{"test", "test", true}, true},
	{"importprivkey", []interface{}{"test", "test", true, 1}, false},
	{"importprivkey", []interface{}{"test", 1.0, true}, false},
	{"importprivkey", []interface{}{"test", "test", "true"}, false},
	{"listunspent", nil, true},
	{"listunspent", []interface{}{1}, true},
	{"listunspent", []interface{}{1, 2}, true},
	{"listunspent", []interface{}{1, 2, 3}, false},
	{"listunspent", []interface{}{1.0}, false},
	{"listunspent", []interface{}{1, 2.0}, false},
	{"sendfrom", nil, false},
	{"sendfrom", []interface{}{"test"}, false},
	{"sendfrom", []interface{}{"test", "test"}, false},
	{"sendfrom", []interface{}{"test", "test", 1.0}, true},
	{"sendfrom", []interface{}{"test", 1, 1.0}, false},
	{"sendfrom", []interface{}{1, "test", 1.0}, false},
	{"sendfrom", []interface{}{"test", "test", 1}, false},
	{"sendfrom", []interface{}{"test", "test", 1.0, 1}, true},
	{"sendfrom", []interface{}{"test", "test", 1.0, 1, "test"}, true},
	{"sendfrom", []interface{}{"test", "test", 1.0, 1, "test", "test"}, true},
	{"move", nil, false},
	{"move", []interface{}{1, 2, 3, 4, 5, 6}, false},
	{"move", []interface{}{1, 2}, false},
	{"move", []interface{}{"test", "test", 1.0}, true},
	{"move", []interface{}{"test", "test", 1.0, 1, "test"}, true},
	{"move", []interface{}{"test", "test", 1.0, 1}, true},
	{"move", []interface{}{1, "test", 1.0}, false},
	{"move", []interface{}{"test", 1, 1.0}, false},
	{"move", []interface{}{"test", "test", 1}, false},
	{"move", []interface{}{"test", "test", 1.0, 1.0, "test"}, false},
	{"move", []interface{}{"test", "test", 1.0, 1, true}, false},
	{"sendtoaddress", nil, false},
	{"sendtoaddress", []interface{}{"test"}, false},
	{"sendtoaddress", []interface{}{"test", 1.0}, true},
	{"sendtoaddress", []interface{}{"test", 1.0, "test"}, true},
	{"sendtoaddress", []interface{}{"test", 1.0, "test", "test"}, true},
	{"sendtoaddress", []interface{}{1, 1.0, "test", "test"}, false},
	{"sendtoaddress", []interface{}{"test", 1, "test", "test"}, false},
	{"sendtoaddress", []interface{}{"test", 1.0, 1.0, "test"}, false},
	{"sendtoaddress", []interface{}{"test", 1.0, "test", 1.0}, false},
	{"sendtoaddress", []interface{}{"test", 1.0, "test", "test", 1}, false},
	{"addmultisignaddress", []interface{}{1, "test", "test"}, true},
	{"addmultisignaddress", []interface{}{1, "test"}, false},
	{"addmultisignaddress", []interface{}{1, 1.0, "test"}, false},
	{"addmultisignaddress", []interface{}{1, "test", "test", "test"}, true},
	{"createrawtransaction", []interface{}{"in1", "out1", "a1", 1.0}, true},
	{"createrawtransaction", []interface{}{"in1", "out1", "a1", 1.0, "test"}, false},
	{"createrawtransaction", []interface{}{}, false},
	{"createrawtransaction", []interface{}{"in1", 1.0, "a1", 1.0}, false},
	{"sendmany", []interface{}{"in1", "out1", 1.0, 1, "comment"}, true},
	{"sendmany", []interface{}{"in1", "out1", 1.0, "comment"}, true},
	{"sendmany", []interface{}{"in1", "out1"}, false},
	{"sendmany", []interface{}{true, "out1", 1.0, 1, "comment"}, false},
	{"sendmany", []interface{}{"in1", "out1", "test", 1, "comment"}, false},
	{"lockunspent", []interface{}{true, "something"}, true},
	{"lockunspent", []interface{}{true}, false},
	{"lockunspent", []interface{}{1.0, "something"}, false},
	{"signrawtransaction", []interface{}{"hexstring", "test", "test2", "test3", "test4"}, true},
	{"signrawtransaction", []interface{}{"hexstring", "test", "test2", "test3"}, false},
	{"signrawtransaction", []interface{}{1.2, "test", "test2", "test3", "test4"}, false},
	{"signrawtransaction", []interface{}{"hexstring", 1, "test2", "test3", "test4"}, false},
	{"signrawtransaction", []interface{}{"hexstring", "test", 2, "test3", "test4"}, false},
	{"signrawtransaction", []interface{}{"hexstring", "test", "test2", 3, "test4"}, false},
	{"fakecommand", nil, false},
}

// TestRpcCreateMessage tests CreateMessage using the table of messages
// in cmdtests.
func TestRpcCreateMessage(t *testing.T) {
	var err error
	for i, tt := range cmdtests {
		if tt.args == nil {
			_, err = btcjson.CreateMessage(tt.cmd)
		} else {
			_, err = btcjson.CreateMessage(tt.cmd, tt.args...)
		}
		if tt.pass {
			if err != nil {
				t.Errorf("Could not create command %d: %s %v.", i, tt.cmd, err)
			}
		} else {
			if err == nil {
				t.Errorf("Should create command. %d: %s", i, tt.cmd)
			}
		}
	}
	return
}

// TestRpcCommand tests RpcCommand by generating some commands and
// trying to send them off.
func TestRpcCommand(t *testing.T) {
	user := "something"
	pass := "something"
	server := "invalid"
	var msg []byte
	_, err := btcjson.RpcCommand(user, pass, server, msg)
	if err == nil {
		t.Errorf("Should fail.")
	}
	msg, err = btcjson.CreateMessage("getinfo")
	if err != nil {
		t.Errorf("Cannot create valid json message")
	}
	_, err = btcjson.RpcCommand(user, pass, server, msg)
	if err == nil {
		t.Errorf("Should not connect to server.")
	}

	badMsg := []byte("{\"jsonrpc\":\"1.0\",\"id\":\"btcd\",\"method\":\"\"}")
	_, err = btcjson.RpcCommand(user, pass, server, badMsg)
	if err == nil {
		t.Errorf("Cannot have no method in msg..")
	}
	return
}

// FailingReadClose is a type used for testing so we can get something that
// fails past Go's type system.
type FailingReadCloser struct{}

func (f *FailingReadCloser) Close() error {
	return io.ErrUnexpectedEOF
}

func (f *FailingReadCloser) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

// TestRpcReply tests JsonGetRaw by sending both a good and a bad buffer
// to it.
func TestRpcReply(t *testing.T) {
	buffer := new(bytes.Buffer)
	buffer2 := ioutil.NopCloser(buffer)
	_, err := btcjson.GetRaw(buffer2)
	if err != nil {
		t.Errorf("Error reading rpc reply.")
	}
	failBuf := &FailingReadCloser{}
	_, err = btcjson.GetRaw(failBuf)
	if err == nil {
		t.Errorf("Error, this should fail.")
	}
	return
}
