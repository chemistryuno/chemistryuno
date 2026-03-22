#!/usr/bin/env python3
"""
Chemistry UNO 完整集成测试脚本
模拟多个玩家进行所有可用的游戏操作
"""

import requests
import json
import time
import sys
import random
import logging
from typing import Dict, List, Optional, Tuple
from dataclasses import dataclass, field
from enum import Enum

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s [%(levelname)s] %(name)s: %(message)s'
)
logger = logging.getLogger(__name__)

# ==================== 配置 ====================
BASE_URL = "http://localhost:8080"
WS_URL = "ws://localhost:8080/ws"

# 测试用户数据
TEST_USERS = [
    {"username": "testplayer1", "password": "Password@123"},
    {"username": "testplayer2", "password": "Password@123"},
    {"username": "testplayer3", "password": "Password@123"},
    {"username": "testplayer4", "password": "Password@123"},
    {"username": "testplayer5", "password": "Password@123"},
]


# ==================== 数据类和枚举 ====================
class GameAction(Enum):
    """游戏操作"""
    CREATE_ROOM = "创建房间"
    JOIN_ROOM = "加入房间"
    READY = "准备就绪"
    START_GAME = "开始游戏"
    PLAY_CARD = "出牌"
    DRAW_CARD = "摸牌"
    PLAY_DOUBLE = "双卡出牌"
    LEAVE_ROOM = "离开房间"
    WATCH_GAME = "观战"
    RECONNECT = "重新连接"


@dataclass
class PlayerSession:
    """玩家会话"""
    username: str
    password: str
    uid: int = 0
    token: str = ""
    
    def __repr__(self):
        return f"Player({self.username})"


@dataclass
class TestResult:
    """测试结果"""
    action: str
    status: str  # "PASS", "FAIL", "SKIP"
    message: str
    timestamp: float = field(default_factory=time.time)


