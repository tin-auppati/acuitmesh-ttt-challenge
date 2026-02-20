"use client";

import { useState, useEffect, useCallback } from "react";
import { useRouter, useParams } from "next/navigation";

interface GameData {
  id: number;
  room_code: string;
  player1_id: number;
  player2_id: number | null;
  current_turn_id: number;
  board: string;
  status: string;
  winner_id: number | null;
  next_room_code: string | null; 
  rematch_p1: boolean;          
  rematch_p2: boolean;
}

interface MoveData {
  id: number;
  game_id: number;
  player_id: number;
  x: number;
  y: number;
  created_at: string;
}

export default function GameBoardPage() {
  const router = useRouter()
  const params = useParams(); // ใช้ดึง params.id จาก url
  const roomCode = params.id as string;

  const [game, setGame] = useState<GameData | null>(null);
  const [myUserId, setMyUserId] = useState<number | null>(null);
  const [error, setError] = useState("");
  const [copied, setCopied] = useState(false);

  //replay
  const [isReplaying, setIsReplaying] = useState(false);
  const [isPaused, setIsPaused] = useState(false);
  const [movesLog, setMovesLog] = useState<MoveData[]>([]);
  const [replayStep, setReplayStep] = useState(0);

  const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

  useEffect(() => {
    const token = localStorage.getItem("token")
    const storedUserId = localStorage.getItem("user_id")

    if (!token){
      alert("Please login first to join the match!");
      router.push("/");
      return;
    }

    if (storedUserId) setMyUserId(parseInt(storedUserId, 10));
  }, [router]);

  //polling
  const fetchGameState = useCallback(async () => {
    const token = localStorage.getItem("token");
    if (!token) return;

    try{
      const res = await fetch(`${API_URL}/api/games/${roomCode}`, {
        headers: { Authorization: `Bearer ${token}` },
      });

      if (!res.ok){
        if (res.status === 404) {
          setError("Room not found or has been destroyed.");
        }
        return;
      }

      const data: GameData = await res.json()
      setGame(data)

      //เข้าผ่าน invite link
      const isPlayer1 = data.player1_id === myUserId;
      if(data.status === "WAITING" && !isPlayer1 && data.player2_id === null && myUserId !== null){
        autoJoinMatch(token);
      }

    } catch (err:any){
      console.error("Polling error:", err);
    }
    
  }, [API_URL, roomCode, myUserId]);

  const autoJoinMatch = async (token: string) => {
    try {
      await fetch(`${API_URL}/api/games/join`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ room_code: roomCode }),
      });
      //ยิงเสร็จก็รอ polling รอบหน้า
    } catch (err) {
      console.error("Auto-join failed:", err);
    }
  }

  //polling ทุก 1 วินาที
  useEffect(() => {
    fetchGameState() //เริ่มมาดึงเลย
    const intervalId = setInterval(fetchGameState, 1000); //ดึงซ้ำทุก 1 วิ
    return () => clearInterval(intervalId); 
  }, [fetchGameState]);

  useEffect(() => {
    if (game?.next_room_code) {
      router.push(`/game/${game.next_room_code}`);
    }
  }, [game?.next_room_code, router]);

  //เดืนหมาก
  const handleMove = async (index: number) => {
    if (!game || game.status !== "IN_PROGRESS" || game.current_turn_id !== myUserId) return;
    if (game.board[index] !== "-") return;

    // แปลง index (0-8) เป็น x, y สำหรับส่งให้ Backend
    // y แถว x คอลลัมน์
    const y = Math.floor(index / 3);
    const x = index % 3;

    const token = localStorage.getItem("token");
    try {
      const res = await fetch(`${API_URL}/api/games/move`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ room_code: roomCode, x, y }),
      });

      if (!res.ok) {
        const data = await res.json();
        alert(data.error);
        return;
      }

      //เดินสำเร็จ fetch ใหม่ทันที ไม่รอ polling
      fetchGameState();
    } catch (err: any) {
      console.error(err)
    }
  };

  //func copy link
  const copyInviteLink = () => {
    const link = `${window.location.origin}/game/${roomCode}`;
    navigator.clipboard.writeText(link);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  //func cancel match
  const handleCancelMatch = async () => {
    if(!confirm("Are you sure you want to cancel this match and destroy the room?")) return;

    const token = localStorage.getItem("token");
    try {
      // 🌟 ยิง API DELETE ไปหา Backend
      const res = await fetch(`${API_URL}/api/games/${roomCode}`, {
        method: "DELETE",
        headers: { Authorization: `Bearer ${token}` },
      });

      if (res.ok) {
        // ถ้าทำลายสำเร็จ พาวาร์ปกลับ Lobby
        router.push("/lobby");
      } else {
        const data = await res.json();
        alert(data.error || "Failed to cancel match");
      }

    }catch (err:any) {
      console.error(err)
    }
  };

  //func replay
  const handleWatchReplay = async () => {
    const token = localStorage.getItem("token");
    try {
      const res = await fetch(`${API_URL}/api/games/${roomCode}/moves`, {
        headers: { Authorization: `Bearer ${token}` },
      });

      if (!res.ok) throw new Error("Failed to fetch move");

      const data = await res.json()
      const moves: MoveData[] = data.moves || [];
      setMovesLog(moves);

      //เริ่ม replay
      setIsReplaying(true);
      setIsPaused(false);
      setReplayStep(0);

    } catch (err: any) {
      console.error("Replay error:", err);
      alert("ไม่สามารถโหลดประวัติการเดินได้");
    }
  };
  useEffect(() => {
    let timer: NodeJS.Timeout;
    
    if (isReplaying && !isPaused && replayStep < movesLog.length) {
      timer = setTimeout(() => {
        setReplayStep((prev) => prev + 1); // ขยับไป 1 สเต็ป ทุก 1 วิ
      }, 1000);
    } else if (isReplaying && replayStep === movesLog.length) {
      setIsPaused(true);
    }

    // Cleanup เวลากด Pause หรือ Component Unmount
    return () => clearTimeout(timer);
  }, [isReplaying, isPaused, replayStep, movesLog.length]);

  const handleRematch = async () => {
    const token = localStorage.getItem("token");
    try {
      const res = await fetch(`${API_URL}/api/games/${roomCode}/rematch`, {
        method: "POST",
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!res.ok) {
        const data = await res.json();
        alert(data.error || "Failed to request rematch");
        return;
      }
      fetchGameState();
    } catch (err: any) {
      console.error(err);
    }
  }

  const handleLeaveArena = async () => {
    
    if (game?.status === "IN_PROGRESS") {
      if (!confirm("Are you sure you want to surrender and leave the arena?")) return;
    }

    //ยิง api บอกฃ
    const token = localStorage.getItem("token");
    try {
      await fetch(`${API_URL}/api/games/${roomCode}/leave`, {
        method: "POST",
        headers: { Authorization: `Bearer ${token}` },
      });
    } catch (err) {
      console.error("Failed to notify server about leaving", err);
    }
    // วาร์ปกลับ Lobby
    router.push("/lobby");
  }

  if (error) {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center bg-white text-black font-sans">
        <div className="bg-black text-white p-10 border-4 border-black shadow-[8px_8px_0px_0px_rgba(220,38,38,1)] text-center">
          <h1 className="text-3xl font-black text-red-600 mb-4">ERROR</h1>
          <p className="font-bold uppercase tracking-widest">{error}</p>
          <button onClick={() => router.push("/lobby")} className="mt-6 bg-white text-black px-6 py-2 font-bold uppercase hover:bg-red-600 hover:text-white transition-colors">
            Return to Lobby
          </button>
        </div>
      </div>
    );
  }

  if (!game || !myUserId) return <div className="min-h-screen flex items-center justify-center bg-white text-black font-black uppercase tracking-widest text-2xl animate-pulse">Loading Arena...</div>;

  const isPlayer1 = game.player1_id === myUserId;
  const isPlayer2 = game.player2_id === myUserId;
  const isMyTurn = game.current_turn_id === myUserId;
  const mySymbol = isPlayer1 ? "X" : isPlayer2 ? "O" : "Spectator";

  //replay board
  const getDisplayBoard = () => {
    //โหมดปกติ
    if(!isReplaying) return game.board

    //โหมด replay
    let boardArr = "---------".split("");
    for (let i = 0; i < replayStep; i++) {
      const move = movesLog[i];
      const char = move.player_id === game.player1_id ? "X" : "O";
      boardArr[move.y * 3 + move.x] = char;
    }
    return boardArr.join("");
  }

  const displayBoard = getDisplayBoard();
   
  return (
    <div className="min-h-screen flex flex-col items-center py-12 bg-white text-black font-sans selection:bg-red-500 selection:text-white">
      
      {/* STATE: รอผู้เล่น */}
      {game.status === "WAITING" && (
        <div className="bg-black text-white p-10 border-4 border-black shadow-[8px_8px_0px_0px_rgba(220,38,38,1)] text-center max-w-lg w-full">
          <h2 className="text-2xl font-black uppercase text-red-600 mb-2 animate-pulse">Awaiting Challenger</h2>
          <p className="text-gray-400 font-bold uppercase tracking-widest text-sm mb-8">Share this code to invite an opponent</p>
          
          <div className="text-6xl font-black tracking-tighter mb-8 border-y-4 border-white py-4 bg-gray-900">
            {game.room_code}
          </div>

          <button
            onClick={copyInviteLink}
            className={`w-full font-black uppercase tracking-widest py-4 border-2 transition-all ${
              copied ? "bg-green-500 border-green-500 text-black" : "bg-red-600 border-red-600 text-white hover:bg-red-700"
            }`}
          >
            {copied ? "Link Copied!" : "Copy Invite Link"}
          </button>

          {/* ปุ่มทำลายห้อง (โชว์เฉพาะ Player 1) */}
          {isPlayer1 && (
            <button
              onClick={handleCancelMatch}
              className="text-gray-500 font-bold uppercase tracking-widest text-sm hover:text-red-600 transition-colors underline underline-offset-4 mt-2"
            >
              Cancel Match & Destroy Room
            </button>
          )}
        </div>
      )}

      {/* STATE: กำลังเล่น หรือ จบเกมแล้ว */}
      {game.status !== "WAITING" && (
        <div className="w-full max-w-md flex flex-col items-center">
          
          {/* Header แสดงสเตตัส */}
          <div className="w-full bg-black text-white p-4 border-4 border-black shadow-[8px_8px_0px_0px_rgba(220,38,38,1)] flex justify-between items-center mb-8">
            <div className="font-bold uppercase tracking-widest text-sm">
              Room: <span className="text-red-500">{game.room_code}</span>
            </div>
            <div className="font-bold uppercase tracking-widest text-sm bg-white text-black px-2 py-1">
              You are: <span className={mySymbol === "X" ? "text-red-600" : "text-blue-600"}>{mySymbol}</span>
            </div>
          </div>

          {/* ป้ายประกาศผล (ถ้าเกมจบ) */}
          {!isReplaying && (game.status === "FINISHED" || (game.status === "ABANDONED" && game.winner_id !== null)) && (() => {
            // 🌟 เช็คว่าใครชนะ และดึงสีประจำตัวมาใช้
            const isPlayer1Win = game.winner_id === game.player1_id;
            const winnerName = isPlayer1Win ? "Player 1" : "Player 2";
            const winnerSymbol = isPlayer1Win ? "X" : "O";
            const winBgColor = isPlayer1Win ? "bg-red-600" : "bg-blue-600";
            
            let message = "";
            let bgColor = winBgColor; // สีพื้นหลังอิงตามคนชนะเป็นหลัก

            if (mySymbol === "Spectator") {
              // ถ้าเป็นผู้ชม ให้บอกชัดๆ เลยว่าใครชนะ พร้อมพื้นหลังสีคนนั้น
              message = `🏆 ${winnerName} (${winnerSymbol}) WINS!`;
            } else if (game.winner_id === myUserId) {
              message = "🏆 VICTORY!";
            } else {
              message = "☠️ DEFEAT!";
              bgColor = "bg-gray-800"; // ฝั่งที่แพ้ให้ลดความเด่นลงเป็นสีเทาเข้ม
            }

            return (
              <div className={`w-full text-white text-center p-4 font-black uppercase tracking-widest text-2xl border-4 border-black mb-8 animate-bounce ${bgColor}`}>
                {message}
              </div>
            );
          })()}

          {/* ป้ายกรณีเสมอ (ใช้สีเทาเหมือนกันทุกคน) */}
          {!isReplaying && (game.status === "DRAW" || (game.status === "ABANDONED" && game.winner_id === null)) && (
            <div className="w-full bg-yellow-400 text-black text-center p-4 font-black uppercase tracking-widest text-2xl border-4 border-black mb-8">
              🤝 IT'S A DRAW
            </div>
          )}

          {/* ป้ายบอกเทิร์น (ถ้าเกมยังไม่จบ) */}
          {game.status === "IN_PROGRESS" && (() => {
            let turnText = "⏳ OPPONENT'S TURN...";
            let bgColorClass = "bg-white text-gray-400 border-gray-300"; // สีเทาตอนรอเพื่อนเดิน

            if (isMyTurn) {
              turnText = "🔥 YOUR TURN";
              bgColorClass = "bg-black text-white border-black"; // สีดำเข้มตอนตาเราเดิน
            } else if (mySymbol === "Spectator") {
              // เช็คว่าเป็น Player 1 หรือ 2 ที่กำลังเดินอยู่
              const isPlayer1Turn = game.current_turn_id === game.player1_id;
              const activePlayer = isPlayer1Turn ? "Player 1" : "Player 2";
              const activeSymbol = isPlayer1Turn ? "X" : "O";
              
              turnText = `👀 ${activePlayer} (${activeSymbol})'S TURN`;
              // ใส่สีพื้นหลังแยกชัดเจนให้คนดูเห็นเลย (X แดง, O น้ำเงิน)
              bgColorClass = isPlayer1Turn 
                ? "bg-red-600 text-white border-black" 
                : "bg-blue-600 text-white border-black";
            }

            return (
              <div className={`w-full text-center p-3 font-black uppercase tracking-widest text-xl border-4 mb-8 transition-colors ${bgColorClass}`}>
                {turnText}
              </div>
            );
          })()}

          {/* กระดาน Tic-Tac-Toe */}
          <div className="grid grid-cols-3 gap-2 bg-black p-2 border-4 border-black shadow-[8px_8px_0px_0px_rgba(220,38,38,1)]">
            {displayBoard.split("").map((cell, index) => {
              const isX = cell === "X";
              const isO = cell === "O";
              const isEmpty = cell === "-";

              const isInteractable = isEmpty && isMyTurn && game.status === "IN_PROGRESS" && !isReplaying;

              return (
                <button
                  key={index}
                  onClick={() => handleMove(index)}
                  disabled={!isInteractable}
                  className={`w-24 h-24 flex items-center justify-center text-5xl font-black transition-all ${
                    isInteractable 
                      ? "bg-white hover:bg-gray-200 cursor-pointer active:scale-95" // ถ้ากดได้ ให้มี Hover และคลิกยุบตัวได้
                      : isEmpty 
                        ? "bg-white cursor-default" // ถ้าเป็นช่องว่างแต่ไม่ใช่ตาเรา/เป็นผู้ชม ให้ใช้ cursor-default ปกติ
                        : "bg-gray-100 cursor-default" // ช่องที่ถูกกา X/O ไปแล้ว
                  } ${isX ? "text-red-600" : isO ? "text-blue-600" : ""}`}
                >
                  {isEmpty ? "" : cell}
                </button>
              );
            })}
          </div>
          
          {/* โซนปุ่มควบคุมหลังจากเกมจบ */}
          <div className="mt-8 flex flex-col w-full space-y-4">
            {(game.status === "FINISHED" || game.status === "DRAW" || game.status === "ABANDONED") && !isReplaying && (() => {
              // คำนวณสถานะ Rematch
              const rematchCount = (game.rematch_p1 ? 1 : 0) + (game.rematch_p2 ? 1 : 0);
              const hasAgreedToRematch = (isPlayer1 && game.rematch_p1) || (isPlayer2 && game.rematch_p2);
              const isAbandoned = game.status === "ABANDONED";

              return (
                <>
                  {/* ปุ่ม Rematch (โชว์เฉพาะ Player 1 และ 2) */}
                  {mySymbol !== "Spectator" && (
                    <button
                      onClick={handleRematch}
                      disabled={hasAgreedToRematch || isAbandoned}
                      className={`w-full font-black uppercase tracking-widest py-3 border-4 border-black transition-all ${
                        hasAgreedToRematch 
                          ? "bg-gray-400 text-black cursor-wait" 
                          : "bg-green-500 text-black shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] hover:translate-y-1 hover:shadow-none"
                      }`}
                    >
                      {isAbandoned 
                        ? "🚫 Opponent Left" 
                        : hasAgreedToRematch 
                          ? `⏳ Waiting for Opponent (${rematchCount}/2)` 
                          : `🔄 Rematch (${rematchCount}/2)`}
                    </button>
                  )}

                  <button
                    onClick={handleWatchReplay}
                    className="w-full bg-black text-white font-black uppercase tracking-widest py-3 border-4 border-black shadow-[4px_4px_0px_0px_rgba(220,38,38,1)] hover:translate-y-1 hover:shadow-none transition-all"
                  >
                    🎥 Watch Replay
                  </button>
                </>
              );
            })()}
            
            {/* ปุ่มคุม Replay (จะขึ้นมาตอนกำลังฉายซ้ำอยู่เท่านั้น) */}
            {isReplaying && (
              <div className="flex space-x-2">
                <button
                  // ถ้าย้อนกลับไปดูตอนที่มันจบไปแล้ว (replayStep === movesLog.length) พอกดปุ่มนี้ให้รีเซ็ตกลับไปหน้า 0 ใหม่
                  onClick={() => {
                    if (replayStep === movesLog.length) setReplayStep(0);
                    setIsPaused(!isPaused);
                  }}
                  className={`flex-1 font-black uppercase tracking-widest py-3 border-4 border-black shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] hover:translate-y-1 hover:shadow-none transition-all ${
                    isPaused ? "bg-green-500 text-black" : "bg-yellow-400 text-black"
                  }`}
                >
                  {isPaused ? "▶ Resume" : "⏸ Pause"}
                </button>

                <button
                  onClick={() => setIsReplaying(false)}
                  className="flex-1 bg-red-600 text-white font-black uppercase tracking-widest py-3 border-4 border-black shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] hover:translate-y-1 hover:shadow-none transition-all"
                >
                  ⏹ Stop
                </button>
              </div>
            )}


            {/* ประวัติการเดินแบบ Text (จะโชว์เฉพาะตอนกดดู Replay) */}
            {isReplaying && movesLog.length > 0 && (
              <div className="bg-gray-100 border-2 border-black p-4 text-sm font-mono mt-4 max-h-40 overflow-y-auto">
                <h3 className="font-bold mb-2 border-b-2 border-black pb-1 uppercase">Move History Log</h3>
                {movesLog.map((m, i) => (
                  <div key={m.id} className={i + 1 === replayStep ? "bg-yellow-200 font-bold" : "text-gray-600"}>
                    Step {i + 1}: Player {m.player_id === game.player1_id ? "1 (X)" : "2 (O)"} placed at [Row {m.y}, Col {m.x}]
                  </div>
                ))}
              </div>
            )}

          {/* ปุ่มออกจากห้อง */}
          <button
            onClick={handleLeaveArena}
            className="mt-12 text-gray-500 font-bold uppercase tracking-widest text-sm hover:text-red-600 transition-colors underline underline-offset-4"
          >
            Leave Arena
          </button>
          </div>
        </div>
      )}

    </div>
  );
}