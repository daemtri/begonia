package gate

import (
	"encoding/binary"
	"errors"
	"io"
)

var (
	ErrorWrongLength        = errors.New("ErrorWrongLength")
	LengthSize       uint32 = 4
	MessageIdSize    uint32 = 4
	UserIdSize       uint32 = 8
)

// FrameReader
//
//	@Description: 处理一帧数据
//	|--len(4)--|---data----|
//	len包含头，是整个frame的大小
type FrameReader interface {
	//
	// ReadOneFrame
	//  @Description: 从reader中读取完整的一帧数据，会阻塞
	//  @param ctx 上下文
	//  @param reader 通常是conn
	//  @return frame 完整的一帧数据
	//  @return ok 是否成功
	//  @return err 失败时，返回错误信息
	//
	ReadOneFrame(reader io.Reader) (frame []byte, ok bool, err error)
}

// MessageFrameParser
//
//	@Description: 处理一帧数据
//	|--len(4)--|--messageId--|---data----|
//	len包含头，是整个frame的大小
type MessageFrameParser interface {
	FrameReader

	//
	// Parse
	//  @Description: 解析帧
	//  @param frame 完整的一帧数据
	//  @return messageId 解析出来的消息id
	//  @return dataSlice 解析出来的数据部分
	//  @return ok 是否成功
	//
	Parse(frame []byte) (messageId int32, dataSlice []byte, ok bool)

	//
	// Wrap
	//  @Description: 将消息id和buffer合并成一个完整一帧数据
	//  @param messageId 消息id
	//  @param buffer  数据
	//  @return frame 完整的一帧数据
	//
	Wrap(messageId int32, buffer []byte) (frame []byte)
}

func NewMessageFrameParser(maxFrameSize uint32) MessageFrameParser {
	return &messageFrameParserImp{
		maxFrameSize: maxFrameSize,
	}
}

type messageFrameParserImp struct {
	MessageFrameParser
	maxFrameSize uint32
}

func (processor *messageFrameParserImp) ReadOneFrame(reader io.Reader) (frame []byte, ok bool, err error) {
	// 读取长度
	headData := make([]byte, LengthSize)
	if _, err := io.ReadFull(reader, headData); err != nil {
		return nil, false, err
	}

	// 解析数据
	length := processor.unpackHeader(headData)

	if length >= processor.maxFrameSize || length < MessageIdSize+LengthSize {
		// logger.Debug("processor length exceed max")
		return nil, false, ErrorWrongLength
	}

	// frame是一帧数据，包括前面的长度
	frame = processor.allocBuffer(length)
	copy(frame, headData)
	data := frame[LengthSize:]

	if _, err := io.ReadFull(reader, data); err != nil {
		return nil, false, err
	}

	return frame, true, nil
}

func (processor *messageFrameParserImp) unpackHeader(buffer []byte) uint32 {
	return binary.BigEndian.Uint32(buffer)
}

func (processor *messageFrameParserImp) allocBuffer(size uint32) []byte {
	return make([]byte, size)
}

func (processor *messageFrameParserImp) Parse(frame []byte) (messageId int32, dataSlice []byte, ok bool) {
	if uint32(len(frame)) < LengthSize+MessageIdSize {
		return 0, nil, false
	}
	messageId = int32(binary.BigEndian.Uint32(frame[LengthSize:]))
	dataSlice = frame[LengthSize+MessageIdSize:]
	return messageId, dataSlice, true
}

func (processor *messageFrameParserImp) Wrap(messageId int32, buffer []byte) (frame []byte) {
	frameSize := uint32(len(buffer)) + LengthSize + MessageIdSize
	frame = processor.allocBuffer(frameSize)
	binary.BigEndian.PutUint32(frame, frameSize)
	binary.BigEndian.PutUint32(frame[LengthSize:], uint32(messageId))
	if len(buffer) > 0 {
		copy(frame[LengthSize+MessageIdSize:], buffer)
	}
	return frame
}

