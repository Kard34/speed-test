package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"time"

	"speed-test/ftime"
	"speed-test/imaz"

	_ "github.com/mattn/go-sqlite3"
	"github.com/varokas/tis620"
)

type Flatnode struct {
	Index int
	Left  int
	Right int
	Value string
}

type Treenode struct {
	Value string
	Left  *Treenode
	Right *Treenode
}

type chunkData struct {
	Index         int
	Position      int
	Allocate      int
	CountDocument int
	StartPosition int
	CountPosition int
}

type idxHeader struct {
	Start    int32
	Allocate int32
	Used     int32
}

type Hdl struct {
	Size        uint16
	InvID       uint64
	Time64      int64
	DisplayTime string
	Timex       time.Time
	IDSize      int
	DocID       string
	IMAZSize    uint16
	Position    int
	Headline    string
	Txid        int
}

var (
	Path     = "./Index/"
	Filename = "20240129"

	Fidx *os.File
	Fhdl *os.File
	Fmap *os.File

	// Query = []Flatnode{{0, 1, 2, "phrase2"}, {1, -1, -1, "ค่าเงิน"}, {2, -1, -1, "บาท"}}
	// Query = []Flatnode{{0, -1, -1, "ประเทศไทย"}}
	// Query = []Flatnode{{0, -1, -1, "2"}} // result : 13811
	// Query = []Flatnode{{0, -1, -1, "all"}} // result : 6507
	// Query = []Flatnode{{0, -1, -1, "it"}} // result : 10204
	// Query = []Flatnode{{0, -1, -1, "the"}} // result : 16742
	Query = []Flatnode{{0, -1, -1, "ที่"}} // result : 16749
	// Query = []Flatnode{{0, 1, 2, "phrase2"}, {1, -1, -1, "และ"}, {2, -1, -1, "ที่"}} // 16749/16069
	// Query = []Flatnode{{0, -1, -1, "และ"}} // result : 16096
	Limit = 10000
	START ftime.CTime
	END   ftime.CTime

	WordPosition  idxHeader
	WordData      idxHeader
	ChunkPosition idxHeader
	ChunkData     idxHeader

	CkData  map[string]chunkData
	HdlData map[uint64]Hdl
)

func main() {
	// ! Open .idx file
	fidx, err := os.Open(Path + Filename + ".idx")
	checkerror(err)
	Fidx = fidx
	defer Fidx.Close()
	// * Open .hdl file
	fhdl, err := os.Open(Path + Filename + ".hdl")
	checkerror(err)
	Fhdl = fhdl
	defer Fhdl.Close()
	// * Open .map file
	fmap, err := os.Open(Path + Filename + ".map")
	checkerror(err)
	Fmap = fmap
	defer Fmap.Close()
	// * Load ...
	Load()
	START.Parse("2024-01-29T00:00:00")
	END.Parse("2024-01-29T23:59:59")
	Root := MakeTree(Query)
	Result := Search(Root, Limit, START, END)
	fmt.Println(Result)
}

func Search(tree *Treenode, limit int, timex, timey ftime.CTime) (listdata []string) {
	Timex := ParseStr(TimeToStr(timex))
	Timey := ParseStr(TimeToStr(timey))
	Buffx := docInvert(Timex.Year(), int(Timex.Month()), Timex.Day(), Timex.Hour(), 0)
	Buffx = append(Buffx, []byte{0, 0, 0}...)
	t := 0
	if Timey.Minute() > 0 || Timey.Second() > 0 {
		t = 1
	}
	Buffy := docInvert(Timey.Year(), int(Timey.Month()), Timey.Day(), Timey.Hour()+t, 0)
	Buffy = append(Buffy, []byte{0, 0, 0}...)
	ID_List := SearchData(tree, Buffx, Buffy)

	for _, x := range ID_List {
		if HdlData[x].Time64 >= timex.UnixMilli() && HdlData[x].Time64 <= timey.UnixMilli() {
			listdata = append(listdata, HdlData[x].DisplayTime+" "+HdlData[x].DocID+" "+HdlData[x].Headline+"\n")
			if len(listdata) == limit {
				break
			}
		}
	}

	sort.Slice(listdata, func(i, j int) bool {
		return listdata[i] < listdata[j]
	})
	return
}

