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
}

export default function GameBoardPage() {
  const router = useRouter()
  const params = useParams(); // ใช้ดึง params.id จาก url
  const roomCode = params.id as string;

  const [game, setGame] = useState<GameData | null>(null);
  const [myUserId, setMyUserId] = useState<number | null>(null);
  const [error, setError] = useState("");
  const [copied, setCopied] = useState(false);

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

  //เดืนหมาก
  const handleMove = async (index: number) => {
    if(!game || game.status == "IN_PROGRESS" || game.current_turn_id !== myUserId) return;
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
          {game.status === "FINISHED" && (
            <div className="w-full bg-red-600 text-white text-center p-4 font-black uppercase tracking-widest text-2xl border-4 border-black mb-8 animate-bounce">
              {game.winner_id === myUserId ? "🏆 VICTORY!" : "☠️ DEFEAT!"}
            </div>
          )}
          {game.status === "DRAW" && (
            <div className="w-full bg-gray-400 text-black text-center p-4 font-black uppercase tracking-widest text-2xl border-4 border-black mb-8">
              🤝 IT'S A DRAW
            </div>
          )}

          {/* ป้ายบอกเทิร์น (ถ้าเกมยังไม่จบ) */}
          {game.status === "IN_PROGRESS" && (
            <div className={`w-full text-center p-3 font-black uppercase tracking-widest text-xl border-4 mb-8 transition-colors ${
              isMyTurn ? "bg-black text-white border-black" : "bg-white text-gray-400 border-gray-300"
            }`}>
              {isMyTurn ? "🔥 YOUR TURN" : "OPPONENT'S TURN..."}
            </div>
          )}

          {/* กระดาน Tic-Tac-Toe */}
          <div className="grid grid-cols-3 gap-2 bg-black p-2 border-4 border-black shadow-[8px_8px_0px_0px_rgba(220,38,38,1)]">
            {game.board.split("").map((cell, index) => {
              const isX = cell === "X";
              const isO = cell === "O";
              const isEmpty = cell === "-";
              
              return (
                <button
                  key={index}
                  onClick={() => handleMove(index)}
                  disabled={!isEmpty || !isMyTurn || game.status !== "IN_PROGRESS"}
                  className={`w-24 h-24 flex items-center justify-center text-5xl font-black transition-all ${
                    isEmpty ? "bg-white hover:bg-gray-200 cursor-pointer" : "bg-gray-100 cursor-not-allowed"
                  } ${isX ? "text-red-600" : isO ? "text-blue-600" : ""}`}
                >
                  {isEmpty ? "" : cell}
                </button>
              );
            })}
          </div>

          {/* ปุ่มออกจากห้อง */}
          <button
            onClick={() => router.push("/lobby")}
            className="mt-12 text-gray-500 font-bold uppercase tracking-widest text-sm hover:text-red-600 transition-colors underline underline-offset-4"
          >
            Leave Arena
          </button>
        </div>
      )}

    </div>
  );
}