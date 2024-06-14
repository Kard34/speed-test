package imaz

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"time"
)

func ParserHDL(buff []byte) (Hdl, error) {
	msgLen, _ := verify(buff)
	var values Field
	// var i2b = []bool{false, true}
	buff = buff[13 : 13+msgLen]
	var err error
	msg := Hdl{}
	msg.AbstractData = ""
	for len(buff) > 0 {
		values, buff, err = getField(buff)
		if err != nil {
			return msg, err
		}
		switch values.FID {

		case FID_Proximity:
			msg.Proximity = values.Data.(uint8)

		case FID_NewsID:
			msg.NewsID = string(values.Data.([]byte))

		case FID_DisplayTime:
			msg.DisplayTime = values.Data.(time.Time)

		case FID_SourceID:
			msg.SourceID = string(values.Data.([]byte))

		case FID_LanguageList:
			msg.Language = Byte2string874(values.Data.([]byte))

		case FID_OriginalCodeList:
			msg.OriginalCode = Byte2string874(values.Data.([]byte))

		case FID_PictureRef:
			msg.PicRef = Byte2string874(values.Data.([]byte))

		case FID_ExAbstract:
			msg.AbstractData = string(values.Data.([]byte))

		case FID_Transaction:
			msg.Transaction = int16(values.Data.(uint8))

		case FID_isAtt:
			s := values.Data.(uint8)
			if s > 0 {
				msg.IsAtt = true
			}

		case FID_isVdo:
			s := values.Data.(uint16)
			if s > 0 {
				msg.IsVdo = true
			}

		case FID_isImg:
			s := values.Data.(uint8)
			if s > 0 {
				msg.IsImg = true
			}

		case FID_Abstract:
			if len(msg.AbstractData) > 0 {
				continue
			}
			msg.AbstractData = Byte2string874(values.Data.([]byte))

		case FID_Headline:
			msg.Headline = Byte2string874(values.Data.([]byte))

		case FID_StoryTime:
		case 105:
		case 106:
		case 107:
		case 201:
		case 216:
		case 0:
		case FID_AttachmentList:
		default:
			fmt.Printf("unknow field %d\n", values.FID)
		}
	}

	return msg, nil
}

func getField(buff []byte) (Field, []byte, error) {
	//values = append(values, 1, 2, 3, nil, 4, "ok")

	//fType := binary.LittleEndian.Uint32(buff[0:4])
	fType := fieldType(buff[0])
	fid := FieldID(int(binary.BigEndian.Uint32(buff[1:5])))

	switch fType {
	case fType_Str1:
		dlen := int(buff[5])
		data := buff[6 : 6+dlen]
		return Field{fType, fid, data}, buff[6+dlen:], nil
	case fType_Str2:
		dlen := int(binary.LittleEndian.Uint16(buff[5:7]))
		data := buff[7 : 7+dlen]
		return Field{fType, fid, data}, buff[7+dlen:], nil
	case fType_Str4:
		dlen := binary.LittleEndian.Uint32(buff[5:9])
		data := buff[9 : 9+dlen]
		return Field{fType, fid, data}, buff[9+dlen:], nil
	case fType_Str0:
		dlen := 0
		blen := len(buff)
		for i := 5; i < blen; i++ {
			if buff[i] == 0 {
				break
			}
			dlen++
		}
		data := buff[5 : 5+dlen]
		// h := Byte2string874(data)
		return Field{fType, fid, data}, buff[6+dlen:], nil
	case fType_Date0:
		f := float64FromBytes(buff[5:13])
		epo := int64((f - 25569) * 24 * 3600 * 1000)
		data := time.UnixMilli(epo)
		return Field{fType, fid, data}, buff[13:], nil
	case fType_Byte0:
		data := buff[5]
		return Field{fType, fid, data}, buff[6:], nil
	case fType_Char0:
		data := buff[5]
		return Field{fType, fid, data}, buff[6:], nil
	case fType_Long0:
		data := binary.BigEndian.Uint32(buff[5:9])
		return Field{fType, fid, data}, buff[9:], nil
	case fType_List2:
		if fid == FID_HDLList {
			dd := binary.LittleEndian.Uint16(buff[5:7])
			buff = buff[7:]
			var vlst2 Field
			var values [][]byte
			for dd > 0 {

				vlst2, buff, _ = getField(buff)
				hdlBuff := vlst2.Data.([]byte)
				values = append(values, hdlBuff)
				dd--
			}
			return Field{fType, fid, values}, buff, nil
		}
		if fid == FID_AttachmentList || fid == FID_Attachment {
			dd := int(binary.LittleEndian.Uint16(buff[5:7]))
			buff = buff[7:]
			var vlst2 Field
			values := map[int]any{}
			for dd > 0 {
				vlst2, buff, _ = getField(buff)
				values[dd] = vlst2.Data
				dd--
			}
			return Field{fType, fid, values}, buff, nil
		}

		// if fid == FID_ExecutionPlan {
		// 	data, uselen := GetExec(buff)
		// 	return Field{fType, fid, data}, buff[uselen:], nil
		// }
		return Field{}, nil, nil
	case fType_Short0:
		data := binary.BigEndian.Uint16(buff[5:7])
		return Field{fType, fid, data}, buff[7:], nil
	case fType_ULong0:
		data := binary.BigEndian.Uint16(buff[5:7])
		return Field{fType, fid, data}, buff[7:], nil
	case fType_Word0:
		data := binary.BigEndian.Uint16(buff[5:7])
		return Field{fType, fid, data}, buff[7:], nil
	default:
		fmt.Println("What is this fType?", fType)
	}
	return Field{}, nil, nil
}

func verify(buff []byte) (uint32, error) {

	if len(buff) < 14 {
		return 0, errors.New("buff is too short")
	}
	pref := []byte{170, 170, 170, 170, 170, 171}

	if !bytes.Equal(pref, buff[0:6]) {
		return 0, errors.New("invalid prefix")
	}

	msgLen := binary.BigEndian.Uint32(buff[9:13])

	if msgLen+15 > uint32(len(buff)) {
		return 0, errors.New("size package invalid")
	}

	crcval := crc(buff[6 : msgLen+13])
	crc := buff[msgLen+13 : msgLen+15]
	if !bytes.Equal(crcval, crc) {
		return 0, errors.New("invalid crc")
	}
	return msgLen, nil
}