func SearchData(tree *Treenode, buffx, buffy []byte) (invdocid_list []uint64) {
	Chunkdata, Buff := SearchMatching(tree, buffx, buffy)
	for i := 0; i < Chunkdata.CountDocument; i++ {
		Buff8 := make([]byte, 0)
		Buff8 = append(Buff8, Buff[i*10:(i*10)+5]...)
		Buff8 = append(Buff8, []byte{0, 0, 0}...)
		INVDOCID := binary.LittleEndian.Uint64(Buff8)
		invdocid_list = append(invdocid_list, INVDOCID)
	}
	return
}

func SearchMatching(tree *Treenode, buffx, buffy []byte) (chunkdata chunkData, buff []byte) {
	var Chunk1 chunkData
	var Buff1 []byte
	var Chunk2 chunkData
	var Buff2 []byte
	if tree.Left != nil {
		Chunk1, Buff1 = SearchMatching(tree.Left, buffx, buffy)
	}
	if tree.Right != nil {
		Chunk2, Buff2 = SearchMatching(tree.Right, buffx, buffy)
	}
	if tree.Left == nil && tree.Right == nil {
		chunkdata, buff = LoadWord(tree.Value, buffx, buffy)
	} else {
		chunkdata, buff = Match(Chunk1, Chunk2, Buff1, Buff2, tree.Value)
	}
	return
}

func LoadWord(word string, buffx, buffy []byte) (chunkdata chunkData, buff []byte) {
	Lpos1, Found1 := BinaryChunkBuff(buffx, word)
	Lpos2, Found2 := BinaryChunkBuff(buffy, word)
	_ = Found1
	if Found2 {
		Lpos2++
	}
	Allocate := 16
	CountDocument := 0
	CountPosition := 0
	StartPoint := ChunkData.Start + int32(CkData[word].Position) + 16
	INVDOCID_LIST := make([]byte, 0)
	PositionPoint := make([]int32, 2)

	for i := Lpos1; i < Lpos2; i++ {
		Buff1 := make([]byte, 10)
		Fidx.Seek(int64(StartPoint+int32(i*10)), io.SeekStart)
		Fidx.Read(Buff1)
		INVID := Buff1[0:5]
		INDEX := []byte{byte(CountPosition & 255), byte((CountPosition >> 8) & 255), byte((CountPosition >> 16) & 255)}
		LENGTH := Buff1[8:10]
		LengthValue := int(binary.LittleEndian.Uint16(LENGTH))
		Buff10 := make([]byte, 0)
		Buff10 = append(Buff10, INVID...)
		Buff10 = append(Buff10, INDEX...)
		Buff10 = append(Buff10, LENGTH...)
		Allocate += 10
		CountDocument++
		CountPosition += LengthValue
		INVDOCID_LIST = append(INVDOCID_LIST, Buff10...)
		if i == Lpos1 {
			Temp := Buff1[5:8]
			Temp = append(Temp, []byte{0}...)
			PositionPoint[0] = int32(binary.LittleEndian.Uint32(Temp))
		} else if i == Lpos2-1 {
			Temp := Buff1[5:8]
			Temp = append(Temp, []byte{0}...)
			PositionPoint[1] = int32(binary.LittleEndian.Uint32(Temp) + uint32(LengthValue))
		}
	}
	Buff := make([]byte, (PositionPoint[1]-PositionPoint[0])*2)
	Fidx.Seek(int64(StartPoint+int32(CkData[word].StartPosition)+(PositionPoint[0]*2)), io.SeekStart)
	Fidx.Read(Buff)
	buff = append(buff, INVDOCID_LIST...)
	buff = append(buff, Buff...)
	chunkdata = chunkData{CkData[word].Index, CkData[word].Position, Allocate + (CountPosition * 2), CountDocument, Allocate - 16, CountPosition}
	return
}

