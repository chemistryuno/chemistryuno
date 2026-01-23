// -*- coding: utf-8 -*-
import express, { Request, Response, NextFunction } from 'express';
import cors from 'cors';
import http from 'http';
import { Server as SocketIOServer } from 'socket.io';
import QRCode from 'qrcode';
import fs from 'fs';
import path from 'path';
import * as gameLogic from './gameLogic';
import database = require('./database');
import GameRules = require('./rules');
import configService = require('./configService');

const app = express();
const server = http.createServer(app);

// 辅助函数：确定项目根目录 (提升到顶部以供全局使用)
const getRootDir = () => {
  const isDist = __dirname.endsWith('dist') || __dirname.includes(path.join('server', 'dist'));
  // 如果在 server/dist 下 -> 回退两级到 chemistryuno/
  // 如果在 server 下 (dev) -> 回退一级到 chemistryuno/
  return isDist ? path.join(__dirname, '..', '..') : path.join(__dirname, '..');
};

// 配置CORS - 支持移动设备访问
const allowedOrigins: (string | RegExp)[] = [
  'http://localhost:3000',
  'http://127.0.0.1:3000',
  'http://localhost:4000',
  'http://127.0.0.1:4000',
  // 支持局域网IP访问
  /^http:\/\/192\.168\.\d+\.\d+:(3000|4000)$/,
  /^http:\/\/10\.\d+\.\d+\.\d+:(3000|4000)$/,
  /^http:\/\/172\.(1[6-9]|2\d|3[01])\.\d+\.\d+:(3000|4000)$/,
  // 支持任意IP地址（开发和生产环境）
  /^http:\/\/[\d.]+(:(3000|4000))?$/,
  /^https?:\/\/[\w.-]+(:(3000|4000|80|443))?$/
];

const io = new SocketIOServer(server, {
  cors: {
    origin: (origin, callback) => {
      // 允许没有origin的请求（如移动应用、Postman等）
      if (!origin) return callback(null, true);
      
      // 检查origin是否在允许列表中
      const isAllowed = allowedOrigins.some(allowed => {
        if (typeof allowed === 'string') {
          return allowed === origin;
        } else if (allowed instanceof RegExp) {
          return allowed.test(origin);
        }
        return false;
      });
      
      if (isAllowed) {
        callback(null, true);
      } else {
        
        callback(null, true); // 开发环境仍然允许，生产环境应该设为 false
      }
    },
    methods: ["GET", "POST"],
    credentials: true
  }
});

app.use(cors({
  origin: (origin, callback) => {
    // 允许所有来源（开发环境）
    callback(null, true);
  },
  credentials: true
}));
app.use(express.json({ limit: '1mb' }));

// 类型定义
interface Player {
  id: number;
  name: string;
  hand: string[];
  compounds: string[];
  isHost: boolean;
  isOffline?: boolean;
}

interface Spectator {
  id: number;
  name: string;
  joinedAt: string;
}

interface GameState {
  roomCode: string;
  players: Player[];
  spectators: Spectator[];
  maxPlayers: number;
  deck: string[];
  currentPlayer: number;
  direction: number;
  lastCompound: string | null;
  lastCard?: string | null;
  gameStarted: boolean;
  gameActive: boolean;
  history: any[];
  pendingDraws: number;
  createdAt: string;
}

interface SocketPlayerInfo {
  roomCode: string;
  playerId: number | null;
  playerName: string;
  isHost: boolean;
  isSpectator: boolean;
}

interface PendingCleanup {
  isHost: boolean;
  playerName: string;
  timeoutId: NodeJS.Timeout;
}

interface PlayerSession {
  roomCode: string;
  playerId: number;
  joinTime: string;
}

