"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

export default function AuthPage() {
  const router = useRouter()
  const [isLogin, setIsLogin] = useState(true);
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  
  //ดึง url จาก backend
  const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

  const handleSubmit = async (e: React.SyntheticEvent<HTMLFormElement>) => {
    e.preventDefault();
    setError("");

    if (!isLogin && password !== confirmPassword) {
      setError("Passwords do not match!");
      return;
    }

    setLoading(true);

  const endpoint = isLogin ? "/api/login" : "/api/register";

  try {
    const res = await fetch(`${API_URL}${endpoint}`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ username,password })
    })

    const data = await res.json();

    if (!res.ok) {
      throw new Error(data.error || "Something went wrong");
    }

    if(isLogin) {
      // ถ้า Login สำเร็จ เราต้องเก็บ Token และ UserID ไว้ใช้ต่อ
      localStorage.setItem("token", data.token);
      localStorage.setItem("user_id", data.user_id.toString());
      localStorage.setItem("username", username);

      router.push("/lobby");
    } else {
      alert("Registration successful! Please login.");
      setIsLogin(true);
      setPassword("");
      setConfirmPassword("");
    }
  } catch (err: any) {
    setError(err.message);
  } finally {
    setLoading(false)
  }
  };
  return (
    <div className="min-h-screen flex items-center justify-center bg-white text-black font-sans selection:bg-red-500 selection:text-white">
      <div className="bg-black text-white p-10 rounded-none shadow-[8px_8px_0px_0px_rgba(220,38,38,1)] border-4 border-black w-full max-w-md">
        <div className="text-center mb-8">
          <h1 className="text-4xl font-black tracking-tighter uppercase text-white mb-2">
            Tic<span className="text-red-600">-</span>Tac<span className="text-red-600">-</span>Toe
          </h1>
          <p className="text-gray-400 font-bold uppercase tracking-widest text-sm">
            Arena
          </p>
        </div>

        <h2 className="text-2xl font-bold mb-6 text-center uppercase border-b-2 border-red-600 pb-2 inline-block w-full">
          {isLogin ? "Sign In" : "Register"}
        </h2>

        <form onSubmit={handleSubmit} className="space-y-5">
          <div>
            <label className="block text-sm font-bold text-gray-300 mb-1 uppercase tracking-wide">
              Username
            </label>
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className="w-full px-4 py-3 bg-white text-black font-bold border-2 border-white focus:outline-none focus:border-red-600 transition-colors"
              placeholder="PLAYER_1"
              required
              minLength={3}
              maxLength={20}
            />
          </div>

          <div>
            <label className="block text-sm font-bold text-gray-300 mb-1 uppercase tracking-wide">
              Password
            </label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full px-4 py-3 bg-white text-black font-bold border-2 border-white focus:outline-none focus:border-red-600 transition-colors"
              placeholder="••••••••"
              required
              minLength={6}
            />
          </div>

          {!isLogin && (
            <div className="animate-fade-in">
              <label className="block text-sm font-bold text-gray-300 mb-1 uppercase tracking-wide">
                Confirm Password
              </label>
              <input
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                className="w-full px-4 py-3 bg-white text-black font-bold border-2 border-white focus:outline-none focus:border-red-600 transition-colors"
                placeholder="••••••••"
                required={!isLogin}
                minLength={6}
              />
            </div>
          )}

          {error && (
            <div className="p-3 bg-red-600 text-white font-bold text-sm text-center border-2 border-red-800">
              {error}
            </div>
          )}

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-red-600 hover:bg-red-700 text-white font-black uppercase tracking-widest py-4 px-4 border-2 border-red-600 transition-all transform active:scale-95 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {loading ? "Processing..." : isLogin ? "Enter Match" : "Create Profile"}
          </button>
        </form>

        <div className="mt-8 text-center">
          <button
            type="button"
            onClick={() => {
              setIsLogin(!isLogin);
              setError("");
              setConfirmPassword("");
            }}
            className="text-gray-400 hover:text-red-500 font-bold uppercase text-sm transition-colors"
          >
            {isLogin
              ? "New Challenger? Register"
              : "Already registered? Sign In"}
          </button>
        </div>
      </div>
    </div>
  );
}