func BinaryChunkBuff(buffsearch []byte, word string) (lpos int, found bool) {
	LposLo := 0
	LposHi := CkData[word].CountDocument - 1
	CompareResult := -1
	lpos = 0
	StartPoint := ChunkData.Start + int32(CkData[word].Position) + 16
	for LposLo <= LposHi {
		lpos = (LposLo + LposHi) / 2
		Buff := make([]byte, 10)
		Fidx.Seek(int64(StartPoint+int32(lpos*10)), io.SeekStart)
		Fidx.Read(Buff)
		CompareResult = bytes.Compare(buffsearch[0:5], Buff[0:5])
		if CompareResult < 0 {
			LposHi = lpos - 1
		} else if CompareResult > 0 {
			LposLo = lpos + 1
		} else {
			break
		}
	}
	if CompareResult > 0 {
		lpos += 1
	}
	found = CompareResult == 0
	return
}

func Match(cho1, cho2 chunkData, buffw1, buffw2 []byte, op string) (cho chunkData, buff []byte) {
	idx := 0
	jdx := 0
	cho = chunkData{-1, -1, 0, 0, 0, 0}
	buffdoc := make([]byte, 0)
	buffpos := make([]byte, 0)
	buff0 := make([]byte, 5)
	nCompareResult := -1
	start_doc_pos := 0
	len_doc_post := 0
	buff3 := make([]byte, 4)
	buff2 := make([]byte, 2)
	for idx < cho1.CountDocument && jdx < cho2.CountDocument {
		b1 := buffw1[idx*10 : (idx*10)+10]
		b2 := buffw2[jdx*10 : (jdx*10)+10]
		nCompareResult = bytes.Compare(b1[0:5], b2[0:5])
		if nCompareResult < 0 {
			if op == "or" {
				buffdoc = append(buffdoc, b1[0:5]...)
				buffdoc = append(buffdoc, buff0...)
			}
			idx++
		} else if nCompareResult > 0 {
			if op == "or" {
				buffdoc = append(buffdoc, b2[0:5]...)
				buffdoc = append(buffdoc, buff0...)
			}
			jdx++
		} else {
			if op == "or" {
				buffdoc = append(buffdoc, b1[0:5]...)
				buffdoc = append(buffdoc, buff0...)
			} else if op == "and" {
				buffdoc = append(buffdoc, b1[0:5]...)
				buffdoc = append(buffdoc, buff0...)
			} else {
				diff := 3
				if op == "phrase2" {
					diff = 2
				}

				st1, len1 := invposition(b1)
				st2, len2 := invposition(b2)
				if cho1.StartPosition+((st1+len1)*2) > len(buffw1) {
					fmt.Println("error")
				}
				if cho2.StartPosition+((st2+len2)*2) > len(buffw2) {
					fmt.Println("error")
				}

				bo1 := buffw1[cho1.StartPosition+(st1*2) : cho1.StartPosition+((st1+len1)*2)]
				bo2 := buffw2[cho2.StartPosition+(st2*2) : cho2.StartPosition+((st2+len2)*2)]

				pos := comparepharse(bo1, bo2, diff)

				if len(pos) > 0 {
					buffpos = append(buffpos, pos...)
					buffdoc = append(buffdoc, b1[0:5]...)

					binary.LittleEndian.PutUint32(buff3, uint32(start_doc_pos))
					buffdoc = append(buffdoc, buff3[0:3]...)
					binary.LittleEndian.PutUint16(buff2, uint16(len(pos)/2))
					buffdoc = append(buffdoc, buff2...)
					start_check := int(binary.LittleEndian.Uint32(buff3))
					len_check := int(binary.LittleEndian.Uint16(buff2))
					if start_check != start_doc_pos || len_check != len(pos)/2 {
						fmt.Print("error")
					}
					_ = len_doc_post
					cho.CountDocument++
					len_pos := len(pos) / 2
					cho.CountDocument += len_pos
					start_doc_pos += len_pos
				}
			}
			idx++
			jdx++
		}
	}
	buff = append(buff, buffdoc...)
	buff = append(buff, buffpos...)
	cho.Allocate = len(buff)
	cho.CountDocument = len(buffdoc) / 10
	cho.CountPosition = len(buffpos) / 2
	cho.StartPosition = len(buffdoc)
	return
}