// 根路由 - 服务器状态检查
app.get('/', (req: Request, res: Response) => {
  const html = `
    <!DOCTYPE html>
    <html lang="zh-CN">
    <head>
      <meta charset="UTF-8">
      <meta name="viewport" content="width=device-width, initial-scale=1.0">
      <title>化学UNO - Chemistry UNO Game Server</title>
      <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
          font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
          background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
          min-height: 100vh;
          display: flex;
          align-items: center;
          justify-content: center;
          padding: 20px;
        }
        .container {
          background: white;
          border-radius: 12px;
          box-shadow: 0 20px 60px rgba(0,0,0,0.3);
          padding: 40px;
          max-width: 800px;
          width: 100%;
        }
        h1 {
          color: #667eea;
          margin-bottom: 10px;
          font-size: 2.5em;
        }
        .subtitle {
          color: #666;
          margin-bottom: 30px;
          font-size: 1.1em;
        }
        .status {
          display: flex;
          align-items: center;
          margin-bottom: 20px;
          padding: 15px;
          background: #f0f9ff;
          border-radius: 8px;
          border-left: 4px solid #667eea;
        }
        .status-indicator {
          width: 12px;
          height: 12px;
          background: #22c55e;
          border-radius: 50%;
          margin-right: 10px;
          animation: pulse 2s infinite;
        }
        @keyframes pulse {
          0%, 100% { opacity: 1; }
          50% { opacity: 0.5; }
        }
        .endpoints {
          margin-top: 30px;
        }
        .endpoints h2 {
          color: #333;
          margin-bottom: 15px;
          font-size: 1.3em;
        }
        .endpoint-item {
          background: #f5f5f5;
          padding: 12px 15px;
          margin-bottom: 10px;
          border-radius: 6px;
          border-left: 3px solid #667eea;
          font-family: 'Courier New', monospace;
          font-size: 0.9em;
        }
        .method {
          display: inline-block;
          background: #667eea;
          color: white;
          padding: 2px 8px;
          border-radius: 4px;
          margin-right: 10px;
          font-weight: bold;
          min-width: 45px;
          text-align: center;
        }
        .frontend-link {
          display: inline-block;
          margin-top: 30px;
          padding: 12px 24px;
          background: #667eea;
          color: white;
          text-decoration: none;
          border-radius: 8px;
          font-weight: bold;
          transition: all 0.3s;
        }
        .frontend-link:hover {
          background: #764ba2;
          transform: translateY(-2px);
          box-shadow: 0 10px 20px rgba(102, 126, 234, 0.3);
        }
        .info-box {
          background: #fef3c7;
          padding: 15px;
          border-radius: 8px;
          margin-top: 20px;
          border-left: 4px solid #f59e0b;
        }
        .info-box strong {
          color: #d97706;
        }
      </style>
    </head>
    <body>
      <div class="container">
        <h1>⚗️ 化学UNO 游戏服务器</h1>
        <p class="subtitle">Chemistry UNO Game Server v1.0.0</p>
        
        <div class="status">
          <div class="status-indicator"></div>
          <span style="font-weight: bold; color: #22c55e;">服务器运行中</span>
        </div>

        <div class="endpoints">
          <h2>📡 可用API接口</h2>
          <div class="endpoint-item"><span class="method">GET</span> /api/compounds</div>
          <div class="endpoint-item"><span class="method">POST</span> /api/game/create</div>
          <div class="endpoint-item"><span class="method">GET</span> /api/game/:gameId/:playerId</div>
          <div class="endpoint-item"><span class="method">POST</span> /api/reaction/check</div>
          <div class="endpoint-item"><span class="method">GET</span> /api/game/:gameId/stats</div>
          <div class="endpoint-item"><span class="method">WS</span> WebSocket - 实时通信</div>
        </div>

        <div class="info-box">
          <strong>⚠️ 注意：</strong> 前端应在 <a href="http://localhost:4000" style="color: #d97706; font-weight: bold;">http://localhost:4000</a> 运行。
          如看不到游戏界面，请确保执行了 <code>npm start</code> 命令。
        </div>

        <a href="http://localhost:4000" class="frontend-link">▶️ 进入游戏</a>
      </div>
    </body>
    </html>
  `;
  res.setHeader('Content-Type', 'text/html; charset=utf-8');
  res.send(html);
});

// 存储游戏会话
const gameSessions = new Map<string, GameState>();
const playerSockets = new Map<string, string>(); // playerName -> socketId
const socketToPlayer = new Map<string, SocketPlayerInfo>(); // socketId -> player info
const pendingCleanup = new Map<string, PendingCleanup>(); // roomCode -> cleanup info
const playerToRoom = new Map<string, PlayerSession>(); // playerName -> session info

// 获取游戏设置
function getGameSettings() {
  const config = configService.getConfig();
  return config.game_settings || {
    reconnect_timeout: 30000,
    host_timeout: 30000
  };
}

// 生成6位房间号
function generateRoomCode(): string {
  return Math.floor(100000 + Math.random() * 900000).toString();
}

// 确保房间号唯一
function generateUniqueRoomCode(): string {
  let code: string;
  do {
    code = generateRoomCode();
  } while (gameSessions.has(code));
  return code;
}

// 路由：创建新游戏
app.post('/api/game/create', (req: Request, res: Response) => {
  const { playerName } = req.body;
  
  // 生成唯一房间号
  const roomCode = generateUniqueRoomCode();
  
  // 初始化游戏状态（最多12人）
  const gameState: GameState = {
    roomCode: roomCode,
    players: [{
      id: 0,
      name: playerName,
      hand: [],
      compounds: [],
      isHost: true
    }],
    spectators: [],  // 观战者列表
    maxPlayers: 12,
    deck: [],
    currentPlayer: 0,
    direction: 1,
    lastCompound: null,
    gameStarted: false,
    gameActive: false,
    history: [],  // 游戏历史记录
    pendingDraws: 0,  // 累加的抽牌数
    createdAt: new Date().toISOString()
  };
  
  gameSessions.set(roomCode, gameState);
  
  // 记录玩家昵称到房间的映射
  playerToRoom.set(playerName, {
    roomCode: roomCode,
    playerId: 0,
    joinTime: new Date().toISOString()
  });
  
  res.json({
    roomCode: roomCode,
    playerId: 0,
    gameState: sanitizeGameState(gameState, 0)
  });
});

