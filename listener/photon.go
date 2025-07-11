package listener

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/google/gopacket"
)

const (
	// commands
	acknowledgeType          = 1
	connectType              = 2
	verifyConnectType        = 3
	disconnectType           = 4
	pingType                 = 5
	sendReliableType         = 6
	sendUnreliableType       = 7
	sendReliableFragmentType = 8
	// Message types
	operationRequest       = 2
	otherOperationResponse = 3
	eventDataType          = 4
	operationResponse      = 7
)

var photonCommandHeaderLength = 12

type photonCommand struct {
	// Header
	commandType            uint8
	channelID              uint8
	flags                  uint8
	reservedByte           uint8
	length                 int32
	reliableSequenceNumber int32

	// Body
	data []byte
}

type reliableMessage struct {
	// Header
	signature   uint8
	messageType uint8

	// operationRequest
	operationCode uint8

	// EventData
	eventCode uint8

	// operationResponse
	operationResponseCode uint16
	operationDebugString  string

	parameterCount int16
	data           []byte
}

type reliableFragment struct {
	sequenceNumber int32
	fragmentCount  int32
	fragmentNumber int32
	totalLength    int32
	fragmentOffset int32

	data []byte
}

type photonLayer struct {
	// Header
	peerID       uint16
	crcEnabled   uint8
	commandCount uint8
	timestamp    uint32
	challenge    int32

	// commands
	commands []photonCommand

	// Interface stuff
	contents []byte
	payload  []byte
}

var photonLayerType = gopacket.RegisterLayerType(
	5056,
	gopacket.LayerTypeMetadata{
		Name:    "photonLayerType",
		Decoder: gopacket.DecodeFunc(decodePhotonPacket)})

func (p photonLayer) LayerType() gopacket.LayerType { return photonLayerType }
func (p photonLayer) LayerContents() []byte         { return p.contents }
func (p photonLayer) LayerPayload() []byte          { return p.payload }

func decodePhotonPacket(data []byte, p gopacket.PacketBuilder) error {
	layer := photonLayer{}
	buf := bytes.NewBuffer(data)

	// Header
	if err := binary.Read(buf, binary.BigEndian, &layer.peerID); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.BigEndian, &layer.crcEnabled); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.BigEndian, &layer.commandCount); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.BigEndian, &layer.timestamp); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.BigEndian, &layer.challenge); err != nil {
		return err
	}

	// commands
	var commands []photonCommand
	for i := 0; i < int(layer.commandCount); i++ {
		var command photonCommand

		// Command header
		if err := binary.Read(buf, binary.BigEndian, &command.commandType); err != nil {
			return err
		}
		if err := binary.Read(buf, binary.BigEndian, &command.channelID); err != nil {
			return err
		}
		if err := binary.Read(buf, binary.BigEndian, &command.flags); err != nil {
			return err
		}
		if err := binary.Read(buf, binary.BigEndian, &command.reservedByte); err != nil {
			return err
		}
		if err := binary.Read(buf, binary.BigEndian, &command.length); err != nil {
			return err
		}
		if err := binary.Read(buf, binary.BigEndian, &command.reliableSequenceNumber); err != nil {
			return err
		}

		// Command data
		dataLength := int(command.length) - photonCommandHeaderLength
		if dataLength > buf.Len() {
			return fmt.Errorf("dataLength %d is greater than buffer length %d", dataLength, buf.Len())
		}

		command.data = make([]byte, dataLength)
		if _, err := buf.Read(command.data); err != nil {
			return err
		}

		commands = append(commands, command)
	}

	layer.commands = commands

	// Split and store the read and unread data
	dataUsed := len(data) - buf.Len()
	layer.contents = data[0:dataUsed]
	layer.payload = buf.Bytes()

	p.AddLayer(layer)
	return p.NextDecoder(gopacket.LayerTypePayload)
}

func (c *photonCommand) reliableMessage() (msg reliableMessage, err error) {
	if c.commandType != sendReliableType {
		return msg, fmt.Errorf("command can't be converted")
	}

	buf := bytes.NewBuffer(c.data)

	if err = binary.Read(buf, binary.BigEndian, &msg.signature); err != nil {
		return msg, err
	}
	if err = binary.Read(buf, binary.BigEndian, &msg.messageType); err != nil {
		return msg, err
	}

	if msg.messageType > 128 {
		return msg, fmt.Errorf("encryption not supported")
	}

	if msg.messageType == otherOperationResponse {
		msg.messageType = operationResponse
	}

	switch msg.messageType {
	case operationRequest:
		if err = binary.Read(buf, binary.BigEndian, &msg.operationCode); err != nil {
			return msg, err
		}
	case eventDataType:
		if err = binary.Read(buf, binary.BigEndian, &msg.eventCode); err != nil {
			return msg, err
		}
	case operationResponse:
		if err = binary.Read(buf, binary.BigEndian, &msg.operationCode); err != nil {
			return msg, err
		}
		if err = binary.Read(buf, binary.BigEndian, &msg.operationResponseCode); err != nil {
			return msg, err
		}

		var paramType uint8
		if err = binary.Read(buf, binary.BigEndian, &paramType); err != nil {
			return msg, err
		}

		paramValue, err := decodeType(buf, paramType)
		if err != nil {
			return msg, err
		}

		if paramValue != nil {
			msg.operationDebugString = paramValue.(string)
		}
	}

	if err = binary.Read(buf, binary.BigEndian, &msg.parameterCount); err != nil {
		return msg, err
	}

	msg.data = buf.Bytes()
	return
}

func (c *photonCommand) reliableFragment() (msg reliableFragment, err error) {
	if c.commandType != sendReliableFragmentType {
		return msg, fmt.Errorf("command can't be converted")
	}

	buf := bytes.NewBuffer(c.data)

	if err = binary.Read(buf, binary.BigEndian, &msg.sequenceNumber); err != nil {
		return msg, err
	}
	if err = binary.Read(buf, binary.BigEndian, &msg.fragmentCount); err != nil {
		return msg, err
	}
	if err = binary.Read(buf, binary.BigEndian, &msg.fragmentNumber); err != nil {
		return msg, err
	}
	if err = binary.Read(buf, binary.BigEndian, &msg.totalLength); err != nil {
		return msg, err
	}
	if err = binary.Read(buf, binary.BigEndian, &msg.fragmentOffset); err != nil {
		return msg, err
	}

	msg.data = buf.Bytes()
	return
}
