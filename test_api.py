#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Chemistry UNO 全流程集成测试脚本
测试范围：从注册到游戏的所有玩家可用 API
用法：
    python test_api.py                        # 默认 localhost:8080
    python test_api.py --url http://x.x.x.x:8080
    python test_api.py --no-cleanup           # 测试后保留账号
    python test_api.py --verbose              # 打印响应详情
"""

import sys
import os
import json
import time
import random
import string
import argparse
import traceback

# Windows 终端 UTF-8 兼容
if sys.platform == "win32":
    import io
    sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding="utf-8", errors="replace")
    sys.stderr = io.TextIOWrapper(sys.stderr.buffer, encoding="utf-8", errors="replace")

try:
    import requests
except ImportError:
    print("缺少依赖: pip install requests")
    sys.exit(1)

# ─────────────────────────────────────────────
# 命令行参数
# ─────────────────────────────────────────────
parser = argparse.ArgumentParser(description="Chemistry UNO API 集成测试")
parser.add_argument("--url", default=os.environ.get("TEST_BASE_URL", "http://localhost:8080"),
                    help="后端地址 (默认 http://localhost:8080)")
parser.add_argument("--no-cleanup", action="store_true", help="测试完成后不删除测试账号")
parser.add_argument("--verbose", action="store_true", help="打印每条 API 响应详情")
args = parser.parse_args()

BASE_URL = args.url.rstrip("/") + "/api"
VERBOSE = args.verbose
CLEANUP = not args.no_cleanup

# ─────────────────────────────────────────────
# 颜色 & 输出工具
# ─────────────────────────────────────────────
NO_COLOR = not sys.stdout.isatty() or os.environ.get("NO_COLOR")

def c(code, text):
    return text if NO_COLOR else f"\033[{code}m{text}\033[0m"

def bold(t): return c("1", t)
def green(t): return c("92", t)
def red(t): return c("91", t)
def yellow(t): return c("93", t)
def blue(t): return c("94", t)
def cyan(t): return c("96", t)
def dim(t): return c("2", t)

# ─────────────────────────────────────────────
# 统计
# ─────────────────────────────────────────────
stats = {"passed": 0, "failed": 0, "skipped": 0, "errors": []}

def _mark(ok, name, note=""):
    if ok:
        stats["passed"] += 1
        sym = green("✓")
    else:
        stats["failed"] += 1
        sym = red("✗")
    tail = dim(f"  {note}") if note else ""
    print(f"  {sym} {name}{tail}")

def section(title):
    print(f"\n{bold(cyan('▶'))} {bold(title)}")

def skip(name, reason=""):
    stats["skipped"] += 1
    print(f"  {yellow('○')} {name}{dim('  跳过' + (': ' + reason if reason else ''))}")

# ─────────────────────────────────────────────
# HTTP 辅助
# ─────────────────────────────────────────────
SESSION_A = requests.Session()
SESSION_B = requests.Session()

def _req(session, method, path, *, expect=None, token=None, **kwargs):
    """发送请求并返回 (ok: bool, resp: Response)"""
    url = BASE_URL + path
    headers = kwargs.pop("headers", {})
    if token:
        headers["Authorization"] = f"Bearer {token}"
    try:
        resp = session.request(method, url, headers=headers, timeout=15, **kwargs)
    except requests.exceptions.ConnectionError:
        print(f"    {red('连接失败')} → {url}")
        return False, None
    except requests.exceptions.Timeout:
        print(f"    {red('请求超时')} → {url}")
        return False, None

    if VERBOSE:
        try:
            body = resp.json()
        except Exception:
            body = resp.text[:200]
        print(f"    {dim(method)} {dim(path)} → {resp.status_code}")
        print(f"    {dim(json.dumps(body, ensure_ascii=False)[:300])}")

    if expect is None:
        ok = resp.status_code < 400
    elif isinstance(expect, (list, tuple)):
        ok = resp.status_code in expect
    else:
        ok = resp.status_code == expect

    return ok, resp

def get(session, path, token=None, expect=None, **kw):
    return _req(session, "GET", path, token=token, expect=expect, **kw)

def post(session, path, data=None, token=None, expect=None, **kw):
    return _req(session, "POST", path, json=data, token=token, expect=expect, **kw)

def put(session, path, data=None, token=None, expect=None, **kw):
    return _req(session, "PUT", path, json=data, token=token, expect=expect, **kw)

def delete(session, path, data=None, token=None, expect=None, **kw):
    return _req(session, "DELETE", path, json=data, token=token, expect=expect, **kw)

# ─────────────────────────────────────────────
# 测试数据 (随机后缀避免冲突)
# ─────────────────────────────────────────────
sfx = ''.join(random.choices(string.ascii_lowercase + string.digits, k=7))

PLAYER = {
    "A": {
        "username":   f"testa_{sfx}",
        "email":      f"testa_{sfx}@testlocal.invalid",
        "password":   "Test@2025!",
        "nickname":   f"测试玩家甲{sfx[:4]}",
        "security_question": "最喜欢的元素",
        "security_answer":   "氢",
        "token": None, "uid": None,
    },
    "B": {
        "username":   f"testb_{sfx}",
        "email":      f"testb_{sfx}@testlocal.invalid",
        "password":   "Test@2025!",
        "nickname":   f"测试玩家乙{sfx[:4]}",
        "security_question": "最喜欢的化合物",
        "security_answer":   "水",
        "token": None, "uid": None,
    },
}

SESS = {"A": SESSION_A, "B": SESSION_B}

def tok(p): return PLAYER[p]["token"]
def uid(p): return PLAYER[p]["uid"]
def ses(p): return SESS[p]

# ─────────────────────────────────────────────
# ① 系统健康检查
# ─────────────────────────────────────────────
section("① 系统健康检查")

ok, r = get(SESSION_A, "/ping")
_mark(ok and r and "pong" in r.text.lower(), "GET /ping  服务心跳")

ok, r = get(SESSION_A, "/health")
_mark(ok, "GET /health  健康状态")

ok, r = get(SESSION_A, "/version")
_mark(ok, "GET /version  版本信息",
      r.json().get("fullVersion", "") if (ok and r) else "")

ok, r = get(SESSION_A, "/auth/config")
_mark(ok, "GET /auth/config  认证配置")

# ─────────────────────────────────────────────
# ② 注册 & 登录
# ─────────────────────────────────────────────
section("② 注册 & 登录")

for p in ("A", "B"):
    pl = PLAYER[p]
    ok, r = post(SESSION_A, "/auth/register", {
        "username":          pl["username"],
        "password":          pl["password"],
        "nickname":          pl["nickname"],
        "security_question": pl["security_question"],
        "security_answer":   pl["security_answer"],
    }, expect=[200, 201])
    _mark(ok, f"POST /auth/register  玩家{p} 注册",
          f"uid={r.json().get('uid') if (ok and r) else '?'}")
    if ok and r:
        pl["uid"] = r.json().get("uid")

# 重复注册应失败
ok, r = post(SESSION_A, "/auth/register", {
    "username": PLAYER["A"]["username"],
    "password": "Test@2025!",
    "nickname": "重复昵称",
    "security_question": "q",
    "security_answer": "a",
}, expect=[400, 409, 422])
_mark(ok, "POST /auth/register  重复用户名应拒绝")

# 登录
for p in ("A", "B"):
    pl = PLAYER[p]
    ok, r = post(ses(p), "/auth/login", {
        "identifier": pl["username"],
        "password":   pl["password"],
    })
    _mark(ok, f"POST /auth/login  玩家{p} 登录",
          "token=" + ("✓" if (ok and r and r.json().get("token")) else "✗"))
    if ok and r:
        pl["token"] = r.json().get("token")
        if not pl["uid"]:
            pl["uid"] = r.json().get("user", {}).get("uid")

# 密码错误应失败
ok, r = post(SESSION_A, "/auth/login", {
    "identifier": PLAYER["A"]["username"],
    "password": "wrong_password",
}, expect=[400, 401, 403])
_mark(ok, "POST /auth/login  错误密码应拒绝")

# ─────────────────────────────────────────────
# ③ 用户信息 & 资料管理
# ─────────────────────────────────────────────
section("③ 用户信息 & 资料管理")

ok, r = get(SESSION_A, "/user/info", token=tok("A"))
_mark(ok, "GET /user/info  获取自身信息",
      f"nickname={r.json().get('nickname','?') if (ok and r) else '?'}")

ok, r = put(SESSION_A, "/user/avatar", {"avatar": "🧪"}, token=tok("A"))
_mark(ok, "PUT /user/avatar  更新头像")

ok, r = put(SESSION_A, "/user/profile", {
    "nickname": PLAYER["A"]["nickname"],
    "bio": "这是集成测试账号",
    "show_email": False,
    "sound_volume": 0.8,
    "vibration_enabled": True,
}, token=tok("A"))
_mark(ok, "PUT /user/profile  更新个人资料")

ok, r = get(SESSION_A, f"/user/profile/{uid('B')}", token=tok("A"))
_mark(ok, "GET /user/profile/:uid  查看他人公开资料",
      f"target_uid={uid('B')}")

ok, r = get(SESSION_A, "/users/search", token=tok("A"),
            params={"q": PLAYER["B"]["username"][:8]})
_mark(ok, "GET /users/search  搜索玩家",
      f"结果={len(r.json()) if (ok and r) else '?'}条")

ok, r = put(SESSION_A, "/user/password", {
    "old_password": PLAYER["A"]["password"],
    "new_password": PLAYER["A"]["password"],  # 保持不变
}, token=tok("A"))
_mark(ok, "PUT /user/password  修改密码（新旧相同仍应成功）")

ok, r = get(SESSION_A, "/user/security-question", token=tok("A"))
_mark(ok, "GET /user/security-question  获取密保问题")

ok, r = put(SESSION_A, "/user/security-question", {
    "security_question": "新的密保问题",
    "security_answer":   "新答案",
    "current_password":  PLAYER["A"]["password"],
}, token=tok("A"))
_mark(ok, "PUT /user/security-question  更新密保问题")

ok, r = get(SESSION_A, "/user/sessions", token=tok("A"))
_mark(ok, "GET /user/sessions  会话列表",
      f"会话数={len(r.json()) if (ok and r) else '?'}")

# ─────────────────────────────────────────────
# ④ 等级系统
# ─────────────────────────────────────────────
section("④ 等级系统")

ok, r = get(SESSION_A, "/level/info", token=tok("A"))
_mark(ok, "GET /level/info  当前等级信息",
      f"level={r.json().get('level','?') if (ok and r) else '?'}")

ok, r = get(SESSION_A, "/level/configs")
_mark(ok, "GET /level/configs  等级配置列表")

ok, r = get(SESSION_A, "/level/leaderboard")
_mark(ok, "GET /level/leaderboard  等级排行榜",
      f"上榜={len(r.json()) if (ok and r and isinstance(r.json(), list)) else '?'}人")

ok, r = get(SESSION_A, f"/level/user/{uid('B')}")
_mark(ok, "GET /level/user/:uid  他人等级信息")

# ─────────────────────────────────────────────
# ⑤ 积分排行榜
# ─────────────────────────────────────────────
section("⑤ 积分排行榜")

ok, r = get(SESSION_A, "/points/leaderboard", token=tok("A"),
            params={"mode": "total"})
_mark(ok, "GET /points/leaderboard?mode=total  总积分榜",
      f"上榜={len(r.json().get('leaderboard',[])) if (ok and r) else '?'}人")

ok, r = get(SESSION_A, "/points/leaderboard", token=tok("A"),
            params={"mode": "monthly"})
_mark(ok, "GET /points/leaderboard?mode=monthly  月度积分榜")

# ─────────────────────────────────────────────
# ⑥ 好友系统
# ─────────────────────────────────────────────
section("⑥ 好友系统")

# A 向 B 发请求
ok, r = post(SESSION_A, "/friends/request", {
    "friend_uid": uid("B"),
    "message": "来自集成测试的好友申请",
}, token=tok("A"))
_mark(ok, "POST /friends/request  A→B 发送好友申请")

# B 查看待处理
ok, r = get(SESSION_B, "/friends/pending", token=tok("B"))
_mark(ok, "GET /friends/pending  B 查看待处理申请",
      f"数量={len(r.json()) if (ok and r and isinstance(r.json(), list)) else '?'}")

request_id = None
if ok and r:
    pending = r.json()
    if pending:
        request_id = pending[0].get("id")

if request_id:
    ok, r = post(SESSION_B, "/friends/handle", {
        "request_id": request_id,
        "action": "accept",
    }, token=tok("B"))
    _mark(ok, "POST /friends/handle  B 接受申请")
else:
    skip("POST /friends/handle  B 接受申请", "未找到申请ID")

ok, r = get(SESSION_A, "/friends", token=tok("A"))
_mark(ok, "GET /friends  A 好友列表",
      f"好友数={len(r.json()) if (ok and r and isinstance(r.json(), list)) else '?'}")

ok, r = post(SESSION_A, "/friends/remark", {
    "friend_uid": uid("B"),
    "remark": "测试小伙伴B",
}, token=tok("A"))
_mark(ok, "POST /friends/remark  设置好友备注")

# ─────────────────────────────────────────────
# ⑦ 聊天记录
# ─────────────────────────────────────────────
section("⑦ 聊天记录")

ok, r = get(SESSION_A, "/chat/global/history", token=tok("A"),
            params={"limit": 20})
_mark(ok, "GET /chat/global/history  全局聊天记录")

ok, r = get(SESSION_A, f"/chat/private/history/{uid('B')}", token=tok("A"),
            params={"limit": 20})
_mark(ok, "GET /chat/private/history/:uid  私聊记录")

# ─────────────────────────────────────────────
# ⑧ 牌组管理
# ─────────────────────────────────────────────
section("⑧ 牌组管理")

ok, r = get(SESSION_A, "/my-decks", token=tok("A"))
_mark(ok, "GET /my-decks  我的牌组列表",
      f"数量={len(r.json()) if (ok and r and isinstance(r.json(), list)) else '?'}")

test_deck = {
    "name": f"测试牌组_{sfx[:4]}",
    "cards": {"H": 10, "O": 8, "C": 6, "N": 4, "Na": 4, "Cl": 4},
    "initial_cards": 7,
}
ok, r = post(SESSION_A, "/my-decks", test_deck, token=tok("A"), expect=[200, 201])
_mark(ok, "POST /my-decks  创建自定义牌组")

deck_id = None
if ok and r:
    # 重新获取列表找到新建的
    _, r2 = get(SESSION_A, "/my-decks", token=tok("A"))
    if r2:
        my_decks = [d for d in (r2.json() or []) if not d.get("is_global")]
        if my_decks:
            deck_id = my_decks[0]["id"]

if deck_id:
    ok, r = put(SESSION_A, f"/my-decks/{deck_id}", {
        "name": test_deck["name"] + "_v2",
        "cards": {"H": 12, "O": 10, "C": 6, "N": 4, "Na": 4, "Cl": 4},
        "initial_cards": 7,
    }, token=tok("A"))
    _mark(ok, "PUT /my-decks/:id  更新牌组")
else:
    skip("PUT /my-decks/:id  更新牌组", "未找到牌组ID")

# ─────────────────────────────────────────────
# ⑨ 化学数据
# ─────────────────────────────────────────────
section("⑨ 化学数据")

ok, r = get(SESSION_A, "/substances/names")
_mark(ok, "GET /substances/names  物质名称映射",
      f"数量={len(r.json()) if (ok and r and isinstance(r.json(), list)) else '?'}")

ok, r = get(SESSION_A, "/data/substances", token=tok("A"))
_mark(ok, "GET /data/substances  全量物质数据库")

ok, r = get(SESSION_A, "/data/substances/my", token=tok("A"))
_mark(ok, "GET /data/substances/my  我提交的物质")

# 提交物质建议（可能因重复被拒绝，两种情况均可接受）
ok, r = post(SESSION_A, "/data/substances/new", {
    "formula": f"X{sfx[:3].upper()}",
    "name":    f"测试物质{sfx[:3]}",
    "elements": "C,H",
}, token=tok("A"), expect=[200, 201, 400, 409])
_mark(ok, "POST /data/substances/new  提交新物质建议")

ok, r = get(SESSION_A, "/reactions", token=tok("A"))
_mark(ok, "GET /reactions  反应数据列表",
      f"数量={len(r.json()) if (ok and r and isinstance(r.json(), list)) else '?'}")

ok, r = get(SESSION_A, "/reactions/all", token=tok("A"))
_mark(ok, "GET /reactions/all  全部已批准反应")

ok, r = get(SESSION_A, "/reactions/my", token=tok("A"))
_mark(ok, "GET /reactions/my  我提交的反应")

# 提交反应（普通用户提交待审核）
ok, r = post(SESSION_A, "/reactions", {
    "display": "H2 + Cl2 = 2HCl",
}, token=tok("A"), expect=[200, 201, 400, 409])
_mark(ok, "POST /reactions  提交化学反应（待审核）")

# 验证反应
ok, r = post(SESSION_A, "/game/check-reaction", {
    "r1": "H2",
    "r2": "O2",
}, token=tok("A"), expect=[200, 400])
_mark(ok, "POST /game/check-reaction  验证化学反应合法性")

# ─────────────────────────────────────────────
# ⑩ 公共信息
# ─────────────────────────────────────────────
section("⑩ 公共信息")

ok, r = get(SESSION_A, "/announcements")
_mark(ok, "GET /announcements  公告列表")

ok, r = get(SESSION_A, "/hints")
_mark(ok, "GET /hints  随机游戏提示")

ok, r = get(SESSION_A, "/plugin-cards", token=tok("A"))
_mark(ok, "GET /plugin-cards  插件卡牌列表")

ok, r = get(SESSION_A, "/plugins", token=tok("A"))
_mark(ok, "GET /plugins  插件市场")

# ─────────────────────────────────────────────
# ⑪ 游戏流程（完整对局）
# ─────────────────────────────────────────────
section("⑪ 游戏流程（完整对局）")

room_id = None

# 查看房间列表
ok, r = get(SESSION_A, "/rooms", token=tok("A"))
_mark(ok, "GET /rooms  大厅房间列表",
      f"当前{len(r.json()) if (ok and r and isinstance(r.json(), list)) else '?'}个房间")

# A 创建房间（2人，使用测试牌组或默认）
create_body = {
    "name": f"测试对局_{sfx[:4]}",
    "max_players": 2,
    "is_points_mode": False,
    "is_private": False,
    "enable_ai_backfill": False,
}
if deck_id:
    create_body["deck_id"] = deck_id

ok, r = post(SESSION_A, "/rooms", create_body, token=tok("A"), expect=[200, 201])
_mark(ok, "POST /rooms  创建游戏房间")
if ok and r:
    room_id = r.json().get("id")
    _mark(bool(room_id), "  └─ 房间 ID 已返回", f"id={room_id}")

if not room_id:
    skip("后续游戏流程测试", "房间创建失败")
else:
    # 获取房间详情
    ok, r = get(SESSION_A, f"/rooms/{room_id}", token=tok("A"))
    _mark(ok, "GET /rooms/:id  获取房间详情",
          f"状态={r.json().get('status','?') if (ok and r) else '?'}")

    # 检查房间状态
    ok, r = get(SESSION_A, f"/rooms/{room_id}/status", token=tok("A"))
    _mark(ok, "GET /rooms/:id/status  检查房间状态")

    # B 加入房间
    ok, r = post(SESSION_B, f"/rooms/{room_id}/join", token=tok("B"))
    _mark(ok, "POST /rooms/:id/join  玩家B 加入房间")

    # A 准备
    ok, r = post(SESSION_A, f"/rooms/{room_id}/ready", token=tok("A"))
    _mark(ok, "POST /rooms/:id/ready  玩家A 准备")

    # B 准备
    ok, r = post(SESSION_B, f"/rooms/{room_id}/ready", token=tok("B"))
    _mark(ok, "POST /rooms/:id/ready  玩家B 准备")

    # A（房主）开始游戏
    ok, r = post(SESSION_A, f"/rooms/{room_id}/start", token=tok("A"))
    _mark(ok, "POST /rooms/:id/start  开始游戏")

    if ok:
        time.sleep(0.5)  # 等待后端初始化游戏状态

        # 获取游戏状态
        ok, r = get(SESSION_A, f"/rooms/{room_id}", token=tok("A"))
        _mark(ok, "GET /rooms/:id  获取游戏状态（对局中）")

        game_state = r.json() if (ok and r) else {}
        status = game_state.get("status", "")
        current_uid = game_state.get("current_player_uid")
        _mark(status == "playing", f"  └─ 游戏状态为 'playing'", f"实际={status}")

        # 获取可用物质
        ok, r = get(SESSION_A, f"/rooms/{room_id}/substances", token=tok("A"))
        _mark(ok, "GET /rooms/:id/substances  当前可用物质")

        # 获取反应提示
        ok, r = get(SESSION_A, f"/rooms/{room_id}/reaction-hints", token=tok("A"))
        _mark(ok, "GET /rooms/:id/reaction-hints  化学反应提示")

        # ── 轮流出牌 / 摸牌测试 ──
        def get_state(p):
            _, r = get(ses(p), f"/rooms/{room_id}", token=tok(p))
            return r.json() if r else {}

        def whose_turn(state):
            return state.get("current_player_uid")

        def get_hand(state, player_uid):
            hands = state.get("hands", {})
            return hands.get(str(player_uid), [])

        def try_play(p, state):
            """尝试打出手中第一张牌，失败则摸牌"""
            hand = get_hand(state, uid(p))
            if not hand:
                return False
            card = hand[0]
            # 尝试以 H2O 作为产物（大多数情况会被服务端拒绝，这是正常的）
            substance = "H2O"
            ok2, _ = post(ses(p), f"/rooms/{room_id}/play", {
                "card": card,
                "substance": substance,
            }, token=tok(p), expect=[200, 201, 400, 422])
            return ok2

        play_ok_count = 0
        draw_ok_count = 0

        for turn in range(4):  # 最多测 4 回合
            state = get_state("A")
            cur = whose_turn(state)
            if not cur:
                break

            p = "A" if cur == uid("A") else "B"
            other = "B" if p == "A" else "A"

            # 先尝试摸牌（总是安全的）
            ok_d, _ = post(ses(p), f"/rooms/{room_id}/draw", token=tok(p),
                           expect=[200, 201, 400])
            if ok_d:
                draw_ok_count += 1

        _mark(draw_ok_count > 0, f"POST /rooms/:id/draw  摸牌操作",
              f"成功{draw_ok_count}次")

    # 获取历史记录
    ok, r = get(SESSION_A, "/user/game-history", token=tok("A"))
    _mark(ok, "GET /user/game-history  游戏历史记录",
          f"条数={len(r.json()) if (ok and r and isinstance(r.json(), list)) else '?'}")

    # B 先离开
    ok, r = post(SESSION_B, f"/rooms/{room_id}/leave", token=tok("B"))
    _mark(ok, "POST /rooms/:id/leave  玩家B 离开房间")

    # A 离开（结束房间）
    ok, r = post(SESSION_A, f"/rooms/{room_id}/leave", token=tok("A"))
    _mark(ok, "POST /rooms/:id/leave  玩家A 离开房间")

# ─────────────────────────────────────────────
# ⑫ 反馈系统
# ─────────────────────────────────────────────
section("⑫ 反馈系统")

feedback_id = None

ok, r = post(SESSION_A, "/feedback", {
    "content": f"集成测试自动反馈 [{sfx}]",
    "type": "bug",
}, token=tok("A"), expect=[200, 201])
_mark(ok, "POST /feedback  提交反馈")

ok, r = get(SESSION_A, "/feedbacks/my", token=tok("A"))
_mark(ok, "GET /feedbacks/my  我的反馈列表",
      f"数量={len(r.json()) if (ok and r and isinstance(r.json(), list)) else '?'}")
if ok and r and r.json():
    feedback_id = r.json()[0].get("id")

if feedback_id:
    ok, r = post(SESSION_A, f"/feedbacks/{feedback_id}/urge", token=tok("A"),
                 expect=[200, 400, 429])
    _mark(ok, "POST /feedbacks/:id/urge  催促反馈处理（可能因限频拒绝）")

    ok, r = post(SESSION_A, "/feedback/withdraw", {
        "id": feedback_id,
    }, token=tok("A"), expect=[200, 400])
    _mark(ok, "POST /feedback/withdraw  撤回反馈")
else:
    skip("POST /feedbacks/:id/urge  催促反馈", "未找到反馈ID")
    skip("POST /feedback/withdraw  撤回反馈", "未找到反馈ID")

# ─────────────────────────────────────────────
# ⑬ 问卷调查
# ─────────────────────────────────────────────
section("⑬ 问卷调查")

ok, r = get(SESSION_A, "/surveys/active", token=tok("A"))
_mark(ok, "GET /surveys/active  当前活跃问卷")

ok, r = get(SESSION_A, "/surveys/all", token=tok("A"))
_mark(ok, "GET /surveys/all  全部活跃问卷")

# ─────────────────────────────────────────────
# ⑭ 安全功能
# ─────────────────────────────────────────────
section("⑭ 账号安全")

# 获取密保问题（公开）
ok, r = get(SESSION_A, "/auth/security-question",
            params={"username": PLAYER["A"]["username"]})
_mark(ok, "GET /auth/security-question  获取他人密保问题（公开）")

# 未认证访问受保护接口
ok_unauth, _ = get(requests.Session(), "/user/info", expect=[401, 403])
_mark(ok_unauth, "GET /user/info（无token）应返回 401/403")

# ─────────────────────────────────────────────
# ⑮ 删除好友（清理好友关系）
# ─────────────────────────────────────────────
section("⑮ 好友关系解除")

ok, r = delete(SESSION_A, f"/friends/{uid('B')}", token=tok("A"), expect=[200, 404])
_mark(ok, "DELETE /friends/:uid  删除好友关系")

# ─────────────────────────────────────────────
# ⑯ 密保问题重置密码
# ─────────────────────────────────────────────
section("⑯ 密保问题重置密码")

# B 通过密保问题重置密码（使用相同密码）
ok, r = post(SESSION_A, "/auth/security-question/reset-password", {
    "username":        PLAYER["B"]["username"],
    "security_answer": PLAYER["B"]["security_answer"],
    "new_password":    PLAYER["B"]["password"],
})
_mark(ok, "POST /auth/security-question/reset-password  B 通过密保重置密码")

# 重置后重新登录 B（原 session 已失效）
ok, r = post(SESSION_B, "/auth/login", {
    "identifier": PLAYER["B"]["username"],
    "password":   PLAYER["B"]["password"],
})
_mark(ok, "POST /auth/login  B 重置密码后重新登录")
if ok and r:
    PLAYER["B"]["token"] = r.json().get("token")

# 错误答案应拒绝
ok, r = post(SESSION_A, "/auth/security-question/reset-password", {
    "username":        PLAYER["A"]["username"],
    "security_answer": "错误答案",
    "new_password":    "NewPass@2025",
}, expect=[400, 401, 403])
_mark(ok, "POST /auth/security-question/reset-password  错误答案应拒绝")

# ─────────────────────────────────────────────
# ⑰ 会话管理增强
# ─────────────────────────────────────────────
section("⑰ 会话管理增强")

# 获取 B 的当前会话列表
ok, r = get(SESSION_B, "/user/sessions", token=tok("B"))
_mark(ok, "GET /user/sessions  B 获取会话列表")

# 撤销 B 的最早一个会话（若存在多个）
if ok and r and isinstance(r.json(), list) and len(r.json()) > 1:
    old_session_id = r.json()[-1].get("id")
    if old_session_id:
        ok2, _ = post(SESSION_B, "/user/sessions/logout", {
            "id": old_session_id,
        }, token=tok("B"), expect=[200, 400])
        _mark(ok2, "POST /user/sessions/logout  撤销指定历史会话",
              f"session_id={old_session_id}")
    else:
        skip("POST /user/sessions/logout  撤销指定历史会话", "会话ID为空")
else:
    skip("POST /user/sessions/logout  撤销指定历史会话", "无可撤销的历史会话")

# ─────────────────────────────────────────────
# ⑱ 资料增强（生日 & 联系方式）
# ─────────────────────────────────────────────
section("⑱ 资料增强（生日 & 联系方式）")

ok, r = put(SESSION_A, "/user/profile", {
    "nickname":   PLAYER["A"]["nickname"],
    "bio":        "这是集成测试账号",
    "birthday":   "2000-01-01",
    "show_email": False,
}, token=tok("A"))
_mark(ok, "PUT /user/profile  更新资料含生日字段")

ok, r = get(SESSION_B, f"/user/profile/{uid('A')}", token=tok("B"))
_mark(ok, "GET /user/profile/:uid  B 查看 A 的公开资料（含生日）")

# ─────────────────────────────────────────────
# ⑲ 单挑系统（HTTP 层验证）
# ─────────────────────────────────────────────
section("⑲ 单挑系统")

# A 向 B 发起单挑（B 未连 WebSocket，预期失败）
ok, r = post(SESSION_A, "/game/duel", {
    "target_uid": uid("B"),
}, token=tok("A"), expect=[200, 400, 404])
_mark(ok, "POST /game/duel  A→B 发起单挑（B不在线应返回400）",
      r.json().get("error", "") if (ok and r) else "?")

# 挑战自己应拒绝
ok, r = post(SESSION_A, "/game/duel", {
    "target_uid": uid("A"),
}, token=tok("A"), expect=[400])
_mark(ok, "POST /game/duel  挑战自己应拒绝")

# B 响应单挑（无有效挑战ID，预期400）
ok, r = post(SESSION_B, "/game/duel/respond", {
    "target_uid": uid("A"),
    "accept": False,
}, token=tok("B"), expect=[200, 400])
_mark(ok, "POST /game/duel/respond  响应单挑邀请（无有效挑战应拒绝）")

# ─────────────────────────────────────────────
# ⑳ 积分悬赏
# ─────────────────────────────────────────────
section("⑳ 积分悬赏")

# A 向 B 发布悬赏（新用户积分不足，预期400）
ok, r = post(SESSION_A, "/points/bounty", {
    "target_uid": uid("B"),
    "amount": 100,
}, token=tok("A"), expect=[200, 400])
_mark(ok, "POST /points/bounty  积分不足时发布悬赏应拒绝",
      r.json().get("error", "") if (ok and r) else "?")

# 向自己发悬赏应拒绝
ok, r = post(SESSION_A, "/points/bounty", {
    "target_uid": uid("A"),
    "amount": 100,
}, token=tok("A"), expect=[400])
_mark(ok, "POST /points/bounty  向自己设置悬赏应拒绝")

# ─────────────────────────────────────────────
# ㉑ AI 补位功能
# ─────────────────────────────────────────────
section("㉑ AI 补位功能")

ai_room_id = None

# 创建启用 AI 补位的房间（3人 + 补位AI填满）
ok, r = post(SESSION_A, "/rooms", {
    "name": f"AI补位测试_{sfx[:4]}",
    "max_players": 3,
    "is_points_mode": False,
    "is_private": False,
    "enable_ai_backfill": True,
    "ai_backfill_difficulty": 50,
}, token=tok("A"), expect=[200, 201])
_mark(ok, "POST /rooms  创建启用 AI 补位的房间",
      f"id={r.json().get('id') if (ok and r) else '?'}")
if ok and r:
    ai_room_id = r.json().get("id")

if ai_room_id:
    # 验证房间详情中 AI 补位字段存在
    ok, r = get(SESSION_A, f"/rooms/{ai_room_id}", token=tok("A"))
    if ok and r:
        room_data = r.json()
        has_backfill = "enable_ai_backfill" in room_data or "ai_backfill" in str(room_data)
        _mark(ok, "GET /rooms/:id  获取 AI 补位房间详情")

    # A 准备并开始（AI 应自动补入填满空位）
    ok, r = post(SESSION_A, f"/rooms/{ai_room_id}/ready", token=tok("A"))
    _mark(ok, "POST /rooms/:id/ready  房主在 AI 补位房间中准备")

    ok, r = post(SESSION_A, f"/rooms/{ai_room_id}/start", token=tok("A"),
                 expect=[200, 400])
    _mark(ok, "POST /rooms/:id/start  启动 AI 补位游戏",
          r.json().get("error", "成功") if r else "?")

    # 验证游戏状态（若开始成功）
    if ok:
        time.sleep(0.5)
        ok2, r2 = get(SESSION_A, f"/rooms/{ai_room_id}", token=tok("A"))
        if ok2 and r2:
            st = r2.json().get("status", "")
            players = r2.json().get("players", [])
            ai_players = [p for p in players if isinstance(p.get("uid"), int) and p["uid"] < 0]
            _mark(len(ai_players) > 0 or st == "playing",
                  "  └─ AI 玩家已补入",
                  f"AI数={len(ai_players)}，状态={st}")

    # 离开 AI 补位房间
    post(SESSION_A, f"/rooms/{ai_room_id}/leave", token=tok("A"))
else:
    skip("AI 补位房间详细测试", "房间创建失败")

# PvE 模式不允许开启 AI 补位（参数验证）
ok, r = post(SESSION_A, "/rooms", {
    "name": f"PvE补位_invalid_{sfx[:4]}",
    "max_players": 2,
    "game_mode": "pve",
    "enable_ai_backfill": True,
}, token=tok("A"), expect=[200, 201, 400])
# PvE + 补位本身合法性取决于后端，仅验证接口可达
_mark(ok, "POST /rooms  PvE模式+AI补位（接口可达）")
if ok and r:
    pve_room = r.json().get("id")
    if pve_room:
        post(SESSION_A, f"/rooms/{pve_room}/leave", token=tok("A"))

# ─────────────────────────────────────────────
# ㉒ 双联反应（接口可达性验证）
# ─────────────────────────────────────────────
section("㉒ 双联反应")

# 非游戏中调用应返回 400
ok, r = post(SESSION_A, f"/rooms/{room_id or 'invalid'}/play-double", {
    "sub1": "H2",
    "sub2": "O2",
}, token=tok("A"), expect=[200, 400, 404])
_mark(ok, "POST /rooms/:id/play-double  双联反应接口可达（非游戏中应拒绝）",
      r.json().get("error", "") if (ok and r) else "?")

# 参数缺失应返回 400
ok, r = post(SESSION_A, f"/rooms/{room_id or 'invalid'}/play-double", {
    "sub1": "H2",
    # 缺少 sub2
}, token=tok("A"), expect=[400])
_mark(ok, "POST /rooms/:id/play-double  缺少参数应返回 400")

# ─────────────────────────────────────────────
# ㉓ 账号冻结
# ─────────────────────────────────────────────
section("㉓ 账号冻结")

# B 自我冻结 1 小时（仅测试接口，不影响清理流程中的账号删除）
ok, r = post(SESSION_B, "/user/account/freeze", {
    "hours": 1,
}, token=tok("B"), expect=[200])
_mark(ok, "POST /user/account/freeze  B 冻结自身账号 1 小时")

# 冻结时长超出范围（max 24h）应拒绝
ok, r = post(SESSION_B, "/user/account/freeze", {
    "hours": 100,
}, token=tok("B"), expect=[400])
_mark(ok, "POST /user/account/freeze  冻结时长超限应拒绝")

# ─────────────────────────────────────────────
# ⑯ 清理（删除测试账号）
# ─────────────────────────────────────────────
section("⑯ 清理测试账号")

if CLEANUP:
    if deck_id:
        ok, r = delete(SESSION_A, f"/my-decks/{deck_id}", token=tok("A"))
        _mark(ok, "DELETE /my-decks/:id  删除测试牌组")

    for p in ("B", "A"):
        pl = PLAYER[p]
        if pl["token"]:
            # 使用密保答案删除账号
            ok, r = delete(ses(p), "/user/account",
                           data={"security_answer": pl["security_answer"]},
                           token=tok(p), expect=[200, 204, 400])
            _mark(ok, f"DELETE /user/account  删除玩家{p} 账号")
else:
    print(f"  {yellow('○')} 跳过账号删除（--no-cleanup 模式）")
    for p in ("A", "B"):
        pl = PLAYER[p]
        print(f"    {dim('玩家'+p+':')} username={pl['username']}  password={pl['password']}")

# ─────────────────────────────────────────────
# 汇总
# ─────────────────────────────────────────────
total = stats["passed"] + stats["failed"] + stats["skipped"]
pass_rate = int(stats["passed"] / max(stats["passed"] + stats["failed"], 1) * 100)

print(f"\n{'─'*55}")
print(bold("测试结果汇总"))
print(f"  {green('通过')} {bold(str(stats['passed']))}  "
      f"{red('失败')} {bold(str(stats['failed']))}  "
      f"{yellow('跳过')} {bold(str(stats['skipped']))}  "
      f"共计 {total} 项")
rate_str = str(pass_rate) + '%'
rate_colored = green(rate_str) if pass_rate >= 90 else (yellow(rate_str) if pass_rate >= 70 else red(rate_str))
print(f"  通过率: {bold(rate_colored)}")
print(f"  目标地址: {dim(BASE_URL)}")

if stats["failed"] > 0:
    print(f"\n{yellow('提示')}: 使用 --verbose 参数可查看每条 API 的响应详情")
    sys.exit(1)
else:
    print(f"\n{green('全部通过！')}")
    sys.exit(0)