// 路由：通过房间号加入游戏
app.post('/api/game/join', (req: Request, res: Response) => {
  const { roomCode, playerName, asSpectator } = req.body;
  
  const gameState = gameSessions.get(roomCode);
  
  if (!gameState) {
    return res.status(404).json({ error: '房间不存在' });
  }
  
  // 检查名称是否已存在（玩家和观战者中）
  const nameExists = gameState.players.some(p => p.name === playerName) ||
                     (gameState.spectators && gameState.spectators.some(s => s.name === playerName));
  if (nameExists) {
    return res.status(400).json({ error: '该昵称已被使用' });
  }
  
  // 检查房间是否还有位置
  const hasSpaceForPlayer = gameState.players.length < gameState.maxPlayers;
  
  // 如果明确选择观战或房间已满，则作为观战者加入
  if (asSpectator || !hasSpaceForPlayer) {
    if (!gameState.spectators) {
      gameState.spectators = [];
    }
    
    const spectatorId = gameState.spectators.length;
    gameState.spectators.push({
      id: spectatorId,
      name: playerName,
      joinedAt: new Date().toISOString()
    });
    
    return res.json({
      roomCode: roomCode,
      playerId: null,
      spectatorId: spectatorId,
      isSpectator: true,
      gameState: sanitizeGameState(gameState, null)
    });
  }
  
  // 作为玩家加入
  const playerId = gameState.players.length;
  gameState.players.push({
    id: playerId,
    name: playerName,
    hand: [],
    compounds: [],
    isHost: false
  });
  
  // 记录玩家到房间的映射
  playerToRoom.set(playerName, {
    roomCode: roomCode,
    playerId: playerId,
    joinTime: new Date().toISOString()
  });
  
  res.json({
    roomCode: roomCode,
    playerId: playerId,
    isSpectator: false,
    gameState: sanitizeGameState(gameState, playerId)
  });
});

// 路由：生成房间二维码
app.get('/api/game/:roomCode/qrcode', async (req: Request, res: Response) => {
  const { roomCode } = req.params;
  const gameState = gameSessions.get(roomCode);
  
  if (!gameState) {
    return res.status(404).json({ error: '房间不存在' });
  }
  
  try {
    // 生成加入链接
    const joinUrl = `http://localhost:4000/join/${roomCode}`;
    
    // 生成二维码（Data URL格式）
    const qrcodeDataUrl = await QRCode.toDataURL(joinUrl, {
      width: 300,
      margin: 2,
      color: {
        dark: '#667eea',
        light: '#ffffff'
      }
    });
    
    res.json({
      qrcode: qrcodeDataUrl,
      joinUrl: joinUrl,
      roomCode: roomCode
    });
  } catch (error) {
    res.status(500).json({ error: '生成二维码失败' });
  }
});

// 路由：开始游戏
app.post('/api/game/:roomCode/start', (req: Request, res: Response) => {
  const { roomCode } = req.params;
  const { playerId } = req.body;
  
  const gameState = gameSessions.get(roomCode);
  
  if (!gameState) {
    return res.status(404).json({ error: '房间不存在' });
  }
  
  const player = gameState.players.find(p => p.id === playerId);
  if (!player || !player.isHost) {
    return res.status(403).json({ error: '只有房主可以开始游戏' });
  }
  
  if (gameState.players.length < 2) {
    return res.status(400).json({ error: '至少需要2名玩家才能开始游戏' });
  }
  
  // 根据玩家数量动态生成牌堆（每2人一组牌）
  const deckMultiplier = Math.ceil(gameState.players.length / 2);
  gameState.deck = gameLogic.initializeDeckForPlayers(gameState.players.length, deckMultiplier);
  
  // 给每个玩家发10张牌
  for (const player of gameState.players) {
    for (let i = 0; i < 10; i++) {
      if (gameState.deck.length > 0) {
        player.hand.push(gameState.deck.pop()!);
      }
    }
  }
  
  gameState.gameStarted = true;
  gameState.gameActive = true;
  gameState.currentPlayer = 0;
  gameState.lastCard = null;  // 初始化最后打出的卡牌
  
  res.json({
    success: true,
    gameState: sanitizeGameState(gameState, playerId)
  });
});

// 路由：获取房间信息（必须在通配路由之前）
app.get('/api/game/:roomCode/info', (req: Request, res: Response) => {
  const { roomCode } = req.params;
  const gameState = gameSessions.get(roomCode);
  
  if (!gameState) {
    return res.status(404).json({ error: '房间不存在' });
  }
  
  res.json({
    roomCode: roomCode,
    playerCount: gameState.players.length,
    maxPlayers: gameState.maxPlayers,
    gameStarted: gameState.gameStarted,
    players: gameState.players.map(p => ({
      id: p.id,
      name: p.name,
      isHost: p.isHost,
      cardCount: p.hand.length
    })),
    spectators: (gameState.spectators || []).map(s => ({
      id: s.id,
      name: s.name
    })),
    spectatorCount: (gameState.spectators || []).length
  });
});

// 路由：获取游戏状态（通配路由，必须在具体路由之后）
app.get('/api/game/:roomCode/:playerId', (req: Request, res: Response) => {
  const { roomCode, playerId } = req.params;
  const gameState = gameSessions.get(roomCode);
  
  if (!gameState) {
    return res.status(404).json({ error: '房间不存在' });
  }
  
  res.json({
    gameState: sanitizeGameState(gameState, parseInt(playerId))
  });
});

// 路由：获取可能的物质
app.post('/api/compounds', (req: Request, res: Response) => {
  const { elements } = req.body;
  
  try {
    const compounds = database.getCompoundsByElements(elements);
    res.json({ compounds });
  } catch (error: any) {
    res.status(500).json({ error: '获取化合物失败' });
  }
});

