import React, { useState, useEffect } from 'react';
import axios from 'axios';
import Card from './Card';
import CompoundSelector from './CompoundSelector';
import { formatFormula } from '../utils/chemistryFormatter';
import './GameBoard.css';
import API_ENDPOINTS from '../config/api';

interface Player {
  name?: string;
  handCount: number;
  hand: string[];
  isOffline?: boolean;
}

interface Spectator {
  id: string;
  name: string;
}

interface GameState {
  players: Player[];
  currentPlayer: string;
  deckCount: number;
  pendingDraws: number;
  lastCard: string | null;
  lastCompound: string | null;
  spectators?: Spectator[];
}

interface GameBoardProps {
  gameState: GameState;
  roomCode: string;
  playerId: string;
  socket: any;
  playerName: string;
  isSpectator: boolean;
}

const GameBoard: React.FC<GameBoardProps> = ({ gameState, roomCode, playerId, socket, playerName, isSpectator }) => {
  const [selectedCard, setSelectedCard] = useState<string | null>(null);
  const [compounds, setCompounds] = useState<string[]>([]);
  const [showCompoundSelector, setShowCompoundSelector] = useState<boolean>(false);
  const [gameStartTime] = useState<Date>(new Date());
  const [elapsedTime, setElapsedTime] = useState<number>(0);
  const [turnStartTime, setTurnStartTime] = useState<Date>(new Date());
  const [turnTimeRemaining, setTurnTimeRemaining] = useState<number>(30);
  const [lastCurrentPlayer, setLastCurrentPlayer] = useState<string>('');

  const isCurrentPlayer = !isSpectator && gameState && gameState.currentPlayer === playerId;

  // 更新全局计时器
  useEffect(() => {
    const timer = setInterval(() => {
      setElapsedTime(Math.floor((new Date().getTime() - gameStartTime.getTime()) / 1000));
    }, 1000);
    return () => clearInterval(timer);
  }, [gameStartTime]);

  // 当轮次变化时，重置轮次计时器
  useEffect(() => {
    if (gameState && gameState.currentPlayer !== lastCurrentPlayer) {
      setTurnStartTime(new Date());
      setTurnTimeRemaining(30);
      setLastCurrentPlayer(gameState.currentPlayer);
    }
  }, [gameState, gameState?.currentPlayer, lastCurrentPlayer]);

  // 更新轮次计时器
  useEffect(() => {
    const timer = setInterval(() => {
      const elapsed = Math.floor((new Date().getTime() - turnStartTime.getTime()) / 1000);
      const remaining = Math.max(0, 30 - elapsed);
      setTurnTimeRemaining(remaining);

      // 如果超过30秒且是当前玩家，自动摸牌
      if (remaining === 0 && isCurrentPlayer && socket) {
        socket.emit('drawCard', {
          roomCode,
          playerId
        });
      }
    }, 100);
    return () => clearInterval(timer);
  }, [turnStartTime, isCurrentPlayer, socket, roomCode, playerId]);

  // 当玩家点击卡牌时，获取可能的物质
  const handleCardClick = async (card: string): Promise<void> => {
    if (!isCurrentPlayer) return;


    // 检查是否是特殊卡牌（+2, +4, Au, He, Ne, Ar, Kr）
    const specialCards = ['+2', '+4', 'Au', 'He', 'Ne', 'Ar', 'Kr'];
    if (specialCards.includes(card)) {
      // 特殊卡牌直接打出，不需要选择物质
      if (socket) {
        socket.emit('playCard', {
          roomCode,
          playerId,
          card: card,
          compound: null, // 特殊卡牌不需要物质
          playerName
        });
      }
      return;
    }

    setSelectedCard(card);

    try {
      const response = await axios.post(API_ENDPOINTS.compounds, {
        elements: [card]
      });

      // 服务器返回的列表已包含对应单质和所有化合物
      const availableOptions = response.data.compounds;
      setCompounds(availableOptions);
      setShowCompoundSelector(true);
    } catch (error) {
      // 获取物质列表失败
      // 即使失败，也显示空列表并提示
      setCompounds([]);
      setShowCompoundSelector(true);
    }
  };

  // 玩家选择物质后
  const handleCompoundSelect = (compound: string): void => {
    
    if (socket) {
      socket.emit('playCard', {
        roomCode,
        playerId,
        card: selectedCard,
        compound,
        playerName
      });
    } else {
    }

    setShowCompoundSelector(false);
    setSelectedCard(null);
    setCompounds([]);
  };

  // 玩家无法打出时摸牌
  const handleDrawCard = (): void => {
    if (!isCurrentPlayer) return;

    if (socket) {
      socket.emit('drawCard', {
        roomCode,
        playerId
      });
    }
  };

  // 格式化时间
  const formatTime = (seconds: number): string => {
    const hrs = Math.floor(seconds / 3600);
    const mins = Math.floor((seconds % 3600) / 60);
    const secs = seconds % 60;
    return `${String(hrs).padStart(2, '0')}:${String(mins).padStart(2, '0')}:${String(secs).padStart(2, '0')}`;
  };

  if (!gameState) {
    return <div className="loading">加载中...</div>;
  }

  const currentPlayerIndex = gameState.players.findIndex((p, idx) => idx.toString() === playerId);
  const currentPlayer = !isSpectator && currentPlayerIndex !== -1 ? gameState.players[currentPlayerIndex] : null;
  
  const currentPlayerIdx = gameState.players.findIndex((p, idx) => idx.toString() === gameState.currentPlayer);
  const currentPlayerName = currentPlayerIdx !== -1 ? 
    (gameState.players[currentPlayerIdx]?.name || `玩家${currentPlayerIdx + 1}`) : 
    '未知玩家';

  return (
    <div className="game-board">
      {isSpectator && (
        <div className="spectator-banner">
          <span className="spectator-icon">👁️</span>
          观战模式 - {playerName}
        </div>
      )}
      
      {/* 游戏信息面板 */}
      <div className="game-info-panel">
        <div className="game-header">
          <h1>⚗️ 化学UNO</h1>
          <div className="game-stats">
            <div className="stat">
              <span className="stat-label">房间号</span>
              <span className="stat-value">{roomCode}</span>
            </div>
            <div className="stat">
              <span className="stat-label">用时</span>
              <span className="stat-value">{formatTime(elapsedTime)}</span>
            </div>
            <div className="stat">
              <span className="stat-label">剩余卡牌</span>
              <span className="stat-value">{gameState.deckCount}</span>
            </div>
            {gameState.pendingDraws > 0 && (
              <div className="stat pending-draws">
                <span className="stat-label">累加抽牌</span>
                <span className="stat-value warning">+{gameState.pendingDraws}</span>
              </div>
            )}
          </div>
        </div>

        {/* 当前玩家及计时器 */}
        <div className="current-player-section">
          <div className="current-player-info">
            <span className="label">当前玩家:</span>
            <span className={`player-name ${isCurrentPlayer ? 'is-me' : ''}`}>
              {currentPlayerName}
            </span>
            {isCurrentPlayer && <span className="your-turn-badge">轮到你了</span>}
          </div>
          <div className={`turn-timer ${turnTimeRemaining <= 10 ? 'warning' : ''} ${turnTimeRemaining <= 5 ? 'critical' : ''}`}>
            <span className="timer-label">剩余时间:</span>
            <span className="timer-value">{turnTimeRemaining}s</span>
          </div>
        </div>

        {/* 游戏中央区域 */}
        <div className="center-area">
          <div className="pile-area">
            <div className="pile-label">最后打出的牌</div>
            <div className="last-played">
              {gameState.lastCard ? (
                <div className="played-card-display">
                  <div className="played-card-label">卡牌: <strong>{gameState.lastCard}</strong></div>
                  {gameState.lastCompound && (
                    <div className="played-compound-label">物质: <strong>{formatFormula(gameState.lastCompound)}</strong></div>
                  )}
                </div>
              ) : (
                <div className="compound-card empty">游戏开始</div>
              )}
            </div>
          </div>

          {/* 其他玩家 */}
          <div className="other-players">
            {gameState.players.map((player, idx) => (
              idx.toString() !== playerId && (
                <div key={idx} className={`player-info ${idx.toString() === gameState.currentPlayer && !player.isOffline ? 'active' : ''} ${player.isOffline ? 'offline' : ''}`}>
                  <span className="player-label">
                    {player.isOffline ? '⚠️ ' : ''}{player.name || `玩家${idx + 1}`}
                    {player.isOffline && <span className="offline-text"> (离线中)</span>}
                  </span>
                  <span className="hand-count">{player.handCount}张</span>
                </div>
              )
            ))}
          </div>
        </div>

        <div className="turn-indicator">
          {isSpectator ? (
            <div className="spectating">观战中...</div>
          ) : isCurrentPlayer ? (
            <div className="your-turn">轮到你了！点击卡牌选择物质</div>
          ) : (
            <div className="waiting">等待中...</div>
          )}
        </div>
      </div>

      {/* 玩家手牌区 - 观战者不显示 */}
      {!isSpectator && currentPlayer && (
        <div className="player-hand-area">
          <div className="hand-label">我的卡牌（{currentPlayer.handCount}张）</div>
          <div className="hand-cards">
            {currentPlayer.hand.map((card, idx) => (
              <Card
                key={idx}
                card={card}
                onClick={() => handleCardClick(card)}
                isSelected={selectedCard === card}
                disabled={!isCurrentPlayer}
              />
            ))}
          </div>

          {isCurrentPlayer && (
            <button className="draw-btn" onClick={handleDrawCard}>
              摸牌 (无法打出)
            </button>
          )}
        </div>
      )}
      
      {/* 观战者视图显示所有玩家信息 */}
      {isSpectator && gameState.spectators && (
        <div className="spectator-info">
          <h3>观战者列表 ({gameState.spectators.length})</h3>
          <div className="spectators-mini-list">
            {gameState.spectators.map((spec) => (
              <span key={spec.id} className="spectator-tag">
                👁️ {spec.name}
              </span>
            ))}
          </div>
        </div>
      )}

      {/* 物质选择器浮窗 */}
      {showCompoundSelector && (
        <CompoundSelector
          compounds={compounds}
          selectedCard={selectedCard!}
          onSelect={handleCompoundSelect}
          onClose={() => setShowCompoundSelector(false)}
        />
      )}
    </div>
  );
};

export default GameBoard;
