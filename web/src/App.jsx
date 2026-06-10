import { useState } from "react";
import LoginScreen from "./screens/LoginScreen";
import MenuScreen from "./screens/MenuScreen";
import OrdersScreen from "./screens/OrdersScreen";
import StatsScreen from "./screens/StatsScreen";

const TOKEN_KEY = "orderflow_token";

export default function App() {
  const [token, setToken] = useState(() => localStorage.getItem(TOKEN_KEY) || "");
  const [screen, setScreen] = useState("menu");

  function handleLogin(newToken) {
    localStorage.setItem(TOKEN_KEY, newToken);
    setToken(newToken);
  }

  function handleLogout() {
    localStorage.removeItem(TOKEN_KEY);
    setToken("");
  }

  function handleApiError(error) {
    // 401 derruba a sessão
    if (error && error.status === 401) {
      handleLogout();
    }
  }

  if (!token) {
    return <LoginScreen onLogin={handleLogin} />;
  }

  return (
    <div className="app">
      <header className="app-header">
        <h1>OrderFlow</h1>
        <nav>
          <button className={screen === "menu" ? "active" : ""} onClick={() => setScreen("menu")}>
            Cardápio
          </button>
          <button className={screen === "orders" ? "active" : ""} onClick={() => setScreen("orders")}>
            Pedidos
          </button>
          <button className={screen === "stats" ? "active" : ""} onClick={() => setScreen("stats")}>
            Stats
          </button>
          <button className="logout" onClick={handleLogout}>
            Sair
          </button>
        </nav>
      </header>
      <main>
        {screen === "menu" && <MenuScreen token={token} onApiError={handleApiError} />}
        {screen === "orders" && <OrdersScreen onApiError={handleApiError} />}
        {screen === "stats" && <StatsScreen onApiError={handleApiError} />}
      </main>
    </div>
  );
}