// 路由：获取/更新可编辑配置（卡牌、物质、反应规则）
app.get('/api/config', (req: Request, res: Response) => {
  res.json({ config: configService.getConfig() });
});

// 检查是否已完成初始化设置
app.get('/api/setup/check', (req: Request, res: Response) => {
  try {
    const rootDir = getRootDir();
    // 检查多个可能的.env文件位置
    const envPaths = [
      path.join(rootDir, '.env'),
      path.join(rootDir, 'client', '.env.production'),
      path.join(process.cwd(), '.env'),
      path.join(__dirname, '..', '.env') // Fallback for dev
    ];
    
    let adminPassword = '';
    
    for (const envPath of envPaths) {
      if (fs.existsSync(envPath)) {
        const envContent = fs.readFileSync(envPath, 'utf8');
        const match = envContent.match(/REACT_APP_ADMIN=(.+)/);
        if (match && match[1]) {
          adminPassword = match[1].trim();
          break;
        }
      }
    }
    
    // 如果文件中没有找到，使用环境变量
    if (!adminPassword) {
      adminPassword = process.env.REACT_APP_ADMIN || '';
    }
    
    const isSetup = adminPassword && adminPassword !== 'your_admin_password_here' && adminPassword.length > 0;
    res.json({ isSetup, message: isSetup ? '已设置' : '需要初始化' });
  } catch (error) {
    // 出错时假定未设置
    res.json({ isSetup: false, message: '检查失败，需要初始化' });
  }
});

// 验证管理密码 (新增) - 解决前端无法获取构建时环境变量的问题
app.post('/api/verify-password', (req: Request, res: Response) => {
  const { password } = req.body;
  const adminPassword = process.env.REACT_APP_ADMIN;
  
  if (!adminPassword) {
    return res.status(400).json({ success: false, message: '服务器未配置管理密码' });
  }
  
  if (password === adminPassword) {
    res.json({ success: true });
  } else {
    res.json({ success: false, message: '密码错误' });
  }
});

// 初始化设置 - 保存管理员密码
app.post('/api/setup', (req: Request, res: Response) => {
  const { adminPassword } = req.body;

  if (!adminPassword || adminPassword.length < 6) {
    return res.status(400).json({ error: '密码长度至少6位' });
  }

  try {
    const rootDir = getRootDir();

    console.log(`[Setup] 当前目录: ${__dirname}`);
    console.log(`[Setup] 项目根目录: ${rootDir}`);

    // 定义需要更新的.env文件路径
    const envFiles = [
      { path: path.join(rootDir, '.env'), name: '根目录.env' },
      // 同时更新client env，虽然在集成模式下主要靠后端验证，但保持一致性是好的
      { path: path.join(rootDir, 'client', '.env.production'), name: 'client/.env.production' }
    ];

    
    const defaultEnvContent = `# 化学UNO - 环境变量配置
NODE_ENV=production
PORT=4001
REACT_APP_ADMIN=${adminPassword}
ALLOWED_ORIGINS=http://localhost:4000,http://127.0.0.1:4000
`;
    
    const clientEnvContent = `REACT_APP_ADMIN=${adminPassword}
PORT=4000
BROWSER=none
SKIP_PREFLIGHT_CHECK=true
DISABLE_ESLINT_PLUGIN=true
`;
    
    // 更新所有.env文件
    for (const envFile of envFiles) {
      let envContent = '';
      
      if (fs.existsSync(envFile.path)) {
        envContent = fs.readFileSync(envFile.path, 'utf8');
      } else {
        // 根据文件类型使用不同的默认内容
        envContent = envFile.name.includes('client') ? clientEnvContent : defaultEnvContent;
      }
      
      // 更新或添加 REACT_APP_ADMIN
      const lines = envContent.split('\n');
      let found = false;
      
      for (let i = 0; i < lines.length; i++) {
        if (lines[i].startsWith('REACT_APP_ADMIN=')) {
          lines[i] = `REACT_APP_ADMIN=${adminPassword}`;
          found = true;
          break;
        }
      }
      
      if (!found) {
        lines.push(`REACT_APP_ADMIN=${adminPassword}`);
      }
      
      // 确保目录存在
      const dir = path.dirname(envFile.path);
      if (!fs.existsSync(dir)) {
        fs.mkdirSync(dir, { recursive: true });
      }
      
      // 写入文件
      fs.writeFileSync(envFile.path, lines.join('\n'), 'utf8');
    }
    
    // 更新当前进程的环境变量
    process.env.REACT_APP_ADMIN = adminPassword;
    
    res.json({ 
      success: true, 
      message: '设置已保存，请重启服务以生效' 
    });
  } catch (error: any) {
    console.error('[Setup] 保存失败:', error);
    res.status(500).json({ 
      error: '保存配置失败', 
      details: error.message,
      path: __dirname
    });
  }
});

// 刷新配置（从磁盘重新加载）
app.post('/api/config/refresh', (req: Request, res: Response) => {
  try {
    const refreshedConfig = configService.refreshFromDisk();
    res.json({ success: true, config: refreshedConfig });
  } catch (err: any) {
    res.status(500).json({ error: '刷新配置失败，请重试' });
  }
});