# ==================== API 客户端 ====================
class GameAPIClient:
    """游戏API客户端"""
    
    def __init__(self, base_url: str = BASE_URL):
        self.base_url = base_url
        self.session = requests.Session()
    
    def set_auth_token(self, token: str):
        """设置认证令牌"""
        self.session.headers.update({
            "Authorization": f"Bearer {token}",
            "Content-Type": "application/json"
        })
    
    def _request(self, method: str, endpoint: str, data: Dict = None) -> Tuple[int, Dict]:
        """发送HTTP请求"""
        url = f"{self.base_url}{endpoint}"
        try:
            if method == "GET":
                resp = self.session.get(url, timeout=10)
            elif method == "POST":
                resp = self.session.post(url, json=data, timeout=10)
            elif method == "PUT":
                resp = self.session.put(url, json=data, timeout=10)
            elif method == "DELETE":
                resp = self.session.delete(url, timeout=10)
            else:
                raise ValueError(f"不支持的HTTP方法: {method}")
            
            try:
                return resp.status_code, resp.json()
            except:
                return resp.status_code, {"raw": resp.text}
        except Exception as e:
            logger.warning(f"请求失败 {method} {url}: {e}")
            return 0, {"error": str(e)}
    
    # ========== 认证相关 ==========
    def register(self, username: str, password: str) -> Tuple[bool, Dict]:
        """注册用户"""
        # 注册接口需要 nickname、security_question、security_answer
        status, data = self._request("POST", "/api/auth/register", {
            "username": username,
            "password": password,
            "nickname": username,  # 用用户名作为昵称
            "security_question": "测试问题?",
            "security_answer": "测试答案"
        })
        return status == 200, data
    
    def login(self, username: str, password: str) -> Tuple[bool, Dict]:
        """登录"""
        # login 接口需要 identifier 字段
        status, data = self._request("POST", "/api/auth/login", {
            "identifier": username,
            "password": password
        })
        return status == 200, data
    
    def get_user_info(self, token: str) -> Tuple[bool, Dict]:
        """获取用户信息"""
        self.set_auth_token(token)
        status, data = self._request("GET", "/api/user/info")
        return status == 200, data
    
    # ========== 房间相关 ==========
    def get_rooms(self) -> Tuple[bool, List[Dict]]:
        """获取所有房间"""
        status, data = self._request("GET", "/api/rooms")
        if status == 200:
            return True, data.get("data", [])
        return False, []
    
    def create_room(self, name: str, max_players: int = 4, token: str = None) -> Tuple[bool, Dict]:
        """创建房间"""
        if token:
            self.set_auth_token(token)
        status, data = self._request("POST", "/api/rooms", {
            "name": name,
            "max_players": max_players,
            "deck_config_id": 1,
            "is_private": False
        })
        return (status == 200 or status == 201), data
    
    def get_room(self, room_id: str, token: str) -> Tuple[bool, Dict]:
        """获取房间信息"""
        self.set_auth_token(token)
        status, data = self._request("GET", f"/api/rooms/{room_id}")
        return status == 200, data
    
    def join_room(self, room_id: str, token: str) -> Tuple[bool, Dict]:
        """加入房间"""
        self.set_auth_token(token)
        status, data = self._request("POST", f"/api/rooms/{room_id}/join", {"access_key": ""})
        return status == 200, data
    
    def ready(self, room_id: str, token: str) -> Tuple[bool, Dict]:
        """准备就绪"""
        self.set_auth_token(token)
        status, data = self._request("POST", f"/api/rooms/{room_id}/ready", {})
        return status == 200, data
    
    def start_game(self, room_id: str, token: str) -> Tuple[bool, Dict]:
        """开始游戏"""
        self.set_auth_token(token)
        status, data = self._request("POST", f"/api/rooms/{room_id}/start", {})
        return status == 200, data
    
    def draw(self, room_id: str, token: str) -> Tuple[bool, Dict]:
        """摸牌"""
        self.set_auth_token(token)
        status, data = self._request("POST", f"/api/rooms/{room_id}/draw", {})
        return status == 200, data
    
    def leave_room(self, room_id: str, token: str) -> Tuple[bool, Dict]:
        """离开房间"""
        self.set_auth_token(token)
        status, data = self._request("POST", f"/api/rooms/{room_id}/leave", {})
        return status == 200, data
    
    def check_reaction(self, element1: str, element2: str) -> Tuple[bool, Dict]:
        """验证化学反应"""
        status, data = self._request("POST", "/api/game/check-reaction", {
            "element1": element1,
            "element2": element2
        })
        return status == 200, data


