package imaz

import "time"

type Field struct {
	Ftype fieldType
	FID   FieldID
	Data  interface{}
}

type Hdl struct {
	Proximity    uint8
	NewsID       string
	NewsIDKF4    string
	DisplayTime  time.Time
	SourceID     string
	Headline     string
	Story        string
	Language     string
	OriginalCode string
	AbstractData string
	Symbol       string
	StandingHead string
	PicRef       string
	IsAtt        bool
	IsImg        bool
	IsVdo        bool
	Transaction  int16
}

type fieldType byte

const (
	fType_Char0  fieldType = 0
	fType_Byte0  fieldType = 1
	fType_Short0 fieldType = 2
	fType_Word0  fieldType = 3
	fType_Long0  fieldType = 4
	fType_ULong0 fieldType = 5
	fType_Str0   fieldType = 6
	fType_Str1   fieldType = 7
	fType_Str2   fieldType = 8
	fType_Str4   fieldType = 9
	fType_List1  fieldType = 10
	fType_List2  fieldType = 11
	fType_Date0  fieldType = 12
	fType_Ustr   fieldType = 13
)

type FieldID int

const (
	FID_BaseFieldID FieldID = 100
	FID_Transaction FieldID = 101
	FID_StoryTime   FieldID = 102
	FID_DisplayTime FieldID = 103
	//FID_xx FieldID = 106
	FID_Headline             FieldID = 110
	FID_Story                FieldID = 111
	FID_NewsID               FieldID = 112
	FID_UserName             FieldID = 116
	FID_GroupName            FieldID = 117
	FID_EID                  FieldID = 118
	FID_ErrorCode            FieldID = 119
	FID_SID                  FieldID = 121
	FID_QID                  FieldID = 122
	FID_ExecutionPlan        FieldID = 123
	FID_ExecutionPlanElement FieldID = 124
	FID_StartTime            FieldID = 125
	FID_EndTime              FieldID = 126
	FID_IsOldDateFirst       FieldID = 127
	FID_TotalRequest         FieldID = 128
	FID_FirstResponse        FieldID = 129
	FID_EstimatedTotal       FieldID = 130
	FID_IsFirst              FieldID = 131
	FID_IsMiddle             FieldID = 132
	FID_IsLast               FieldID = 133
	FID_IsEndOfSearch        FieldID = 134
	FID_HDLList              FieldID = 135
	FID_IsTotalExact         FieldID = 136
	FID_HDL                  FieldID = 137
	FID_Proximity            FieldID = 138
	FID_SourceID             FieldID = 200
	FID_LanguageList         FieldID = 202
	FID_OriginalCodeList     FieldID = 204
	FID_Abstract             FieldID = 205
	FID_AttachmentList       FieldID = 210
	FID_Attachment           FieldID = 211
	FID_AttachmentID         FieldID = 212
	FID_AttachmentName       FieldID = 213
	FID_AttachmentType       FieldID = 214
	FID_AttachmentDetail     FieldID = 215
	FID_PictureRef           FieldID = 217
	FID_isAtt                FieldID = 218
	FID_isImg                FieldID = 219
	FID_isVdo                FieldID = 220
	FID_ExAbstract           FieldID = 221
)