app.put('/api/config', (req: Request, res: Response) => {
  const incoming = req.body;

  if (!incoming || typeof incoming !== 'object') {
    return res.status(400).json({ error: '配置格式无效' });
  }

  try {
    const saved = configService.saveConfig(incoming);
    res.json({ success: true, config: saved });
  } catch (err: any) {
    res.status(400).json({ error: '配置格式错误或保存失败' });
  }
});

// 路由：获取配置中的元素列表
app.get('/api/elements', (req: Request, res: Response) => {
  try {
    const elements = configService.getElementsList();
    res.json({ elements });
  } catch (error: any) {
    res.status(500).json({ error: '获取元素列表失败' });
  }
});

// 路由：获取所有房间列表（大厅）
app.get('/api/rooms', (req: Request, res: Response) => {
  const rooms: any[] = [];
  
  for (const [roomCode, gameState] of gameSessions.entries()) {
    rooms.push({
      roomCode: roomCode,
      playerCount: gameState.players.length,
      maxPlayers: gameState.maxPlayers,
      gameStarted: gameState.gameStarted,
      hostName: gameState.players[0]?.name || '未知',
      spectatorCount: (gameState.spectators || []).length,
      createdAt: gameState.createdAt
    });
  }
  
  // 按创建时间排序，最新的在前
  rooms.sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime());
  
  res.json({ rooms });
});

// 路由：检查玩家是否有未完成的游戏
app.get('/api/player/:playerName/session', (req: Request, res: Response) => {
  const { playerName } = req.params;
  
  // 检查playerToRoom映射
  const sessionInfo = playerToRoom.get(playerName);
  
  if (!sessionInfo) {
    return res.json({ hasSession: false, session: null });
  }
  
  // 检查游戏是否还存在
  const gameState = gameSessions.get(sessionInfo.roomCode);
  
  if (!gameState) {
    // 游戏已结束，清理映射
    playerToRoom.delete(playerName);
    return res.json({ hasSession: false, session: null });
  }
  
  // 检查玩家是否仍在游戏中
  const player = gameState.players.find(p => p.id === sessionInfo.playerId);
  
  if (!player) {
    // 玩家已不在游戏中，清理映射
    playerToRoom.delete(playerName);
    return res.json({ hasSession: false, session: null });
  }
  
  // 返回可用的游戏会话
  return res.json({
    hasSession: true,
    session: {
      roomCode: sessionInfo.roomCode,
      playerId: sessionInfo.playerId,
      playerName: player.name,
      gameStarted: gameState.gameStarted,
      isOffline: player.isOffline || false
    }
  });
});

// 路由：检查物质是否能够反应
app.post('/api/reaction/check', (req: Request, res: Response) => {
  const { compound1, compound2 } = req.body;
  
  try {
    const canReact = database.getReactionBetweenCompounds(compound1, compound2);
    res.json({
      canReact: canReact
    });
  } catch (error: any) {
    res.status(500).json({ error: '检查反应失败' });
  }
});

// 路由：获取游戏统计
app.get('/api/game/:roomCode/stats', (req: Request, res: Response) => {
  const { roomCode } = req.params;
  const gameState = gameSessions.get(roomCode);
  
  if (!gameState) {
    return res.status(404).json({ error: '房间不存在' });
  }
  
  res.json({
    stats: GameRules.getGameStats(gameState)
  });
});

