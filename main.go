package main

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Kard34/speed-test/ftime"

	_ "github.com/mattn/go-sqlite3"
)

type Flatnode struct {
	Index int    `json: "index"`
	Left  int    `json: "left"`
	Right int    `json: "right"`
	Value string `json: "value"`
}

type Treenode struct {
	Value string
	Left  *Treenode
	Right *Treenode
}

type ChunkData struct {
	Index         int
	Position      int
	Allocate      int
	CountDocument int
	StartPosition int
	CountPosition int
}

var (
	Path     = "./Index/"
	Filename = "20240129"

	Fidx *os.File
	Db   *sql.DB

	// Query = []Flatnode{{0, 1, 2, "phrase2"}, {1, -1, -1, "ค่าเงิน"}, {2, -1, -1, "บาท"}}
	// query = "('ค่าเงิน','บาท')"

	// Query = []Flatnode{{0, -1, -1, "ประเทศไทย"}}
	// query = "('ประเทศไทย')"

	// Query = []Flatnode{{0, -1, -1, "2"}} // result : 13811
	// query = "('2')"

	// Query = []Flatnode{{0, -1, -1, "all"}} // result : 6507
	// query = "('all')"

	// Query = []Flatnode{{0, -1, -1, "it"}} // result : 10204
	// query = "('it')"

	// Query = []Flatnode{{0, -1, -1, "the"}} // result : 16742
	// query = "('the')"

	// Query = []Flatnode{{0, -1, -1, "ที่"}} // result : 16749
	// query = "('ที่')"

	// Query = []Flatnode{{0, 1, 2, "or"}, {1, -1, -1, "อธิบาย"}, {2, -1, -1, "ทองคำ"}}
	// query = "('อธิบาย','ทองคำ')"

	// Query = []Flatnode{{0, -1, -1, "ยิว"}} // result : 13
	// query = "('ยิว')"

	Query = []Flatnode{{0, 1, 2, "and"}, {1, -1, -1, "the"}, {2, 3, 4, "phrase2"}, {3, -1, -1, "ที่"}, {4, -1, -1, "2"}} //16742, 16749, 13811
	query = "('the','ที่','2')"

	// Query = []Flatnode{{0, 1, 2, "phrase2"}, {1, -1, -1, "ที่"}, {2, -1, -1, "2"}}
	// query = "('ที่','2')"

	// Query = []Flatnode{{0, -1, -1, "and"}}
	// query = "('and')"

	// Query = []Flatnode{{0, 1, 2, "and"}, {1, -1, -1, "and"}, {2, -1, -1, "2"}}
	// query = "('and','2')"

	// Query = []Flatnode{{0, 1, 2, "and"}, {1, -1, -1, "and"}, {2, 3, 4, "and"}, {3, -1, -1, "2"}, {4, 5, 6, "and"},
	// 	{5, -1, -1, "3"}, {6, 7, 8, "and"}, {7, -1, -1, "4"}, {8, 9, 10, "and"},
	// 	{9, -1, -1, "5"}, {10, 11, 12, "and"}, {11, -1, -1, "6"}, {12, 13, 14, "and"},
	// 	{13, -1, -1, "7"}, {14, 15, 16, "and"}, {15, -1, -1, "8"}, {16, -1, -1, "9"}}
	// query  = "('and','2','3','4','5','6','7','8','9')"
	Limit  = 10
	Offset = 0
	START  ftime.CTime
	END    ftime.CTime

	CkData map[string]ChunkData
	CkBuff []byte
)

func main() {
	OAx := time.Now()
	// ! Open .idx file
	fidx, err := os.Open(Path + Filename + ".idx")
	checkERROR(err)
	Fidx = fidx
	defer Fidx.Close()

	// ! Open .sqlite file
	db, err := sql.Open("sqlite3", Path+Filename+".sqlite")
	checkERROR(err)
	Db = db
	defer Db.Close()
	Lx := time.Now()

	// ! Load word chunkdata header
	Load()
	Ly := time.Now()

	START.Parse("2024-01-29T00:00:00")
	END.Parse("2024-01-29T23:59:59")
	Root := MakeTree(Query)
	Sx := time.Now()
	Result := Search(Root, Limit, Offset, START, END)
	Sy := time.Now()
	// fmt.Println(Result)
	fmt.Println("Result : ", len(Result))
	OAy := time.Now()

	fmt.Println("Load Time : ", Ly.Sub(Lx))
	fmt.Println("Search Time : ", Sy.Sub(Sx))
	fmt.Println("Overall Time : ", OAy.Sub(OAx))
}

