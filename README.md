# Hyperliquid Trade Reconciliation
A lightweight trade reconciliation system for Hyperliquid exchange that fetches historical trades and calculates daily P&L.

## Features
### Core Functionality
- **Incremental Trade Caching**: Smart caching system that only fetches new trades after initial load
- **Flexible Time Ranges**: Support for 24H, 7D, 30D, and 90D historical data
- **Daily P&L Calculation**: Automatic calculation with cumulative tracking
- **Multi-Account Support**: Switch between multiple accounts with dropdown selector
- **Fast Auto-Refresh**: Updates every 10 seconds using incremental fetching (configurable)

### Performance & Optimization
- **Intelligent Caching**: Caches trades per account, only fetching new data on refresh
- **Reduced API Usage**: Up to 90% fewer API calls after initial load
- **Smart Rate Limiting**: 300ms delay between API batches to prevent throttling
- **Pagination Handling**: Automatic handling of accounts with >2000 trades

## Project Structure

```
hyperliquid-trade-recon/
├── backend/
│   ├── api/              # HTTP handlers
│   ├── config/           # Configuration constants
│   ├── models/           # Data models
│   ├── services/         # Business logic
│   ├── main.go           # Entry point
│   └── go.mod
├── frontend/
│   ├── src/
│   │   ├── components/   # React components
│   │   ├── config/       # Frontend configuration
│   │   ├── services/     # API services
│   │   ├── App.js
│   │   └── App.css
│   ├── public/
│   └── package.json
└── README.md
```
## Installation
### Backend Setup
```
cd backend/
go mod download
go mod verify
```
### Frontend Setup
```
cd frontend/
npm install
```

## Running the Application
### Option 1: Development Mode (Recommended)

#### Start Backend Server

```
cd backend/
CGO_ENABLED=0 go run main.go
```

The backend server will start on `http://localhost:8080`

#### Start Frontend Development Server

```
cd frontend/
npm start
```
The frontend will start on `http://localhost:3000` and automatically open in your browser.

### Option 2: Single Executable - PRODUCTION MODE

Build everything into a single executable file - no external dependencies needed!

**Linux / macOS:**
```
./build.sh
./hyperliquid-recon
```

**Windows:**
```
build.bat
hyperliquid-recon.exe
```
Then open your browser at `http://localhost:8080`

## API Endpoints

### GET `/api/health`
Health check endpoint

**Response:**
```json
{
  "status": "ok"
}
```

### GET `/api/pnl`
Get P&L summary

**Response:**
```json
{
  "dailyRecords": [
    {
      "date": "2025-01-28",
      "tradeCount": 42,
      "dailyPnL": 1234.56,
      "cumulativePnL": 5678.90
    }
  ],
  "totalPnL": 5678.90
}
```

### POST `/api/refresh?address={address}&timeRange={days}`
Trigger data refresh for a specific account

**Parameters:**
- `address` (query, required): Ethereum address of the account
- `timeRange` (query, optional): Number of days to fetch (default: 30). Supported values: 1, 7, 30, 90

**Response:**
```json
{
  "message": "Data refreshed successfully"
}
```

**Example:**
```
curl -X POST "http://localhost:8080/api/refresh?address=0x091144e651b334341eabdbbbfed644ad0100023e&timeRange=30"
```

## Features in Detail

### Trade Fetching
- Fetches trades with flexible time ranges: 1, 7, 30, or 90 days
- Handles pagination for accounts with >2000 trades
- Rate limiting: 200ms delay between batch requests
- Aggregates trades by time for efficient processing

### P&L Calculation
- Groups trades by date and coin
- Calculates daily P&L: (Total Sells Value - Total Buys Value)
- Tracks cumulative P&L over time
- Supports multiple trading pairs

### UI Features
- Responsive layout
- Loading states and error handling
- Account switching without page reload
- Table pagination with configurable rows per page

### Data Refresh Mechanism
The application uses a **client-side polling** architecture combined with **backend incremental caching**. There is **NO backend cron job** - all refresh triggers come from the frontend.

#### Refresh Triggers

1. **Automatic Refresh (Frontend Timer)**
   - Frontend uses `setInterval` to poll every 10 seconds by default (configurable)
   - Timer runs continuously while browser tab is open
   - Automatically fetches latest data from backend
   - Location: `frontend/src/App.js:69-73`
   - Configuration: `AUTO_REFRESH_INTERVAL_MS` in `frontend/src/config/config.js`

2. **Manual Refresh**
   - User clicks "Refresh Data" button
   - User changes account selection
   - User changes time range filter
   - Each trigger immediately fetches fresh data

3. **Initial Load**
   - Data fetches automatically when page first loads
   - Uses default account and time range settings

#### How It Works

