# Turn-Based Tic-Tac-Toe (Multiplayer)

โปรเจกต์ Web Application เกม Tic-Tac-Toe แบบ Multiplayer สำหรับทดสอบ Fullstack Developer Internship Challenge (@acuitmeshdev) 

โปรเจกต์นี้ได้รับการพัฒนาภายใต้ข้อจำกัด **Stateless RESTful API** อย่างเคร่งครัด (ไม่ใช้ WebSockets, SSE หรือ WebRTC) โดยให้ความสำคัญสูงสุดกับ **ความเสถียรของระบบ (System Integrity), การจัดการ Concurrency, ความถูกต้องของข้อมูล (Data Consistency) และการออกแบบ UI/UX ที่เข้าใจง่าย**

---

## หมายเหตุสำหรับกรรมการ (Note for Reviewers)
**เรื่องไฟล์ `.env` และ Environment Variables:**
เพื่อความสะดวกและประหยัดเวลาในการตรวจโปรเจกต์ โค้ดชุดนี้ได้ทำการตั้งค่า Config พื้นฐาน และ Hardcode ค่า Environment Variables ที่จำเป็นบางส่วน (เช่น `JWT_SECRET` และข้อมูลเชื่อมต่อ Database ใน Docker) ไว้ในซอร์สโค้ดและ `docker-compose.yml` เรียบร้อยแล้ว 

ดังนั้น กรรมการจึง **ไม่จำเป็นต้องสร้างหรือตั้งค่าไฟล์ `.env` ด้วยตัวเอง** สามารถใช้คำสั่ง Docker รันโปรเจกต์และเข้าทดสอบระบบได้ทันที (Zero Configuration)

---

## Live Demo
- **Frontend URL:** `https://your-frontend.vercel.app`
- **Backend API:** `https://ttt-backend-api.onrender.com`

---

## Tech Stack
- **Frontend:** React, Next.js, Tailwind CSS (ออกแบบ UI High-Contrast)
- **Backend:** Go (Golang), Gin Framework
- **Database:** PostgreSQL
- **Infrastructure:** Docker & Docker Compose

---

## Key Technical Features

### 1. Active Session Protection (Anti-Cheat)
ระบบมีการตรวจสอบสถานะผู้เล่นแบบ Real-time เพื่อป้องกันการสร้างห้องหรือจอยห้องซ้ำซ้อน (1 User : 1 Active Game) หากผู้เล่นมีเกมที่อยู่ในสถานะ `WAITING` หรือ `IN_PROGRESS` ระบบจะไม่อนุญาตให้เริ่มเกมใหม่จนกว่าจะจบเกมเดิมหรือกด Leave Arena ก่อน

### 2. Pessimistic Locking (Server-Side)
ใช้ท่า `SELECT ... FOR UPDATE` ภายใน Database Transaction เพื่อล็อกแถวข้อมูลเกมขณะประมวลผลตาเดิน การันตีว่าไม่มีทางเกิดการ "เล่นซ้อนตา" หรือ "ลงหมากทับกัน" แม้จะมีการส่งข้อมูลเข้ามาพร้อมกันก็ตาม

### 3. Robust Polling & Ghost 404 Handling
ในระบบที่ใช้ Short Polling มักเกิดปัญหา Race Condition ตอนเปลี่ยนหน้าห้อง (เช่น หลัง Rematch) ผมได้เพิ่ม **Retry Logic** ที่ฝั่ง Frontend เพื่อจัดการปัญหา 404 ชั่วคราว (Ghost 404) ทำให้การเปลี่ยนผ่านระหว่างห้องเกมลื่นไหลและมั่นคงขึ้น

### 4. Auto-Reconnect & Zombie Session Mitigation
เนื่องจากโปรเจกต์นี้ใช้สถาปัตยกรรมแบบ Stateless RESTful API (ไม่มี WebSocket) หากผู้เล่นปิดเบราว์เซอร์หนี เน็ตหลุด หรือแบตหมดกะทันหัน สถานะเกมจะค้างอยู่ใน Database (Zombie Session) ส่งผลให้ผู้เล่นโดน Soft-locked 
* **Solution:** ระบบได้รับการออกแบบให้มีกลไก **Auto-Reconnect** โดยทุกครั้งที่ผู้เล่นเข้าสู่หน้า Lobby หรือ Login เข้ามาใหม่ ระบบจะยิงเช็ค `Active Session` ทันที หากพบว่ามีเกมที่ค้างอยู่ (สถานะ `IN_PROGRESS` หรือ `WAITING`) ระบบจะแจก Alert แจ้งเตือนและ **"บังคับ Redirect (วาร์ป)"** ผู้เล่นคนนั้นกลับเข้าสู่กระดานเดิมที่ค้างอยู่โดยอัตโนมัติ เพื่อให้เขาสามารถเล่นต่อ หรือกดปุ่ม Leave Arena เพื่อเคลียร์ห้องได้อย่างถูกต้อง
---

