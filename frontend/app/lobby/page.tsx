"use client";

import { useState,useEffect } from "react";
import { useRouter } from "next/navigation";
import { RouteMatcher } from "next/dist/server/route-matchers/route-matcher";

export default function LobbyPage() {
    const router = useRouter()
    const [username, setUsername] = useState("");
    const [joinRoomId, setJoinRoomId] = useState("");
    const [error, setError] = useState("");
    const [loading, setLoading] = useState(false);

    const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

    //เช็คว่ามี token มั้ย
    useEffect(() => {
        const token = localStorage.getItem("token");
        const storedUser = localStorage.getItem("username");

        if (!token) {
            router.push("/");
        } else {
            setUsername(storedUser || "Player")
        }
    }, [router]);

    // เช็คว่ามีเกมที่ค้างอยู่ (Zombie Session) ไหม ถ้ามีให้ดึงตัวกลับไป
    useEffect(() => {
      const checkZombieSession = async () => {
        const token = localStorage.getItem("token");
          if (!token) return;
          
          try {
            const res = await fetch(`${API_URL}/api/games/me/active`, {
              headers: { Authorization: `Bearer ${token}` },
            })

            if (!res.ok) return;

            const data = await res.json();
            if(data.has_active_game) {
              alert("You have an ongoing match! Reconnecting to the arena...");
              router.push(`/game/${data.room_code}`);
            }
          } catch (err:any) {
            console.error("Failed to check active game", err);
          }
      };
      checkZombieSession();
    }, [router, API_URL]);

    //logout 
    const handleLogout = () => {
        localStorage.clear();;
        router.push("/");
    };

    // สร้างห้องใหม่
    const handleCreateRoom = async () => {
        setError("")
        setLoading(true);
        const token = localStorage.getItem("token")

        try {
            const res = await fetch(`${API_URL}/api/games`, {
                method: "POST",
                headers: {
                "Content-Type": "application/json",
                "Authorization": `Bearer ${token}`,
                },
            })

            const data = await res.json()

            if (!res.ok){
                throw new Error(data.error || "Failed to create room")
            }

            //ได้ game id มาเข้าไปหน้ากระดานเกม
            router.push(`/game/${data.room_code}`);
        } catch (err: any) {
            setError(err.message)
            setLoading(false)
        }
    };

    //เข้าร่วมห้อง
    const handleJoinRoom = async (e: React.SyntheticEvent<HTMLFormElement>) => {
        e.preventDefault()
        setError("")

        if (!joinRoomId){
            setError("Please enter a Room ID");
            return;
        }

        setLoading(true);
        const token = localStorage.getItem("token")

        try {
            const res = await fetch(`${API_URL}/api/games/join`, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                    "Authorization": `Bearer ${token}`,
                },
                body: JSON.stringify({ room_code: joinRoomId }),
            });

            const data = await res.json();

            if (!res.ok) {
              //ถ้าห้องเต็มเข้าไปเป็นผู้ชมได้
              if (res.status === 409) {
                router.push(`/game/${joinRoomId}`);
                return;
              }
              throw new Error(data.error || "Failed to join room");
            }

            //ถ้าสำเร็จ 
            router.push(`/game/${joinRoomId}`);
        } catch (err:any) {
            setError(err.message);
            setLoading(false);
        }
    }
    return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-white text-black font-sans selection:bg-red-500 selection:text-white p-4">
      
      {/* ส่วนหัวแสดงชื่อผู้เล่นและปุ่ม Logout */}
      <div className="absolute top-4 right-4 flex items-center space-x-4">
        <span className="font-bold uppercase tracking-widest text-sm border-b-2 border-black">
          CMD: {username}
        </span>
        <button
          onClick={handleLogout}
          className="text-xs font-bold uppercase tracking-widest bg-black text-white px-3 py-1 hover:bg-red-600 transition-colors"
        >
          Logout
        </button>
      </div>

      <div className="bg-black text-white p-10 rounded-none shadow-[8px_8px_0px_0px_rgba(220,38,38,1)] border-4 border-black w-full max-w-md">
        
        <div className="text-center mb-10">
          <h1 className="text-4xl font-black tracking-tighter uppercase text-white mb-2">
            Match<span className="text-red-600">Making</span>
          </h1>
          <p className="text-gray-400 font-bold uppercase tracking-widest text-sm">
            Select Your Path
          </p>
        </div>

        {error && (
          <div className="mb-6 p-3 bg-red-600 text-white font-bold text-sm text-center border-2 border-red-800 animate-pulse">
            {error}
          </div>
        )}

        <div className="space-y-8">
          
          {/* โซนสร้างห้อง */}
          <div className="border-2 border-white p-6 relative">
            <span className="absolute -top-3 left-4 bg-black px-2 text-xs font-bold uppercase tracking-widest text-red-500">
              Option 1
            </span>
            <button
              onClick={handleCreateRoom}
              disabled={loading}
              className="w-full bg-white text-black font-black uppercase tracking-widest py-4 border-2 border-white hover:bg-red-600 hover:text-white hover:border-red-600 transition-all transform active:scale-95 disabled:opacity-50"
            >
              {loading ? "Initializing..." : "Create New Room"}
            </button>
            <p className="text-center text-xs text-gray-400 mt-3 uppercase tracking-wide">
              Host a game and wait for an opponent
            </p>
          </div>

          <div className="text-center font-black text-gray-600 uppercase">OR</div>

          {/* โซนเข้าร่วมห้อง */}
          <form onSubmit={handleJoinRoom} className="border-2 border-white p-6 relative">
            <span className="absolute -top-3 left-4 bg-black px-2 text-xs font-bold uppercase tracking-widest text-red-500">
              Option 2
            </span>
            <div className="space-y-3">
              <input
                type="text"
                maxLength={6}
                value={joinRoomId}
                onChange={(e) => setJoinRoomId(e.target.value)}
                placeholder="ENTER ROOM ID"
                className="w-full px-4 py-3 bg-white text-black font-bold border-2 border-white focus:outline-none focus:border-red-600 text-center tracking-widest"
                required
              />
              <button
                type="submit"
                disabled={loading}
                className="w-full bg-red-600 text-white font-black uppercase tracking-widest py-4 border-2 border-red-600 hover:bg-red-700 transition-all transform active:scale-95 disabled:opacity-50"
              >
                {loading ? "Connecting..." : "Join Room"}
              </button>
            </div>
          </form>

        </div>
      </div>
    </div>
  );
}