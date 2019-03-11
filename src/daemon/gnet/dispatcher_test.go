package gnet

import (
	"bytes"
	"errors"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	_sendByteMessage = sendByteMessage
)

func resetHandler() {
	sendByteMessage = _sendByteMessage
}

func TestMsgIDStringSafe(t *testing.T) {
	var id [4]byte
	require.Equal(t, "\\x00\\x00\\x00\\x00", msgIDStringSafe(id))

	id = [4]byte{'F', 'O', 'O', 'B'}

	require.Equal(t, "FOOB", msgIDStringSafe(id))

	id = [4]byte{200, 2, '\n', '\t'}

	require.Equal(t, "\\xc8\\x02\\n\\t", msgIDStringSafe(id))

	id = [4]byte{'\'', '\\', ' ', '"'}

	require.Equal(t, "'\\\\ \\\"", msgIDStringSafe(id))
}

func TestConvertToMessage(t *testing.T) {
	EraseMessages()
	resetHandler()
	RegisterMessage(BytePrefix, ByteMessage{})
	VerifyMessages()
	c := &Connection{}
	b := make([]byte, 0)
	b = append(b, BytePrefix[:]...)
	b = append(b, byte(7))
	m, err := convertToMessage(c.ID, b, testing.Verbose())
	require.NoError(t, err)
	require.NotNil(t, m)
	if m == nil {
		t.Fatalf("ConvertToMessage failed")
	}
	bm := m.(*ByteMessage)
	require.Equal(t, bm.X, byte(7))
}

func TestConvertToMessageNoMessageID(t *testing.T) {
	EraseMessages()
	resetHandler()
	c := &Connection{}
	b := []byte{}
	m, err := convertToMessage(c.ID, b, testing.Verbose())
	require.Nil(t, m)
	require.Error(t, err)
	require.Equal(t, ErrDisconnectTruncatedMessageID, err)
}

func TestConvertToMessageUnknownMessage(t *testing.T) {
	EraseMessages()
	resetHandler()
	c := &Connection{}
	b := MessagePrefix{'C', 'C', 'C', 'C'}
	m, err := convertToMessage(c.ID, b[:], testing.Verbose())
	require.Error(t, err)
	require.Equal(t, ErrDisconnectUnknownMessage, err)
	require.Nil(t, m)
}

func TestConvertToMessageBadDeserialize(t *testing.T) {
	EraseMessages()
	resetHandler()
	RegisterMessage(DummyPrefix, DummyMessage{})
	RegisterMessage(BytePrefix, ByteMessage{})
	VerifyMessages()
	c := &Connection{}
	// Test with too many bytes
	b := append(DummyPrefix[:], []byte{0, 1, 1, 1}...)
	m, err := convertToMessage(c.ID, b, testing.Verbose())
	require.Error(t, err)
	require.Nil(t, m)

	// Test with not enough bytes
	b = append([]byte{}, BytePrefix[:]...)
	m, err = convertToMessage(c.ID, b, testing.Verbose())
	require.Error(t, err)
	require.Equal(t, ErrDisconnectMalformedMessage, err)
	require.Nil(t, m)
}

func TestConvertToMessageNotMessage(t *testing.T) {
	EraseMessages()
	resetHandler()
	RegisterMessage(NothingPrefix, Nothing{})
	// don't verify messages
	c := &Connection{}
	require.Panics(t, func() {
		_, _ = convertToMessage(c.ID, NothingPrefix[:], testing.Verbose()) // nolint: errcheck
	})
}

func TestDeserializeMessageTrapsPanic(t *testing.T) {
	resetHandler()
	EraseMessages()
	p := 7
	m := PointerMessage{Ptr: &p}
	b := []byte{4, 4, 4, 4, 4, 4, 4, 4}
	_, err := deserializeMessage(b, reflect.ValueOf(m))
	require.Error(t, err)
	require.Equal(t, err.Error(), "DeserializeRawToValue value must be a ptr, is struct")
}

func TestEncodeMessage(t *testing.T) {
	resetHandler()
	EraseMessages()
	RegisterMessage(BytePrefix, ByteMessage{})
	VerifyMessages()
	m := NewByteMessage(7)
	b := EncodeMessage(m)
	require.True(t, bytes.Equal(b, []byte{5, 0, 0, 0, 'B', 'Y', 'T', 'E', 7}))
}

func TestEncodeMessageUnknownMessage(t *testing.T) {
	resetHandler()
	EraseMessages()
	require.Panics(t, func() { EncodeMessage(&DummyMessage{}) })
}

func TestSendByteMessage(t *testing.T) {
	resetHandler()
	b := []byte{1}
	c := NewCaptureConn()
	err := sendByteMessage(c, b, 0)
	require.NoError(t, err)
	require.True(t, bytes.Equal(c.(*CaptureConn).Wrote, b))
	require.True(t, c.(*CaptureConn).WriteDeadlineSet)
}

func TestSendByteMessageWithTimeout(t *testing.T) {
	resetHandler()
	b := []byte{1}
	c := NewCaptureConn()
	err := sendByteMessage(c, b, time.Minute)
	require.NoError(t, err)
	require.True(t, bytes.Equal(c.(*CaptureConn).Wrote, b))
	require.True(t, c.(*CaptureConn).WriteDeadlineSet)
}

func TestSendByteMessageWriteFailed(t *testing.T) {
	resetHandler()
	c := &FailingWriteConn{}
	err := sendByteMessage(c, nil, 0)
	require.Error(t, err)
}

func TestSendByteMessageWriteDeadlineFailed(t *testing.T) {
	resetHandler()
	c := &FailingWriteDeadlineConn{}
	err := sendByteMessage(c, nil, 0)
	require.Error(t, err)
}

func TestSendMessage(t *testing.T) {
	resetHandler()
	EraseMessages()
	RegisterMessage(BytePrefix, ByteMessage{})
	VerifyMessages()
	m := NewByteMessage(7)
	sendByteMessage = func(conn net.Conn, msg []byte, tm time.Duration) error {
		expect := []byte{5, 0, 0, 0, 'B', 'Y', 'T', 'E', 7}
		require.True(t, bytes.Equal(msg, expect))
		return nil
	}
	err := sendMessage(nil, m, 0)
	require.NoError(t, err)
}

/* Helpers */

func failingSendByteMessage(conn net.Conn, m []byte, tm time.Duration) error {
	return errors.New("send byte message failed")
}

type CaptureConn struct {
	Wrote            []byte
	WriteDeadlineSet bool
	net.Conn
}

func NewCaptureConn() net.Conn {
	return &CaptureConn{Wrote: nil, WriteDeadlineSet: false}
}

func (cc *CaptureConn) Write(b []byte) (int, error) {
	cc.Wrote = b
	return len(b), nil
}

func (cc *CaptureConn) SetWriteDeadline(t time.Time) error {
	cc.WriteDeadlineSet = true
	return nil
}

type FailingWriteDeadlineConn struct {
	net.Conn
}

func (c *FailingWriteDeadlineConn) SetWriteDeadline(t time.Time) error {
	return errors.New("failed")
}

type FailingWriteConn struct {
	net.Conn
}

func (c *FailingWriteConn) Write(b []byte) (int, error) {
	return 0, errors.New("failed")
}

func (c *FailingWriteConn) SetWriteDeadline(t time.Time) error {
	return nil
}