## Architecture & System Design

เพื่อให้เป็นไปตามข้อบังคับ "ไม่อนุญาตให้ใช้การสื่อสารแบบ Realtime" และ "Server ต้องเป็น Single Source of Truth" ระบบจึงถูกออกแบบดังนี้:

1. **Stateless Communication:** การสื่อสารทั้งหมดใช้ HTTP Requests มาตรฐาน โดยใช้ **JWT (JSON Web Tokens)** ในการจัดการ Authentication และ Session ของผู้เล่น
2. **Optimized Short Polling:** Frontend จะดึงข้อมูลเกมเพลย์ทุกๆ 1 วินาที เพื่อป้องกันปัญหา Infinite Re-render, Memory Leak และการส่ง Request ซ้อนทับกัน ระบบได้ใช้ท่า **Recursive `setTimeout`** ร่วมกับการตรวจสอบ Data Equality (`JSON.stringify`) ทำให้ React จะ Re-render หน้าจอเฉพาะตอนที่ข้อมูลมีการเปลี่ยนแปลงจริงๆ เท่านั้น
3. **Single Source of Truth:** Frontend ไม่มีส่วนเกี่ยวข้องกับ Game Logic ใดๆ ทั้งสิ้น ทำหน้าที่เพียง Render ข้อมูล JSON จาก Backend เท่านั้น การตรวจจับผู้ชนะ (Win), เสมอ (Draw) และการสลับเทิร์น ถูกคำนวณและควบคุมโดย Server 100%

---

## Concurrency & Race Condition Handling

การจัดการปัญหา Race Condition (เช่น ผู้เล่นพยายามเดินในช่องเดียวกันพร้อมกัน, กดย้ำๆ หรือการใช้ Bot ยิง Request รัวๆ) ถูกป้องกันด้วยสถาปัตยกรรม **Dual-Layer Protection (การป้องกัน 2 ชั้น)**:

### ชั้นที่ 1: Application Level (Pessimistic Locking)
เพื่อป้องกันไม่ให้ Backend ประมวลผลข้อมูลที่ขัดแย้งกัน โลจิกการเดินหมากทั้งหมดจะถูกทำผ่าน Database Transaction โดยใช้ **Row-Level Locking (`SELECT ... FOR UPDATE`)**:

```go
// โค้ดส่วนหนึ่งจาก backend/game_handler.go
tx, err := DB.Begin()
defer tx.Rollback()

// ล็อกข้อมูล Row ของเกมนี้ไว้ เพื่อไม่ให้ Request อื่นเข้ามาแก้ไขได้จนกว่า Transaction นี้จะ Commit
query := `SELECT status, current_turn_id, board FROM games WHERE id = $1 FOR UPDATE`
err = tx.QueryRow(query, gameID).Scan(&status, &currentTurnID, &board)
```

**กลไกการทำงาน:** เมื่อผู้เล่นเดินหมาก Transaction จะทำการ "ล็อก" ข้อมูลเกมห้องนั้นไว้ หากมี Request อื่นถูกยิงเข้ามาในเสี้ยววินาทีเดียวกัน (เช่น กดเบิ้ล หรือเพื่อนกดพร้อมกัน) Request ที่สองจะถูกบังคับให้รอ (Wait) จนกว่า Request แรกจะอัปเดตกระดานเสร็จ เมื่อ Request ที่สองได้ทำงานต่อและดึงข้อมูล State ใหม่มาเช็ค จะพบว่าช่องนั้นไม่ว่างแล้ว หรือไม่ใช่เทิร์นของตัวเอง และจะถูกรีเจ็คด้วย HTTP `400 Bad Request` ทันที โดยที่ข้อมูลไม่เสียหาย

### ชั้นที่ 2: Database Level (Constraint Integrity)
ตาราง `moves` ในฐานข้อมูลได้ถูกตั้งค่า Unique Constraint ไว้เพื่อเป็นด่านป้องกันสุดท้าย:

```sql
CONSTRAINT unique_move_per_cell UNIQUE (game_id, x, y)
```

