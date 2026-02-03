# Mobile App Development Plan

## Goal
Build a "decent" mobile version of the Solana Watcher to monitor new coin releases, detect scams, and manage positions on the go.

## Architecture Proposal: Client-Server Model
Running the full watcher logic (RPC tailing, heavy filtering) directly on a mobile device is battery and data intensive. A **Client-Server** architecture is recommended:
- **Backend (Existing Go App)**: Refactor to run in "Headless Mode" (no TUI) on a server/desktop. Expose data via an API (REST + WebSockets).
- **Mobile App (Client)**: Connects to the backend to display feed, receive alerts, and send commands (buy/sell/config).

## Tech Stack
- **Mobile**: React Native (TypeScript) or Flutter. (React Native is recommended if you want to share logic with a potential future Web Dashboard).
- **Backend Communication**:
    - **WebSockets**: For real-time coin feeds and alerts.
    - **REST/gRPC**: For configuration and historical data.
- **Notifications**: Firebase Cloud Messaging (FCM) or local polling (if background execution is permitted).

## Development Roadmap

### Phase 1: Backend Refactoring (Go)
The current monolithic `main.go` needs to be modularized.
- [ ] **Decouple Logic**: Extract `RPC Client`, `Analysis Engine`, and `State` into separate packages (`/pkg/...`).
- [ ] **Headless Mode**: Add a flag (e.g., `--headless` or `--server`) to run without the TUI.
- [ ] **API Layer**: Implement a lightweight HTTP/WebSocket server (using `Gin` or standard `net/http` + `gorilla/websocket`) to stream events.
    - Endpoint: `GET /ws` (Stream `chainDataMsg`, `tickMsg`, alerts).
    - Endpoint: `GET /api/status` (Current health, RPC status).

### Phase 2: Mobile App Scaffolding
- [ ] **Init Project**: Initialize React Native project (Expo recommended for ease of use).
- [ ] **Navigation**: Set up tab navigation (Feed, Portfolio, Settings).
- [ ] **Network Layer**: Create a WebSocket client to connect to the Go backend.

### Phase 3: Core Features (The Watcher)
- [ ] **Live Feed**:
    - Display a scrolling list of new coins/events.
    - Highlight "Safe" vs "Scam" (color-coded cards).
    - Show key metrics (Liquidity, Holders, Age).
- [ ] **Push Notifications**:
    - "High Confidence" Gem alerts.
    - Rugpull warnings for held coins.

### Phase 4: Control & Trading
- [ ] **Configuration**: UI to update `config.json` settings (RPC endpoints, risk thresholds) remotely.
- [ ] **Actions**:
    - "Snipe" button (triggers buy on backend).
    - "Panic Sell" button.

## TODO Checklist

### Backend (Go)
- [ ] Audit `main.go` to identify state structs to export.
- [ ] Create `server.go` to handle WebSocket connections.
- [ ] Define JSON schemas for WebSocket messages.

### Frontend (Mobile)
- [ ] Sketch UI Mockups (Feed Card, Detail View).
- [ ] Implement robust reconnection logic (mobile networks are flaky).
- [ ] functionality for "Background Fetch" if we want local alerts without a central push server.

## Future Considerations
- **Authentication**: Secure the WebSocket connection (API Key or Basic Auth) so only you can access your watcher.
- **Cloud Hosting**: Deploy the backend to a VPS (AWS/DigitalOcean) for 24/7 uptime.