func invposition(buff []byte) (start, len int) {
	b3 := make([]byte, 0)
	b3 = append(b3, buff[5:8]...)
	b3 = append(b3, 0)
	start = int(binary.LittleEndian.Uint32(b3))
	len = int(binary.LittleEndian.Uint16(buff[8:10]))
	return
}

func comparepharse(bo1 []byte, bo2 []byte, diff int) (buff []byte) {
	idx := 0
	jdx := 0
	buff = make([]byte, 0)
	if len(bo1) == 0 || len(bo2) == 0 {
		return
	}
	idx = 0
	jdx = 0
	for idx < len(bo1) && jdx < len(bo2) {
		vali := binary.LittleEndian.Uint16(bo1[idx : idx+2])
		valj := binary.LittleEndian.Uint16(bo2[jdx : jdx+2])
		if vali+uint16(diff) > valj {
			jdx += 2
		} else if vali+uint16(diff) < valj {
			idx += 2
		} else {
			buff = append(buff, bo1[idx:idx+2]...)
			idx += 2
			jdx += 2
		}
	}
	return
}

func Load() {
	LoadIDX()
	LoadHDL()
}

func LoadIDX() {
	Head := make([]byte, 48)
	Fidx.Seek(0, io.SeekStart)
	Fidx.Read(Head)
	HeadList := make([]int32, 0)
	for i := 0; i < len(Head)/4; i++ {
		Data := binary.LittleEndian.Uint32(Head[i*4 : (i*4)+4])
		HeadList = append(HeadList, int32(Data))
	}
	WordPosition = idxHeader{HeadList[0], HeadList[1], HeadList[2]}
	WordData = idxHeader{HeadList[3], HeadList[4], HeadList[5]}
	ChunkPosition = idxHeader{HeadList[6], HeadList[7], HeadList[8]}
	ChunkData = idxHeader{HeadList[9], HeadList[10], HeadList[11]}
	word_position := ReadPosition(WordPosition)
	word_list := ReadWord(WordData, word_position)
	chunk_position := ReadPosition(ChunkPosition)
	if len(word_list) == len(chunk_position) {
		CkData = map[string]chunkData{}
		for i := 0; i < len(word_list); i++ {
			Buff := make([]byte, 16)
			Fidx.Seek(int64(ChunkData.Start+int32(chunk_position[i])), io.SeekStart)
			Fidx.Read(Buff)
			temp := make([]int, 0)
			for j := 0; j < 4; j++ {
				Data := binary.LittleEndian.Uint32(Buff[j*4 : (j*4)+4])
				temp = append(temp, int(Data))
			}
			CkData[word_list[i]] = chunkData{i, chunk_position[i], temp[0], temp[1], temp[2], temp[3]}
		}
	}

}

func ReadPosition(head idxHeader) (position_list []int) {
	Buff := make([]byte, head.Used*4)
	Fidx.Seek(int64(head.Start), io.SeekStart)
	Fidx.Read(Buff)
	for i := 0; i < int(head.Used); i++ {
		Data := binary.LittleEndian.Uint32(Buff[i*4 : (i*4)+4])
		position_list = append(position_list, int(Data))
	}
	return
}

func ReadWord(head idxHeader, position_list []int) (word_list []string) {
	Buff := make([]byte, head.Used)
	Fidx.Seek(int64(head.Start), io.SeekStart)
	Fidx.Read(Buff)
	for i := 0; i < len(position_list); i++ {
		var Data []byte
		if i == len(position_list)-1 {
			Data = Buff[position_list[i]:]
		} else {
			Data = Buff[position_list[i] : position_list[i+1]-1]
		}
		word_list = append(word_list, string(tis620.ToUTF8(Data)))
	}
	return
}

func LoadHDL() {
	position_list := ReadMap()
	HdlData = map[uint64]Hdl{}
	for i := range position_list {
		x := ReadHdl(int64(position_list[i]))
		x.Position = int(position_list[i])
		HdlData[x.InvID] = x
	}
}