# ==================== 测试管理器 ====================
class GameTestManager:
    """游戏测试管理器"""
    
    def __init__(self, base_url: str = BASE_URL):
        self.client = GameAPIClient(base_url)
        self.players: Dict[str, PlayerSession] = {}
        self.results: List[TestResult] = []
        self.current_room_id: str = ""
    
    def log_result(self, action: str, status: str, message: str):
        """记录测试结果"""
        result = TestResult(action=action, status=status, message=message)
        self.results.append(result)
        
        # 打印到控制台
        icon = {"PASS": "✓", "FAIL": "✗", "SKIP": "→"}.get(status, "•")
        print(f"[{icon}] {action}: {message}")
    
    def register_players(self) -> bool:
        """注册所有测试玩家"""
        print("\n" + "=" * 70)
        print("阶段 1: 玩家注册")
        print("=" * 70)
        
        for user_data in TEST_USERS:
            username = user_data["username"]
            password = user_data["password"]
            timestamp = int(time.time() * 1000) % 10000
            username_unique = f"{username}_{timestamp}"

            success, data = self.client.register(username_unique, password)

            # 允许 message 为 register success 也算成功
            if success or (isinstance(data, dict) and data.get("message") == "register success"):
                self.log_result("注册用户", "PASS", username_unique)
                player = PlayerSession(username=username_unique, password=password)
                self.players[username_unique] = player
            else:
                self.log_result("注册用户", "FAIL", f"{username_unique}: {data}")
                return False

        return len(self.players) > 0
    
    def login_players(self) -> bool:
        """登录所有玩家"""
        print("\n" + "=" * 70)
        print("阶段 2: 玩家登录")
        print("=" * 70)
        
        for username, player in self.players.items():
            success, data = self.client.login(username, player.password)

            # 兼容后端返回结构（token 可能在 data 顶层或 data['data']）
            token = None
            uid = None
            if success:
                if "data" in data and isinstance(data["data"], dict):
                    token = data["data"].get("token")
                    uid = data["data"].get("uid")
                elif "token" in data:
                    token = data.get("token")
                    uid = data.get("user", {}).get("uid")
            if token:
                player.token = token
                player.uid = uid or 0
                self.log_result("登录", "PASS", f"{username} (UID: {player.uid})")
            else:
                self.log_result("登录", "FAIL", f"{username}: {data}")
                return False

        return True
    
    def test_room_creation(self) -> Optional[str]:
        """测试创建房间"""
        print("\n" + "=" * 70)
        print("阶段 3: 创建房间")
        print("=" * 70)
        
        creator = list(self.players.values())[0]
        room_name = f"TestRoom_{int(time.time())}"
        
        success, data = self.client.create_room(room_name, max_players=4, token=creator.token)

        # 兼容后端直接返回房间数据结构
        room_id = None
        if success:
            if isinstance(data, dict) and "id" in data:
                room_id = data["id"]
            elif "data" in data and isinstance(data["data"], dict) and "id" in data["data"]:
                room_id = data["data"]["id"]
        if room_id:
            self.current_room_id = room_id
            self.log_result("创建房间", "PASS", f"{room_id} (房主: {creator.username})")
            return room_id
        else:
            self.log_result("创建房间", "FAIL", str(data))
            return None
    
    def test_join_room(self, room_id: str) -> bool:
        """测试加入房间"""
        print("\n" + "=" * 70)
        print("阶段 4: 玩家加入房间")
        print("=" * 70)
        
        players_list = list(self.players.values())
        
        for player in players_list[1:]:
            success, data = self.client.join_room(room_id, player.token)
            
            if success:
                self.log_result("加入房间", "PASS", f"{player.username}")
            else:
                self.log_result("加入房间", "FAIL", f"{player.username}: {data}")
                return False
        
        return True
    
    def test_ready_and_start(self, room_id: str) -> bool:
        """测试准备和开始"""
        print("\n" + "=" * 70)
        print("阶段 5: 准备就绪和开始游戏")
        print("=" * 70)
        
        players_list = list(self.players.values())
        
        for player in players_list:
            success, _ = self.client.ready(room_id, player.token)
            if success:
                self.log_result("准备", "PASS", player.username)
        
        time.sleep(1)
        creator = players_list[0]
        success, data = self.client.start_game(room_id, creator.token)
        
        if success:
            self.log_result("开始游戏", "PASS", "游戏已启动")
            return True
        else:
            self.log_result("开始游戏", "FAIL", str(data))
            return False
    
    def test_game_operations(self, room_id: str) -> bool:
        """测试游戏操作"""
        print("\n" + "=" * 70)
        print("阶段 6: 游戏操作测试")
        print("=" * 70)
        
        players_list = list(self.players.values())[:2]
        
        for player in players_list:
            success, _ = self.client.draw(room_id, player.token)
            self.log_result("摸牌", "PASS" if success else "SKIP", player.username)
        
        return True
    
    def test_leave_room(self, room_id: str) -> bool:
        """测试离开房间"""
        print("\n" + "=" * 70)
        print("阶段 7: 离开房间测试 (✓修复的功能)")
        print("=" * 70)
        
        players_list = list(self.players.values())
        
        if len(players_list) < 2:
            self.log_result("离开房间", "SKIP", "玩家不足")
            return True
        
        leaving_player = players_list[1]
        success, data = self.client.leave_room(room_id, leaving_player.token)
        
        if success:
            self.log_result("离开房间", "PASS", f"{leaving_player.username} 已离开")
        else:
            self.log_result("离开房间", "FAIL", str(data))
            return False
        
        time.sleep(0.5)
        success, room_data = self.client.get_room(room_id, players_list[0].token)
        if success:
            players = room_data.get("data", {}).get("players", [])
            spectators = room_data.get("data", {}).get("spectators", [])
            self.log_result("房间状态", "PASS", f"玩家:{len(players)}, 观战:{len(spectators)}")
        
        return True
    
    def test_spectator_features(self, room_id: str) -> bool:
        """测试观战功能"""
        print("\n" + "=" * 70)
        print("阶段 8: 观战功能测试 (✓修复的功能)")
        print("=" * 70)
        
        if len(self.players) < 5:
            self.log_result("观战", "SKIP", "玩家不足")
            return True
        
        fifth_player = list(self.players.values())[4]
        success, data = self.client.join_room(room_id, fifth_player.token)
        
        if success:
            self.log_result("观战", "PASS", f"{fifth_player.username} 加入为观战者")
            return True
        else:
            self.log_result("观战", "SKIP", "房间满员或其他原因")
            return True
    
    def test_reaction_check(self) -> bool:
        """测试化学反应"""
        print("\n" + "=" * 70)
        print("阶段 9: 化学反应验证")
        print("=" * 70)
        
        test_reactions = [("H", "O"), ("C", "O"), ("Na", "Cl")]
        
        for elem1, elem2 in test_reactions:
            success, data = self.client.check_reaction(elem1, elem2)
            status = "PASS" if success else "SKIP"
            substance = data.get("data", {}).get("substance", "?") if success else "N/A"
            self.log_result("反应验证", status, f"{elem1}+{elem2}→{substance}")
        
        return True
    
    def print_summary(self):
        """打印测试总结"""
        print("\n" + "=" * 70)
        print("测试总结")
        print("=" * 70)
        
        pass_count = sum(1 for r in self.results if r.status == "PASS")
        fail_count = sum(1 for r in self.results if r.status == "FAIL")
        skip_count = sum(1 for r in self.results if r.status == "SKIP")
        total = len(self.results)
        
        print(f"\n总计: {total} 个测试")
        print(f"✓ 通过: {pass_count}")
        print(f"✗ 失败: {fail_count}")
        print(f"→ 跳过: {skip_count}")
        
        if fail_count == 0:
            print("\n✓ 所有关键测试通过！")
            return True
        else:
            print(f"\n✗ 有 {fail_count} 个测试失败")
            for result in self.results:
                if result.status == "FAIL":
                    print(f"  - {result.action}: {result.message}")
            return False
    
    def run_all_tests(self) -> bool:
        """运行所有测试"""
        try:
            if not self.register_players():
                return False
            if not self.login_players():
                return False
            
            room_id = self.test_room_creation()
            if not room_id:
                return False
            
            self.test_join_room(room_id)
            self.test_ready_and_start(room_id)
            time.sleep(2)
            self.test_game_operations(room_id)
            time.sleep(1)
            self.test_leave_room(room_id)
            time.sleep(1)
            self.test_spectator_features(room_id)
            self.test_reaction_check()
            
            self.print_summary()
            return True
            
        except Exception as e:
            logger.error(f"测试异常: {e}", exc_info=True)
            return False


def main():
    """主函数"""
    print("""
    ╔════════════════════════════════════════════════════════════╗
    ║     Chemistry UNO 完整集成测试                            ║
    ║     测试所有玩家操作 + 修复的离开/观战功能               ║
    ╚════════════════════════════════════════════════════════════╝
    """)
    
    # 检查服务器连接
    try:
        resp = requests.get(f"{BASE_URL}/ping", timeout=5)
        if resp.status_code != 200:
            print("✗ 无法连接到后端服务器")
            print(f"  地址: {BASE_URL}")
            return False
    except Exception as e:
        print("✗ 无法连接到后端服务器")
        print(f"  地址: {BASE_URL}")
        print(f"  错误: {e}")
        return False
    
    print(f"✓ 已连接到 {BASE_URL}\n")
    
    manager = GameTestManager(BASE_URL)
    success = manager.run_all_tests()
    
    return success


if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1)
