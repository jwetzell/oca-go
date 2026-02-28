package oca_test

import (
	"reflect"
	"slices"
	"testing"

	"github.com/jwetzell/oca-go"
)

func TestGoodUnmarshalOCA(t *testing.T) {
	testCases := []struct {
		description string
		bytes       []byte
		expected    oca.Ocp1MessagePdu
	}{
		{
			description: "keep alive",
			bytes:       []byte{0x3b, 0x00, 0x01, 0x00, 0x00, 0x00, 0x0b, 0x04, 0x00, 0x01, 0x00, 0x01},
			expected: oca.Ocp1MessagePdu{
				SyncVal: 0x3b,
				Header: oca.Ocp1Header{
					ProtocolVersion: 1,
					PduSize:         11,
					PduType:         4,
					MessageCount:    1,
				},
				Data: &oca.Ocp1KeepAliveData{
					HeartbeatTimeout: 1,
					Seconds:          true,
				},
			},
		},
		{
			description: "command - response required",
			bytes:       []byte{0x3b, 0x00, 0x01, 0x00, 0x00, 0x00, 0x2f, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x26, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04, 0x00, 0x03, 0x00, 0x01, 0x05, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x04, 0x1f, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x01, 0x00, 0x00},
			expected: oca.Ocp1MessagePdu{
				SyncVal: 0x3b,
				Header: oca.Ocp1Header{
					ProtocolVersion: 1,
					PduSize:         47,
					PduType:         1,
					MessageCount:    1,
				},
				Data: &oca.Ocp1CommandData{
					{
						CommandSize: 38,
						Handle:      0,
						TargetONo:   4,
						MethodID: oca.OcaMethodID{
							DefLevel:    3,
							MethodIndex: 1,
						},
						Parameters: oca.Ocp1Parameters{
							ParameterCount: 5,
							Bytes:          []byte{0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x04, 0x1f, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x01, 0x00, 0x00},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			var pdu oca.Ocp1MessagePdu
			err := pdu.UnmarshalBinary(tc.bytes)
			if err != nil {
				t.Fatalf("failed to unmarshal: %s", err)
			}
			if !reflect.DeepEqual(pdu, tc.expected) {
				t.Errorf("expected %+v, got %+v", tc.expected.Data, pdu.Data)
			}
		})
	}
}

func TestGoodMarshalOCA(t *testing.T) {
	testCases := []struct {
		description string
		expected    []byte
		message     oca.Ocp1MessagePdu
	}{
		{
			description: "keep alive",
			expected:    []byte{0x3b, 0x00, 0x01, 0x00, 0x00, 0x00, 0x0b, 0x04, 0x00, 0x01, 0x00, 0x01},
			message: oca.Ocp1MessagePdu{
				SyncVal: 0x3b,
				Header: oca.Ocp1Header{
					ProtocolVersion: 1,
					PduSize:         11,
					PduType:         4,
					MessageCount:    1,
				},
				Data: &oca.Ocp1KeepAliveData{
					HeartbeatTimeout: 1,
					Seconds:          true,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			bytes, err := tc.message.MarshalBinary()
			if err != nil {
				t.Fatalf("failed to marshal: %s", err)
			}
			if !slices.Equal(bytes, tc.expected) {
				t.Errorf("expected %x, got %x", tc.expected, bytes)
			}
		})
	}
}