// WebSocket 连接处理
io.on('connection', (socket) => {
  
  
  socket.on('joinRoom', (data: any) => {
    const { roomCode, playerId, playerName } = data;
    const gameState = gameSessions.get(roomCode);
    
    if (!gameState) {
      socket.emit('error', '房间不存在');
      return;
    }
    
    socket.join(roomCode);
    playerSockets.set(playerName, socket.id);
    
    // 记录socket与玩家的关联
    const player = playerId !== null ? gameState.players.find((p: Player) => p.id === playerId) : null;
    const isSpectator = playerId === null;
    
    socketToPlayer.set(socket.id, {
      roomCode: roomCode,
      playerId: playerId,
      playerName: playerName,
      isHost: player ? player.isHost : false,
      isSpectator: isSpectator
    });
    
    // 如果是房主重新连接，取消房间关闭的超时
    if (player && player.isHost) {
      const cleanupKey = roomCode;
      const cleanup = pendingCleanup.get(cleanupKey);
      if (cleanup && cleanup.isHost) {
        
        clearTimeout(cleanup.timeoutId);
        pendingCleanup.delete(cleanupKey);
      }
    } else if (player) {
      // 如果是普通玩家重新连接，标记为在线并取消昵称释放的超时
      const cleanupKey = `${roomCode}:${playerId}`;
      const cleanup = pendingCleanup.get(cleanupKey);
      if (cleanup) {
        
        clearTimeout(cleanup.timeoutId);
        pendingCleanup.delete(cleanupKey);
      }
      
      // 无论是否有超时清理，都要确保标记为在线
      if (player.isOffline) {
        
        player.isOffline = false;
      }
    } else if (isSpectator) {
      // 观战者加入
      
    }
    
    // 向所有玩家广播玩家/观战者加入
    io.to(roomCode).emit('playerJoined', {
      playerId: playerId,
      playerName: playerName,
      isSpectator: isSpectator,
      playerCount: gameState.players.length,
      spectatorCount: gameState.spectators ? gameState.spectators.length : 0
    });
    
    // 向所有玩家广播当前游戏状态
    broadcastGameStateToAll(io, roomCode, gameState);
  });
  
  socket.on('startGame', (data: any) => {
    const { roomCode, playerId } = data;
    const gameState = gameSessions.get(roomCode);
    
    if (!gameState) {
      socket.emit('error', '房间不存在');
      return;
    }
    
    const player = gameState.players.find(p => p.id === playerId);
    if (!player || !player.isHost) {
      socket.emit('error', '只有房主可以开始游戏');
      return;
    }
    
    // 广播游戏开始 - 为每个玩家发送不同的sanitized gameState
    gameState.players.forEach((player) => {
      const sockets = io.sockets.adapter.rooms.get(roomCode);
      if (!sockets) return;
      
      for (const socketId of sockets) {
        const socketInfo = socketToPlayer.get(socketId);
        if (socketInfo && socketInfo.playerId === player.id) {
          // 发送给这个玩家他自己的视图
          io.to(socketId).emit('gameStarted', {
            gameState: sanitizeGameState(gameState, player.id)
          });
        }
      }
    });
    
    // 如果有观战者，发送不包含玩家手牌的视图
    const sockets = io.sockets.adapter.rooms.get(roomCode);
    if (sockets) {
      for (const socketId of sockets) {
        const socketInfo = socketToPlayer.get(socketId);
        // 如果这个socket不是任何玩家，说明是观战者
        if (socketInfo && !gameState.players.find(p => p.id === socketInfo.playerId)) {
          io.to(socketId).emit('gameStarted', {
            gameState: sanitizeGameState(gameState, null)
          });
        }
      }
    }
  });
  
  socket.on('playCard', (data: any) => {
    const { roomCode, playerId, card, compound } = data;
    const gameState = gameSessions.get(roomCode);
    
    
    
    if (!gameState || gameState.currentPlayer !== playerId) {
      
      socket.emit('error', '不是你的回合');
      return;
    }
    
    const player = gameState.players[playerId];
    
    // 检查卡牌是否在手中
    if (!player.hand.includes(card)) {
      
      socket.emit('error', '你没有这张卡牌');
      return;
    }
    
    // 特殊卡牌列表（无需检查反应）
    const specialCards = Object.keys(configService.getSpecialCards());
    let compoundElements: string[] = [];
    
    // 如果打出的是物质，进行合法性与反应性校验
    if (compound && !specialCards.includes(card)) {
      const elements = configService.getElementsList();
      const isElement = elements.includes(compound);

      // 1) 化合物必须存在于 common_compounds 中
      if (!isElement && !database.isKnownCompound(compound)) {
        socket.emit('error', '该物质不在可用列表中');
        return;
      }

      // 2) 获取组成元素：单质则为自身，化合物从数据库查询
      if (isElement) {
        compoundElements = [compound];
      } else {
        compoundElements = (database as any).compoundToElements?.[compound] || [];
      }
      
      if (compoundElements.length === 0 || !compoundElements.includes(card)) {
        socket.emit('error', '所选物质不包含该元素，无法打出');
        return;
      }

      // 3) 检查玩家是否持有组成所需的全部元素（至少各一张）
      const handCounts = player.hand.reduce((acc: Record<string, number>, el: string) => {
        acc[el] = (acc[el] || 0) + 1;
        return acc;
      }, {});

      const missing = compoundElements.find(el => (handCounts[el] || 0) <= 0);
      if (missing) {
        socket.emit('error', `你缺少所需元素: ${missing}`);
        return;
      }

      // 4) 若已有上一物质，则必须在反应列表中存在对应关系
      if (gameState.lastCompound) {
        const canReact = database.getReactionBetweenCompounds(gameState.lastCompound, compound);
        if (!canReact) {
          socket.emit('error', '该物质无法与上一物质反应');
          return;
        }
      }
    }
    
    
    
    // 移除卡牌/所需元素
    if (compound && !specialCards.includes(card)) {
      compoundElements.forEach(el => {
        const idx = player.hand.indexOf(el);
        if (idx !== -1) {
          player.hand.splice(idx, 1);
        }
      });
    } else {
      const index = player.hand.indexOf(card);
      player.hand.splice(index, 1);
    }
    
    // 记录物质和卡牌
    if (compound) {
      gameState.lastCompound = compound;
      player.compounds.push(compound);
    }
    gameState.lastCard = card;
    
    // 如果是特殊卡牌，应用效果并清空上一物质（特殊牌不参与反应链）
    if (GameRules.isSpecialCard(card)) {
      GameRules.applySpecialCard(card, gameState);
      gameState.lastCompound = null;
    }
    
    // 检查是否胜利
    if (GameRules.isWinner(player)) {
      // 计算游戏时长（秒）
      const gameTime = Math.floor((new Date().getTime() - new Date(gameState.createdAt).getTime()) / 1000);
      
      io.to(roomCode).emit('gameOver', {
        winner: playerId,
        playerName: player.name,
        finalScore: GameRules.calculateScore(player.hand),
        gameTime: gameTime
      });
      
      // 清理该房间所有玩家的会话映射
      gameState.players.forEach(p => {
        playerToRoom.delete(p.name);
      });
      
      gameSessions.delete(roomCode);
      return;
    }
    
    // 移到下一个玩家
    GameRules.nextPlayer(gameState);
    
    // 广播游戏状态更新
    broadcastGameStateToAll(io, roomCode, gameState);
  });
  
  socket.on('drawCard', (data: any) => {
    const { roomCode, playerId } = data;
    const gameState = gameSessions.get(roomCode);
    
    if (!gameState || gameState.currentPlayer !== playerId) {
      socket.emit('error', '不是你的回合或房间不存在');
      return;
    }
    
    const player = gameState.players[playerId];
    // 摸2张牌
    GameRules.drawCard(player, gameState, 2);
    
    if (!gameState.history) {
      gameState.history = [];
    }
    
    gameState.history.push({
      action: 'draw',
      player: playerId,
      cardsDrawn: 2
    });
    
    // 玩家无法出牌而摸牌，清除场上物质，下家可自由出牌
    if (gameState.lastCompound) {
      
      gameState.lastCompound = null;
    }
    
    // 移到下一个玩家
    GameRules.nextPlayer(gameState);
    
    // 广播游戏状态更新
    broadcastGameStateToAll(io, roomCode, gameState);
  });
  
  socket.on('disconnect', () => {
    
    
    // 获取断开连接的玩家信息
    const playerInfo = socketToPlayer.get(socket.id);
    if (!playerInfo) return;
    
    const { roomCode, playerId, playerName, isHost } = playerInfo;
    const gameState = gameSessions.get(roomCode);
    
    if (!gameState) {
      socketToPlayer.delete(socket.id);
      playerSockets.delete(playerName);
      return;
    }
    
    // 如果是房主离开，设置30秒后关闭房间
    if (isHost) {
      const settings = getGameSettings();
      
      
      // 设置超时后关闭房间
      const timeoutId = setTimeout(() => {
        
        
        const currentGameState = gameSessions.get(roomCode);
        if (!currentGameState) return;
        
        // 通知所有玩家房间已关闭
        io.to(roomCode).emit('roomClosed', {
          message: '房主长时间未返回，房间关闭',
          reason: 'hostTimeout'
        });
        
        // 清理所有玩家的socket映射
        currentGameState.players.forEach(p => {
          playerSockets.delete(p.name);
        });
        if (currentGameState.spectators) {
          currentGameState.spectators.forEach(s => {
            playerSockets.delete(s.name);
          });
        }
        
        // 删除房间
        gameSessions.delete(roomCode);
        pendingCleanup.delete(roomCode);
        
        
      }, settings.host_timeout);
      
      // 保存超时ID用于可能的取消
      pendingCleanup.set(roomCode, {
        isHost: true,
        playerName: playerName,
        timeoutId: timeoutId
      });
      
      socketToPlayer.delete(socket.id);
    } else {
      // 普通玩家或观战者离开，设置30秒后释放昵称
      
      
      // 标记玩家为离线状态，不立即删除
      const playerIndex = gameState.players.findIndex(p => p.id === playerId);
      if (playerIndex !== -1) {
        gameState.players[playerIndex].isOffline = true;
        
        // 如果游戏已开始且在线玩家数量少于2人，结束游戏
        const onlinePlayerCount = gameState.players.filter(p => !p.isOffline).length;
        if (gameState.gameStarted && onlinePlayerCount < 2) {
          gameState.gameActive = false;
          io.to(roomCode).emit('gameOver', {
            message: '在线玩家不足，游戏结束',
            reason: 'notEnoughPlayers'
          });
        }
      }
      
      socketToPlayer.delete(socket.id);
      
      const settings = getGameSettings();
      // 设置超时后释放昵称
      const timeoutId = setTimeout(() => {
        
        
        const currentGameState = gameSessions.get(roomCode);
        if (!currentGameState) {
          playerSockets.delete(playerName);
          pendingCleanup.delete(`${roomCode}:${playerId}`);
          return;
        }
        
        // 从玩家列表中移除
        const idx = currentGameState.players.findIndex(p => p.id === playerId);
        if (idx !== -1) {
          currentGameState.players.splice(idx, 1);
        } else {
          // 从观战者列表中移除
          if (currentGameState.spectators) {
            const spectatorIdx = currentGameState.spectators.findIndex(s => s.name === playerName);
            if (spectatorIdx !== -1) {
              currentGameState.spectators.splice(spectatorIdx, 1);
            }
          }
        }
        
        // 释放昵称和会话映射
        playerSockets.delete(playerName);
        playerToRoom.delete(playerName);
        pendingCleanup.delete(`${roomCode}:${playerId}`);
        
        // 通知其他玩家
        io.to(roomCode).emit('playerLeft', {
          playerId: playerId,
          playerName: playerName,
          playerCount: currentGameState.players.length,
          spectatorCount: currentGameState.spectators ? currentGameState.spectators.length : 0
        });
        
        // 广播更新游戏状态
        broadcastGameStateToAll(io, roomCode, currentGameState);
        
        
      }, settings.reconnect_timeout);
      
      // 保存超时ID用于可能的取消
      pendingCleanup.set(`${roomCode}:${playerId}`, {
        isHost: false,
        playerName: playerName,
        timeoutId: timeoutId
      });
      
      // 通知其他玩家
      io.to(roomCode).emit('playerLeft', {
        playerId: playerId,
        playerName: playerName,
        playerCount: gameState.players.filter(p => !p.isOffline).length,
        spectatorCount: gameState.spectators ? gameState.spectators.length : 0,
        isTemporary: true
      });
      
      // 广播更新游戏状态
      broadcastGameStateToAll(io, roomCode, gameState);
    }
  });
});

