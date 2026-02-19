"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

export default function AuthPage() {
  const router = useRouter()
  const [isLogin, setIsLogin] = useState(true);

  // States สำหรับเก็บข้อมูลฟอร์ม
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");

  // State สำหรับ UX และ UI
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  
  //ดึง url จาก backend
  const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

  const validateForm = () => {
    // regex username
    const alphanumericRegex = /^[a-zA-Z0-9]+$/;
    if (username.length < 3 || username.length > 20){
      return "Username must be between 3 and 20 characters.";
    }

    if (!alphanumericRegex.test(username)) {
      return "Username can only contain letters and numbers.";
    }

    //validate password
    if (password.length < 6) {
      return "Password must be at least 6 characters long.";
    }

    // confirmation password 
    if (!isLogin && password !== confirmPassword) {
      return "Passwords do not match!";
    }

    return null
  }

  const handleSubmit = async (e: React.SyntheticEvent<HTMLFormElement>) => {
    e.preventDefault();
    setError("");

    const validationError = validateForm();
    if (validationError) {
      setError(validationError);
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
        throw new Error(data.error || "Something went wrong with the server.");
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
        setShowPassword(false);
      }
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false)
    }
  };

  const EyeIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-5 h-5">
      <path strokeLinecap="round" strokeLinejoin="round" d="M2.036 12.322a1.012 1.012 0 010-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178z" />
      <path strokeLinecap="round" strokeLinejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
    </svg>
  );

  const EyeSlashIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-5 h-5">
      <path strokeLinecap="round" strokeLinejoin="round" d="M3.98 8.223A10.477 10.477 0 001.934 12C3.226 16.338 7.244 19.5 12 19.5c.993 0 1.953-.138 2.863-.395M6.228 6.228A10.45 10.45 0 0112 4.5c4.756 0 8.773 3.162 10.065 7.498a10.523 10.523 0 01-4.293 5.774M6.228 6.228L3 3m3.228 3.228l3.65 3.65m7.894 7.894L21 21m-3.228-3.228l-3.65-3.65m0 0a3 3 0 10-4.243-4.243m4.242 4.242L9.88 9.88" />
    </svg>
  );

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
            />
          </div>

          <div className="relative">
            <label className="block text-sm font-bold text-gray-300 mb-1 uppercase tracking-wide">
              Password
            </label>
            <div className="relative">
              <input
                type={showPassword ? "text" : "password"}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full px-4 py-3 pr-12 bg-white text-black font-bold border-2 border-white focus:outline-none focus:border-red-600 transition-colors"
                placeholder="••••••••"
              />
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                className="absolute inset-y-0 right-0 px-3 flex items-center text-gray-500 hover:text-red-600 transition-colors"
              >
                {showPassword ? <EyeSlashIcon /> : <EyeIcon />}
              </button>
            </div>
          </div>

          {!isLogin && (
            <div className="animate-fade-in relative">
              <label className="block text-sm font-bold text-gray-300 mb-1 uppercase tracking-wide">
                Confirm Password
              </label>
              <div className="relative">
                <input
                  type={showPassword ? "text" : "password"}
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  className="w-full px-4 py-3 pr-12 bg-white text-black font-bold border-2 border-white focus:outline-none focus:border-red-600 transition-colors"
                  placeholder="••••••••"
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute inset-y-0 right-0 px-3 flex items-center text-gray-500 hover:text-red-600 transition-colors"
                >
                  {showPassword ? <EyeSlashIcon /> : <EyeIcon />}
                </button>
              </div>
            </div>
          )}

          {error && (
            <div className="p-3 bg-red-600 text-white font-bold text-sm text-center border-2 border-red-800 animate-pulse">
              {error}
            </div>
          )}

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-red-600 hover:bg-red-700 text-white font-black uppercase tracking-widest py-4 px-4 border-2 border-red-600 transition-all transform active:scale-95 disabled:opacity-50 disabled:cursor-not-allowed mt-2"
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
              setShowPassword(false);
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