func Search(tree *Treenode, limit, offset int, timex, timey ftime.CTime) (listdata []string) {
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
	IDList := SearchData(tree, Buffx, Buffy)
	fmt.Println("Result (ID) : ", len(IDList))
	placeholders := make([]string, len(IDList))
	args := make([]interface{}, len(IDList)+4)
	for i, id := range IDList {
		placeholders[i] = "?"
		args[i] = id
	}
	args[len(args)-4] = timex.UnixMilli()
	args[len(args)-3] = timey.UnixMilli()
	args[len(args)-2] = limit
	args[len(args)-1] = offset
	x := `
	SELECT DOCID, TIME64, HEADLINE 
	FROM HDL 
	WHERE INVDOCID IN` + `(` + strings.Join(placeholders, ",") + `)
	AND TIME64 BETWEEN ? AND ?
	ORDER BY TIME64
	LIMIT ? OFFSET ?`
	rows, err := Db.Query(x, args...)
	checkERROR(err)
	defer rows.Close()
	for rows.Next() {
		var DOCID string
		var TIME64 int64
		var HEADLINE string
		err := rows.Scan(&DOCID, &TIME64, &HEADLINE)
		checkERROR(err)

		DisplayTime := time.UnixMilli(int64(TIME64))
		listdata = append(listdata, DisplayTime.UTC().Format("2006-01-02T15:04:05")+" "+DOCID+" "+HEADLINE+"\n")
	}
	if err := rows.Err(); err != nil {
		fmt.Println(err)
	}
	return
}

func SearchData(tree *Treenode, buffx, buffy []byte) (invdocid_list []uint64) {
	Chunkdata, Buff := SearchMatching(tree, buffx, buffy)
	for i := 0; i < Chunkdata.CountDocument; i++ {
		Buff8 := make([]byte, 0)
		Buff8 = append(Buff8, Buff[i*10:(i*10)+5]...)
		Buff8 = append(Buff8, []byte{0, 0, 0}...)
		INVDOCID := binary.LittleEndian.Uint64(Buff8)
		Buff := make([]byte, 8)
		binary.LittleEndian.PutUint64(Buff, INVDOCID)
		invdocid_list = append(invdocid_list, INVDOCID)
	}
	return
}

func SearchMatching(tree *Treenode, buffx, buffy []byte) (chunkdata ChunkData, buff []byte) {
	var Chunk1 ChunkData
	var Buff1 []byte
	var Chunk2 ChunkData
	var Buff2 []byte
	if tree.Left != nil {
		Chunk1, Buff1 = SearchMatching(tree.Left, buffx, buffy)
	}
	if tree.Right != nil {
		Chunk2, Buff2 = SearchMatching(tree.Right, buffx, buffy)
	}
	if tree.Left == nil && tree.Right == nil {
		chunkdata, buff = LoadWordFull(tree.Value)
	} else {
		chunkdata, buff = Match(Chunk1, Chunk2, Buff1, Buff2, tree.Value)
	}
	return
}

func LoadWordFull(word string) (chunkdata ChunkData, buff []byte) {
	query := "SELECT BUFF FROM IDX WHERE WORD='" + word + "'"
	rows, err := Db.Query(query)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var Buff []byte
		err := rows.Scan(&Buff)
		if err != nil {
			fmt.Println(err)
			return
		}
		buff = Buff
	}
	chunkdata = CkData[word]

	err = rows.Err()
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

func Match(cho1, cho2 ChunkData, buffw1, buffw2 []byte, op string) (cho ChunkData, buff []byte) {
	idx := 0
	jdx := 0
	cho = ChunkData{-1, -1, 0, 0, 0, 0}
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
	x := "SELECT WORD, WORDINDEX, POSITION, ALLOCATE, COUNTDOCUMENT, STARTPOSITION, COUNTPOSITION FROM IDX WHERE WORD IN " + query
	rows, err := Db.Query(x)
	checkERROR(err)

	CkData = map[string]ChunkData{}

	for rows.Next() {
		var WORD string
		var INDEX int
		var POSITION int
		var ALLOCATE int
		var COUNTDOCUMENT int
		var STARTPOSITION int
		var COUNTPOSITION int

		err := rows.Scan(&WORD, &INDEX, &POSITION, &ALLOCATE, &COUNTDOCUMENT, &STARTPOSITION, &COUNTPOSITION)
		checkERROR(err)
		CkData[WORD] = ChunkData{INDEX, POSITION, ALLOCATE, COUNTDOCUMENT, STARTPOSITION, COUNTPOSITION}
	}
	if err = rows.Err(); err != nil {
		panic(err)
	}
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

func checkERROR(e error) {
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