// 辅助函数：向房间内的所有玩家广播他们各自的sanitized gameState
function broadcastGameStateToAll(io: SocketIOServer, roomCode: string, gameState: GameState) {
  const sockets = io.sockets.adapter.rooms.get(roomCode);
  if (!sockets) return;
  
  for (const socketId of sockets) {
    const socketInfo = socketToPlayer.get(socketId);
    if (socketInfo) {
      const playerId = socketInfo.playerId;
      io.to(socketId).emit('gameStateUpdate', {
        gameState: sanitizeGameState(gameState, playerId)
      });
    }
  }
}

// 辅助函数：隐藏其他玩家的卡牌
function sanitizeGameState(gameState: GameState, playerId: number | null) {
  // 确保 playerId 是数字进行比较
  const playerIdNum = playerId !== null ? parseInt(String(playerId)) : null;
  
  const sanitized = {
    roomCode: gameState.roomCode,
    currentPlayer: gameState.currentPlayer,
    direction: gameState.direction,
    lastCompound: gameState.lastCompound,
    lastCard: gameState.lastCard,
    pendingDraws: gameState.pendingDraws || 0,
    deckCount: gameState.deck ? gameState.deck.length : 0,
    gameActive: gameState.gameActive,
    gameStarted: gameState.gameStarted,
    maxPlayers: gameState.maxPlayers,
    playerCount: gameState.players.length,
    players: gameState.players.map((player) => ({
      id: player.id,
      name: player.name,
      isHost: player.isHost,
      isOffline: player.isOffline || false,
      hand: player.id === playerIdNum ? player.hand : Array(player.hand.length).fill('unknown'),
      compounds: player.compounds,
      handCount: player.hand.length
    })),
    spectators: (gameState.spectators || []).map(s => ({
      id: s.id,
      name: s.name
    })),
    spectatorCount: (gameState.spectators || []).length
  };
  
  return sanitized;
}