// RouterFrameParser
//
//	@Description: 处理一帧数据
//	|--len(4)--|--userId--|---data----|
//	len包含头，是整个frame的大小
type RouterFrameParser interface {
	FrameReader

	//
	// Parse
	//  @Description: 解析帧
	//  @param frame 完整的一帧数据
	//  @return userId 解析出来的用户id
	//  @return dataSlice 解析出来的数据部分
	//  @return ok 是否成功
	//
	Parse(frame []byte) (userId int64, dataSlice []byte, ok bool)

	//
	// Wrap
	//  @Description: 将用户id和buffer合并成一个完整一帧数据
	//  @param userId 用户id
	//  @param buffer  数据
	//  @return frame 完整的一帧数据
	//
	Wrap(userId int64, buffer []byte) (frame []byte)

	CreateAuthorFrame(appId ServerType, clusterId uint8) (frame []byte)

	DecodeAuthorFrame(frame []byte) (appId ServerType, clusterId uint8, ok bool)
}

func NewRouterFrameParser(maxFrameSize uint32) RouterFrameParser {
	return &routerFrameParserImp{
		maxFrameSize: maxFrameSize,
	}
}

type routerFrameParserImp struct {
	MessageFrameParser
	maxFrameSize uint32
}

func (processor *routerFrameParserImp) ReadOneFrame(reader io.Reader) (frame []byte, ok bool, err error) {
	// 读取长度
	headData := make([]byte, LengthSize)
	if _, err := io.ReadFull(reader, headData); err != nil {
		return nil, false, err
	}

	// 解析数据
	length := processor.unpackHeader(headData)

	if length >= processor.maxFrameSize || length < UserIdSize+LengthSize {
		// logger.Debug("processor length exceed max")
		return nil, false, ErrorWrongLength
	}

	// frame是一帧数据，包括前面的长度
	frame = processor.allocBuffer(length)
	copy(frame, headData)
	data := frame[LengthSize:]

	if _, err := io.ReadFull(reader, data); err != nil {
		return nil, false, err
	}

	return frame, true, nil
}

func (processor *routerFrameParserImp) unpackHeader(buffer []byte) uint32 {
	return binary.BigEndian.Uint32(buffer)
}

func (processor *routerFrameParserImp) allocBuffer(size uint32) []byte {
	return make([]byte, size)
}

func (processor *routerFrameParserImp) Parse(frame []byte) (userId int64, dataSlice []byte, ok bool) {
	if uint32(len(frame)) < LengthSize+UserIdSize {
		return 0, nil, false
	}
	userId = int64(binary.BigEndian.Uint64(frame[LengthSize:]))
	dataSlice = frame[LengthSize+UserIdSize:]
	ok = true
	return
}

func (processor *routerFrameParserImp) Wrap(userId int64, buffer []byte) (frame []byte) {
	frameSize := uint32(len(buffer)) + LengthSize + UserIdSize
	frame = processor.allocBuffer(frameSize)
	binary.BigEndian.PutUint32(frame, frameSize)
	binary.BigEndian.PutUint64(frame[LengthSize:], uint64(userId))
	if len(buffer) > 0 {
		copy(frame[LengthSize+UserIdSize:], buffer)
	}
	return frame
}

func (processor *routerFrameParserImp) CreateAuthorFrame(appId ServerType, clusterId uint8) (frame []byte) {
	buffer := make([]byte, 2)
	buffer[0] = uint8(appId)
	buffer[1] = clusterId
	return processor.Wrap(0, buffer)
}

func (processor *routerFrameParserImp) DecodeAuthorFrame(frame []byte) (appId ServerType, clusterId uint8, ok bool) {
	userId, buffer, ok := processor.Parse(frame)
	if !ok || userId != 0 || len(buffer) != 2 {
		return 0, 0, false
	}
	appId = ServerType(buffer[0])
	clusterId = buffer[1]
	if !IsCluster(appId) {
		return 0, 0, false
	}
	return appId, clusterId, ok
}