สิ่งนี้การันตีว่า ต่อให้เกิดข้อผิดพลาดรุนแรงที่ระดับ Application (Layer 1) ตัว PostgreSQL ก็จะปฏิเสธการบันทึกข้อมูลการเดินซ้ำในพิกัดเดียวกันของเกมเดียวกันในระดับ Database เสมอ

---

## Database Schema Design

โครงสร้าง Database แบบ Relational ถูกออกแบบมา 3 ตารางหลัก เพื่อรองรับการทำ Audit Log และระบบ Replay:

* **`users`**: เก็บข้อมูลผู้เล่น `(id, username, password_hash, created_at)`
* **`games`**: จัดการสถานะห้องเกม, ผู้เล่น, เทิร์นปัจจุบัน, และสถานะการกดยอมรับ Rematch `(id, room_code, player1_id, player2_id, current_turn_id, status, board, winner_id, next_room_code, rematch_p1, rematch_p2)`
* **`moves`**: ประวัติการเดินหมาก (Ledger) สำหรับฟีเจอร์ Replay โดยเก็บพิกัดและลำดับการเดินทุกตา `(id, game_id, player_id, x, y, move_order)`

---

## Features Implementation

**Core Requirements (ครบถ้วน):**
- **Secure Authentication:** ระบบ Login/Register เข้ารหัสผ่านด้วย Bcrypt และส่ง Token ด้วย JWT
- **Matchmaking:** สร้างห้อง และจอยห้องผ่าน Room Code หรือ Invite Link
- **Strict Capacity Limits:** รองรับผู้เล่นสูงสุด 2 คนต่อห้องเท่านั้น ป้องกันคนที่ 3 แย่งเข้า
- **Accurate Game Logic:** การคำนวณ Win/Draw และ Turn-based แม่นยำจากฝั่ง Server

**Bonus & Advanced Enhancements (ฟีเจอร์เสริม):**
- **Spectator Mode (โหมดผู้ชม):** ผู้เล่นคนที่ 3 ขึ้นไปที่เข้าห้องมา จะได้รับสถานะ "ผู้ชม" อัตโนมัติ (ไม่มีสิทธิ์กดกระดาน) พร้อม UI ป้ายบอกเทิร์นแบบ Real-time ตามสีของผู้เล่น
- **Interactive Replay System:** เมื่อเกมจบ สามารถกดดูประวัติการเดินย้อนหลังได้แบบ Step-by-step พร้อมปุ่ม Play, Pause, Resume และแถบประวัติ Move Log
- **Mutual Consent Rematch (ห้องเชื่อมโยงอัตโนมัติ):** ระบบเล่นใหม่อีกตาที่ต้องยินยอมทั้ง 2 ฝ่าย (2/2) เมื่อตกลงครบ Server จะสร้างห้องใหม่ สลับเทิร์นให้แฟร์ (ใครเล่นทีหลังตาที่แล้ว จะได้เริ่มก่อน) และ **วาร์ปผู้เล่นพร้อมผู้ชมทุกคนไปยังห้องใหม่โดยอัตโนมัติ**
- **Active Surrender Mechanic:** หากผู้เล่นกด Leave Arena หนีกลางคันขณะที่เกมยัง `IN_PROGRESS` ระบบจะตัดสินให้ผู้เล่นที่อยู่ต่อ **ชนะทันที** พร้อมขึ้นป้าย "Opponent Left" และอัปเดตสถานะห้องเป็น `ABANDONED` ป้องกันการรอแบบไร้จุดหมาย

---

## How to Run Locally

สามารถรันโปรเจกต์ทั้งหมดผ่าน Docker ได้ทันทีโดยไม่ต้องตั้งค่า Environment เพิ่มเติม:

1. **Clone repository:**
   ```bash
   git clone https://github.com/tin-auppati/acuitmesh-ttt-challenge.git
   cd acuitmesh-ttt-challenge
   ```
2. **Start services ด้วย Docker Compose:**
   ```bash
   docker-compose up -d --build
   ``` 
   หรือ 
   ```bash
   ./run_docker-compose.sh
   ``` 
3. **ใช้งานแอปพลิเคชัน:**
   * **Frontend:** [http://localhost:3000](http://localhost:3000)
   * **Backend API:** [http://localhost:8080](http://localhost:8080)

---

## Evaluation Checklist
- [x] Prevent Race Conditions (จัดการ Concurrency ได้จริง)
- [x] Correct Turn-based & Win/Draw Logic
- [x] Code Quality, Clean Structure & Error Handling
- [x] Docker setup & Deployment