// 静态文件托管 (集成部署模式)
// 必须放在 API 路由之后，404 处理之前
const rootDir = getRootDir();
const clientBuildPath = path.join(rootDir, 'client', 'build');

if (fs.existsSync(clientBuildPath)) {
  console.log(`[Server]启用静态文件托管: ${clientBuildPath}`);
  
  // 托管静态资源
  app.use(express.static(clientBuildPath));
  
  // 所有非 API 请求返回 index.html (SPA 支持)
  app.get('*', (req: Request, res: Response, next: NextFunction) => {
    // 如果是 API 请求但未匹配到上面的路由，交由 404 处理
    if (req.path.startsWith('/api/') || req.path.startsWith('/socket.io/')) {
      return next();
    }
    res.sendFile(path.join(clientBuildPath, 'index.html'));
  });
} else {
  console.log(`[Server] 未找到前端构建文件: ${clientBuildPath}`);
  console.log(`[Server] 以纯 API 模式运行`);
}

// 404 处理
app.use((req: Request, res: Response) => {
  res.status(404).json({
    error: 'Not Found',
    message: `路由 ${req.method} ${req.path} 不存在`,
    availableEndpoints: [
      'GET /',
      'POST /api/game/create',
      'POST /api/game/join',
      'GET /api/game/:roomCode/:playerId',
      'GET /api/game/:roomCode/info',
      'GET /api/game/:roomCode/qrcode',
      'POST /api/game/:roomCode/start',
      'POST /api/compounds',
      'POST /api/reaction/check',
      'GET /api/game/:roomCode/stats'
    ]
  });
});

// 错误处理
app.use((err: Error, req: Request, res: Response, next: NextFunction) => {
  res.status(500).json({
    error: 'Internal Server Error'
  });
});

const PORT = Number(process.env.PORT) || 4001;
const HOST = process.env.HOST || '0.0.0.0';

server.listen(PORT, HOST, () => {
  console.log(`✓ 服务器运行在 http://localhost:${PORT}`);
  console.log(`✓ 服务器监听在 ${HOST}:${PORT}`);
  console.log(`✓ WebSocket 服务已启动，等待连接...`);
  
  // 显示局域网访问地址
  const os = require('os');
  const interfaces = os.networkInterfaces();
  const addresses: string[] = [];
  
  for (const name of Object.keys(interfaces)) {
    for (const iface of interfaces[name]!) {
      // 跳过内部和非ipv4地址
      if (iface.family === 'IPv4' && !iface.internal) {
        addresses.push(iface.address);
      }
    }
  }
  
  if (addresses.length > 0) {
    console.log(`✓ 局域网访问地址:`);
    addresses.forEach(addr => {
      console.log(`  - http://${addr}:${PORT}`);
    });
  }
});
