//backend/logic.go

package main

import ()

//check winner ตรวจสอบผู้ชนะ
//board string 9 ตัวแทนตำแหน่งบนกระดาน เช่น "XOX-O-X--"
func CheckWinner(b string) (string) {
	//ชนะแนวนอน 3 แถว แนวตั้ง 3 แถว แนวทแยง 2 แถว รวมทั้งหมด 8 แบบ
	winLines := [][]int{
		{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, // แนวนอน
		{0, 3, 6}, {1, 4, 7}, {2, 5, 8}, // แนวตั้ง
		{0, 4, 8}, {2, 4, 6}, // แนวทแยง
	}

	// ตรวจสอบแต่ละแบบว่ามีผู้ชนะหรือไม่
	for _, line := range winLines {
		if b[line[0]] != '-' && b[line[0]] == b[line[1]] && b[line[1]] == b[line[2]] {
			return string(b[line[0]]) // คืนค่า "X" หรือ "O"
		}
	}
	// ถ้าไม่มีผู้ชนะและกระดานเต็มแล้วถือว่าเสมอ
	isFull := true
	for i := 0; i < 9; i++ {
		if b[i] == '-' {
			isFull = false
			break
		}
	}

	if isFull{
		return "DRAW"
	}

	return "" //เกมยังไม่จบ 
}