func ReadHdl(seek int64) (hdl Hdl) {
	Head := make([]byte, 50)
	Fhdl.Seek(seek, io.SeekStart)
	Fhdl.Read(Head)
	Size := binary.LittleEndian.Uint16(Head[48:50])
	Buff := make([]byte, Size)
	Fhdl.Read(Buff)
	Buff8 := make([]byte, 0)
	Buff8 = append(Buff8, Head[2:7]...)
	Buff8 = append(Buff8, []byte{0, 0, 0}...)
	INVID := binary.LittleEndian.Uint64(Buff8)

	BuffTime := Head[7:15]
	Float := float64FromBytes(BuffTime)
	TIME64 := int64((Float - 25569) * 24 * 3600 * 1000)
	TIMEX := time.UnixMilli(TIME64)

	xx, _ := imaz.ParserHDL(Buff)
	var HEADLINE string
	if len(xx.AbstractData) > 0 {
		var v map[string]any
		json.Unmarshal([]byte(xx.AbstractData), &v)
		if hl, found := v["HL"]; found {
			if len(hl.(string)) > 0 {
				HEADLINE = hl.(string)
			}
		}
		if hx, found := v["HX"]; found {
			if len(hx.(string)) > 0 {
				HEADLINE = hx.(string)
			}
		}
	}
	hdl = Hdl{
		Size:        binary.LittleEndian.Uint16(Head[0:2]),
		InvID:       INVID,
		Time64:      TIME64,
		DisplayTime: TIMEX.UTC().Format("2006-01-02T15:04:05"),
		Timex:       TIMEX,
		IDSize:      int(Head[15]),
		DocID:       string(Head[16:48]),
		Position:    0,
		Headline:    HEADLINE,
		IMAZSize:    binary.LittleEndian.Uint16(Head[48:50]),
	}
	return
}

func ReadMap() (position_list []int32) {
	temp := make([]byte, 2980)
	Fmap.Seek(0, io.SeekStart)
	Fmap.Read(temp)
	Size := binary.LittleEndian.Uint32(temp[0:4])
	Buff := make([]byte, Size*4)
	Fmap.Read(Buff)
	for i := 0; i < len(Buff)/4; i++ {
		Data := binary.LittleEndian.Uint32(Buff[i*4 : (i*4)+4])
		position_list = append(position_list, int32(Data))
	}
	return
}

func float64FromBytes(bytes []byte) float64 {
	bits := binary.LittleEndian.Uint64(bytes)
	float := math.Float64frombits(bits)
	return float
}

func MakeTree(query []Flatnode) (root *Treenode) {
	Data := map[int]Flatnode{}
	Found := map[int]int{}
	for _, i := range query {
		Data[i.Index] = i
		Found[i.Index]++
		Found[i.Index]--
		Found[i.Left]++
		Found[i.Right]++
	}
	Head := -1
	for x, y := range Found {
		if y == 0 {
			Head = x
		}
	}
	root = maketree(Data, Head)
	return
}

func maketree(data map[int]Flatnode, head int) (root *Treenode) {
	root = &Treenode{}
	root.Value = data[head].Value
	if data[head].Left != -1 {
		root.Left = maketree(data, data[head].Left)
	}
	if data[head].Right != -1 {
		root.Right = maketree(data, data[head].Right)
	}
	return
}

func checkerror(e error) {
	if e != nil {
		fmt.Println(e)
	}
}

func ParseStr(str string) time.Time {
	Prased, err := time.Parse("2006-01-02 15:04:05 -0700 MST", str)
	if err != nil {
		fmt.Println("Error:", err)
		return time.Time{}
	}
	return Prased
}

func TimeToStr(time ftime.CTime) (str string) {
	str = time.Format("2006-01-02 15:04:05 -0700 MST")
	return
}

func docInvert(year, month, day, hour, running int) (buff []byte) {
	year -= 1950
	month -= 1
	day -= 1
	Value := year*12*31*24 + month*31*24 + day*24 + hour
	buff = make([]byte, 5)
	buff[0] = byte(Value >> 12)
	buff[1] = byte(Value >> 4)
	buff[2] = byte(Value<<4) | byte(running>>8)
	buff[3] = byte(running)
	buff[4] = byte(running >> 8)
	return
}
