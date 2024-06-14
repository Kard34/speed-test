package imaz

import (
	"encoding/binary"
	"math"

	"github.com/sigurn/crc16"
	"golang.org/x/text/encoding/charmap"
)

var win874 = charmap.Windows874.NewDecoder().Transformer

func float64FromBytes(bytes []byte) float64 {
	bits := binary.LittleEndian.Uint64(bytes)
	float := math.Float64frombits(bits)
	return float
}

// func Byte2string874(buff []byte) (str string) {
// 	for _, b := range buff {
// 		if b < 160 {
// 			str += string(b)
// 		} else {
// 			if b == 161 {
// 				str += "ก"
// 			} else if b == 162 {
// 				str += "ข"
// 			} else if b == 163 {
// 				str += "ฃ"
// 			} else if b == 164 {
// 				str += "ค"
// 			} else if b == 165 {
// 				str += "ฅ"
// 			} else if b == 166 {
// 				str += "ฆ"
// 			} else if b == 167 {
// 				str += "ง"
// 			} else if b == 168 {
// 				str += "จ"
// 			} else if b == 169 {
// 				str += "ฉ"
// 			} else if b == 170 {
// 				str += "ช"
// 			} else if b == 171 {
// 				str += "ซ"
// 			} else if b == 172 {
// 				str += "ฌ"
// 			} else if b == 173 {
// 				str += "ญ"
// 			} else if b == 174 {
// 				str += "ฎ"
// 			} else if b == 175 {
// 				str += "ฏ"
// 			} else if b == 176 {
// 				str += "ฐ"
// 			} else if b == 177 {
// 				str += "ฑ"
// 			} else if b == 178 {
// 				str += "ฒ"
// 			} else if b == 179 {
// 				str += "ณ"
// 			} else if b == 180 {
// 				str += "ด"
// 			} else if b == 181 {
// 				str += "ต"
// 			} else if b == 182 {
// 				str += "ถ"
// 			} else if b == 183 {
// 				str += "ท"
// 			} else if b == 184 {
// 				str += "ธ"
// 			} else if b == 185 {
// 				str += "น"
// 			} else if b == 186 {
// 				str += "บ"
// 			} else if b == 187 {
// 				str += "ป"
// 			} else if b == 188 {
// 				str += "ผ"
// 			} else if b == 189 {
// 				str += "ฝ"
// 			} else if b == 190 {
// 				str += "พ"
// 			} else if b == 191 {
// 				str += "ฟ"
// 			} else if b == 192 {
// 				str += "ภ"
// 			} else if b == 193 {
// 				str += "ม"
// 			} else if b == 194 {
// 				str += "ย"
// 			} else if b == 195 {
// 				str += "ร"
// 			} else if b == 196 {
// 				str += "ฤ"
// 			} else if b == 197 {
// 				str += "ล"
// 			} else if b == 198 {
// 				str += "ฦ"
// 			} else if b == 199 {
// 				str += "ว"
// 			} else if b == 200 {
// 				str += "ศ"
// 			} else if b == 201 {
// 				str += "ษ"
// 			} else if b == 202 {
// 				str += "ส"
// 			} else if b == 203 {
// 				str += "ห"
// 			} else if b == 204 {
// 				str += "ฬ"
// 			} else if b == 205 {
// 				str += "อ"
// 			} else if b == 206 {
// 				str += "ฮ"
// 			} else if b == 207 {
// 				str += "ฯ"
// 			} else if b == 208 {
// 				str += "ะ"
// 			} else if b == 209 {
// 				str += "ั"
// 			} else if b == 210 {
// 				str += "า"
// 			} else if b == 211 {
// 				str += "ำ"
// 			} else if b == 212 {
// 				str += "ิ"
// 			} else if b == 213 {
// 				str += "ี"
// 			} else if b == 214 {
// 				str += "ึ"
// 			} else if b == 215 {
// 				str += "ื"
// 			} else if b == 216 {
// 				str += "ุ"
// 			} else if b == 217 {
// 				str += "ู"
// 			} else if b == 218 {
// 				str += "."
// 			} else if b == 219 {
// 				str += ""
// 			} else if b == 220 {
// 				str += ""
// 			} else if b == 221 {
// 				str += ""
// 			} else if b == 222 {
// 				str += ""
// 			} else if b == 223 {
// 				str += "฿"
// 			} else if b == 224 {
// 				str += "เ"
// 			} else if b == 225 {
// 				str += "แ"
// 			} else if b == 226 {
// 				str += "โ"
// 			} else if b == 227 {
// 				str += "ใ"
// 			} else if b == 228 {
// 				str += "ไ"
// 			} else if b == 229 {
// 				str += "ๅ"
// 			} else if b == 230 {
// 				str += "ๆ"
// 			} else if b == 231 {
// 				str += "็"
// 			} else if b == 232 {
// 				str += "่"
// 			} else if b == 233 {
// 				str += "้"
// 			} else if b == 234 {
// 				str += "๊"
// 			} else if b == 235 {
// 				str += "๋"
// 			} else if b == 236 {
// 				str += "์"
// 			} else if b == 237 {
// 				str += "ํ"
// 			} else if b == 238 {
// 				str += ""
// 			} else if b == 239 {
// 				str += ""
// 			} else if b == 240 {
// 				str += "๐"
// 			} else if b == 241 {
// 				str += "๑"
// 			} else if b == 242 {
// 				str += "๒"
// 			} else if b == 243 {
// 				str += "๓"
// 			} else if b == 244 {
// 				str += "๔"
// 			} else if b == 245 {
// 				str += "๕"
// 			} else if b == 246 {
// 				str += "๖"
// 			} else if b == 247 {
// 				str += "๗"
// 			} else if b == 248 {
// 				str += "๘"
// 			} else if b == 249 {
// 				str += "๙"
// 			} else {
// 				str += " "
// 			}
// 		}
// 	}
// 		return
// 	}

// func W() {
// }

func Byte2string874(buff []byte) string {

	b := make([]byte, len(buff)*3)

	bLen, _, _ := win874.Transform(b, buff, false)
	str := string(b[:bLen])
	return str
}

func crc(buff []byte) []byte {

	table := crc16.MakeTable(crc16.CRC16_ARC)
	crc := crc16.Checksum(buff, table)
	// fmt.Printf("CRC-16 MAXIM: %X\n", crc)

	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, crc)

	return b
}