```
┌─────────────┐                    ┌──────────────┐                    ┌─────────────────┐
│   Frontend  │                    │   Backend    │                    │  Hyperliquid    │
│  (Browser)  │                    │   (Go API)   │                    │      API        │
└─────────────┘                    └──────────────┘                    └─────────────────┘
       │                                   │                                     │
       │ 1. Timer triggers (every 10s)     │                                     │
       ├──────────────────────────────────>│                                     │
       │  POST /api/refresh?address=...&   │                                     │
       │       timeRange=30                │                                     │
       │                                   │ 2. Check cache & fetch incrementally│
       │                                   │    (only new trades if cached)      │
       │                                   ├────────────────────────────────────>│
       │                                   │    POST /info (userFillsByTime)     │
       │                                   │<────────────────────────────────────│
       │                                   │    Returns new trades only          │
       │                                   │                                     │
       │                                   │    (200ms delay - rate limiting)    │
       │                                   │                                     │
       │                                   │ 3. Merge with cached trades         │
       │                                   │    - Deduplicate by trade key       │
       │                                   │    - Sort by timestamp              │
       │                                   │                                     │
       │                                   │ 4. Calculate P&L                    │
       │                                   │    - Group by date & coin           │
       │                                   │    - SellValue - BuyValue           │
       │                                   │    - Store in memory                │
       │<──────────────────────────────────│                                     │
       │  200 OK                           │                                     │
       │                                   │                                     │
       │ 5. Get calculated data            │                                     │
       ├──────────────────────────────────>│                                     │
       │  GET /api/pnl                     │                                     │
       │<──────────────────────────────────│                                     │
       │  Returns cached P&L summary       │                                     │
       │                                   │                                     │
       │ (10 seconds pass...)              │                                     │
       │                                   │                                     │
       │ 6. Timer triggers again           │                                     │
       ├──────────────────────────────────>│                                     │
       │  Repeat cycle...                  │                                     │
```



## Design Decisions

### Backend Architecture
- **Service Layer Pattern**: Separated business logic into services (`HyperliquidService`, `ReconciliationService`) for better testability and maintainability
- **Incremental Caching System**: Implements per-account caching with intelligent full vs incremental fetching
  - Caches trades in memory with last fetch timestamp
  - Only fetches new trades since last fetch (< 1 hour)
  - Automatically merges and deduplicates trades
  - Reduces API calls by up to 90% after initial load
- **In-Memory Data Storage**: Current implementation stores reconciliation data in memory. Suitable for lightweight applications; database integration recommended for production
- **Pagination Strategy**: Implemented batch fetching with `startTime` parameter to handle accounts with >2000 trades, avoiding API limitations
- **Smart Rate Limiting**: 300ms delay between API calls to respect Hyperliquid's rate limits and prevent throttling

### Frontend Architecture
- **Component-Based Design**: Separated concerns into reusable components (`PnLTable`, `TimeRangeSelector`)
- **Centralized Configuration**: All configurable values (accounts, API URLs, intervals) are in `config.js` for easy modification
- **Auto-Refresh Pattern**: Implemented using `useEffect` with cleanup to prevent memory leaks
- **State Management**: Used React Hooks for local state management, avoiding unnecessary complexity of Redux for this scale
- **Monospace Fonts for Numbers**: Used 'Courier New' for financial data to improve readability and align decimal points


## Assumptions
- **P&L Calculation**: Assumes simple buy/sell logic without considering:
  - Fees and commissions
  - Funding rates for perpetual contracts
  - Unrealized P&L from open positions
- **USD Denomination**: All P&L values are assumed to be in USD
- **Trade Finality**: All fetched trades are assumed to be settled and final
- **Single User**: Application assumes single-user usage (no authentication/authorization)
- **Time Zone**: All timestamps are processed in the system's local time zone

## Possible Improvements
### Backend Enhancements
- **Persistent Caching**: Extend current in-memory caching with Redis for persistence across restarts
- **Websocket Support**: Add real-time updates via WebSocket instead of polling
- **Rate Limiting**: Implement token bucket rate limiting for API endpoints
- **Error Handling**: More granular error types and better error messages
- **Unit Tests**: Add comprehensive unit tests for services and handlers

### Frontend Enhancements
- **TypeScript**: Convert to TypeScript for better type safety
- **Advanced Filtering**: Filter by coin, date range, profit/loss only
- **Dark/Light Mode Toggle**: User-selectable theme preference

## Limitations
- **No data when browsers closed**: If all users close tabs, data stops updating
- **Memory-only storage**: Data lost on server restart
- **Multiple requests**: Each open browser makes separate API calls
- **Manual initial load**: First page load requires manual data fetch


Authored by, Dilshan Harshana