package plugins

import (
	"chemistryuno/backend/database"
	"chemistryuno/backend/repository"
	"chemistryuno/backend/websocket"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dop251/goja"
)

const (
	EventServerStart = "onServerStart"
	EventRoomCreated = "onRoomCreated"
	EventRoomClosed  = "onRoomClosed"
	EventPlayerJoin  = "onPlayerJoin"
	EventPlayerLeave = "onPlayerLeave"
	EventGameStart   = "onGameStart"
	EventTurnChanged = "onTurnChanged"
	EventCardPlayed  = "onCardPlayed"
)

var supportedHookEvents = []string{
	EventServerStart,
	EventRoomCreated,
	EventRoomClosed,
	EventPlayerJoin,
	EventPlayerLeave,
	EventGameStart,
	EventTurnChanged,
	EventCardPlayed,
}

type serverRuntime struct {
	plugin        database.Plugin
	vm            *goja.Runtime
	mu            sync.Mutex
	onLoad        goja.Callable
	onUnload      goja.Callable
	eventHandlers map[string]goja.Callable
	pluginMeta    map[string]interface{}
	api           *goja.Object
	intervals     map[int]*time.Ticker
	intervalStops map[int]chan struct{}
	nextInterval  int
}

var (
	runtimesMu sync.Mutex
	runtimes   = make(map[uint]*serverRuntime)
)

// LoadServerScripts loads all active server plugin scripts.
func LoadServerScripts() {
	runtimesMu.Lock()
	old := runtimes
	runtimes = make(map[uint]*serverRuntime)
	runtimesMu.Unlock()

	for _, rt := range old {
		rt.shutdown()
	}

	allPlugins, err := repository.PluginRepo.GetAllPlugins()
	if err != nil {
		log.Printf("[Plugin] failed to load server scripts: %v", err)
		return
	}

	for _, p := range allPlugins {
		if !p.IsActive || strings.TrimSpace(p.ServerScript) == "" {
			continue
		}
		rt, err := newServerRuntime(p)
		if err != nil {
			log.Printf("[Plugin] failed to initialize server script %s: %v", p.Name, err)
			continue
		}
		runtimesMu.Lock()
		runtimes[p.ID] = rt
		runtimesMu.Unlock()
	}
}

// Emit dispatches a server-side lifecycle/game event to all active plugin runtimes.
func Emit(event string, payload map[string]interface{}) {
	runtimesMu.Lock()
	snapshot := make([]*serverRuntime, 0, len(runtimes))
	for _, rt := range runtimes {
		snapshot = append(snapshot, rt)
	}
	runtimesMu.Unlock()

	for _, rt := range snapshot {
		rt.emit(event, payload)
	}
}

func newServerRuntime(p database.Plugin) (*serverRuntime, error) {
	rt := &serverRuntime{
		plugin:        p,
		vm:            goja.New(),
		eventHandlers: make(map[string]goja.Callable),
		intervals:     make(map[int]*time.Ticker),
		intervalStops: make(map[int]chan struct{}),
		nextInterval:  1,
	}

	rt.pluginMeta = map[string]interface{}{
		"id":      p.ID,
		"name":    p.Name,
		"version": p.Version,
		"author":  p.Author,
	}

	console := rt.vm.NewObject()
	_ = console.Set("log", func(call goja.FunctionCall) goja.Value {
		parts := make([]string, 0, len(call.Arguments))
		for _, arg := range call.Arguments {
			parts = append(parts, arg.String())
		}
		log.Printf("[Plugin:%s] %s", p.Name, strings.Join(parts, " "))
		return goja.Undefined()
	})
	rt.vm.Set("console", console)
	rt.vm.Set("plugin", rt.pluginMeta)

	rt.api = rt.vm.NewObject()
	_ = rt.api.Set("send", rt.sendMessageFn())
	_ = rt.api.Set("sendToAll", rt.sendToAllFn())
	_ = rt.api.Set("sendToRoom", rt.sendToRoomFn())
	_ = rt.api.Set("sendToUser", rt.sendToUserFn())
	_ = rt.api.Set("setInterval", rt.setIntervalFn())
	_ = rt.api.Set("clearInterval", rt.clearIntervalFn())
	rt.vm.Set("api", rt.api)

	exports := rt.vm.NewObject()
	rt.vm.Set("exports", exports)

	// 带看门狗超时执行脚本，防止插件脚本死循环无限阻塞 goroutine
	if err := rt.runWithTimeout(scriptExecTimeout, func() error {
		_, e := rt.vm.RunString(p.ServerScript)
		return e
	}); err != nil {
		return nil, err
	}

	if obj := exports; obj != nil {
		if fn, ok := goja.AssertFunction(obj.Get("onLoad")); ok {
			rt.onLoad = fn
		}
		if fn, ok := goja.AssertFunction(obj.Get("onUnload")); ok {
			rt.onUnload = fn
		}
		for _, event := range supportedHookEvents {
			if fn, ok := goja.AssertFunction(obj.Get(event)); ok {
				rt.eventHandlers[event] = fn
			}
		}
	}

	if rt.onLoad != nil {
		rt.call(rt.onLoad, rt.newContext("onLoad", nil))
	}

	return rt, nil
}

func (rt *serverRuntime) newContext(event string, payload interface{}) *goja.Object {
	ctx := rt.vm.NewObject()
	_ = ctx.Set("event", event)
	_ = ctx.Set("plugin", rt.pluginMeta)
	_ = ctx.Set("api", rt.api)
	if payload != nil {
		_ = ctx.Set("payload", payload)
	}
	return ctx
}

func (rt *serverRuntime) emit(event string, payload map[string]interface{}) {
	fn, ok := rt.eventHandlers[event]
	if !ok {
		return
	}
	rt.call(fn, rt.newContext(event, clonePayload(payload)))
}

func clonePayload(payload map[string]interface{}) map[string]interface{} {
	if payload == nil {
		return nil
	}
	cloned := make(map[string]interface{}, len(payload))
	for k, v := range payload {
		cloned[k] = v
	}
	return cloned
}

func (rt *serverRuntime) call(fn goja.Callable, args ...goja.Value) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	// 带看门狗超时执行事件回调，避免单个插件脚本卡死整个 goroutine
	err := rt.runWithTimeout(scriptExecTimeout, func() error {
		_, e := fn(goja.Undefined(), args...)
		return e
	})
	if err != nil {
		log.Printf("[Plugin:%s] script error: %v", rt.plugin.Name, err)
	}
}

// scriptExecTimeout 单次插件脚本执行的最大时长，超时后中断 VM。
const scriptExecTimeout = 2 * time.Second

// runWithTimeout 在给定超时内执行 goja VM 操作，超时通过 vm.Interrupt 强制中断。
// 调用方需持有 rt.mu（保证同一 VM 不会被并发执行）。
func (rt *serverRuntime) runWithTimeout(timeout time.Duration, run func() error) error {
	timer := time.AfterFunc(timeout, func() {
		rt.vm.Interrupt("script execution timeout")
	})
	defer func() {
		timer.Stop()
		rt.vm.ClearInterrupt()
	}()
	return run()
}

func (rt *serverRuntime) shutdown() {
	rt.mu.Lock()
	for id, stop := range rt.intervalStops {
		close(stop)
		delete(rt.intervalStops, id)
	}
	for id, ticker := range rt.intervals {
		ticker.Stop()
		delete(rt.intervals, id)
	}
	onUnload := rt.onUnload
	rt.mu.Unlock()

	if onUnload != nil {
		rt.call(onUnload)
	}
}

func (rt *serverRuntime) sendMessage(scope string, target interface{}, payload interface{}) {
	hub := websocket.GlobalHub
	if hub == nil {
		return
	}

	data := map[string]interface{}{
		"plugin_id": rt.plugin.ID,
		"payload":   payload,
		"scope":     scope,
	}

	msg := websocket.Message{
		Type: "plugin_message",
		Data: data,
	}

	switch scope {
	case "global":
		hub.BroadcastToAll(msg)
	case "room":
		roomID, _ := target.(string)
		if roomID != "" {
			data["room_id"] = roomID
			hub.BroadcastToRoom(roomID, msg)
		}
	case "user":
		uid := 0
		switch v := target.(type) {
		case int:
			uid = v
		case int64:
			uid = int(v)
		case float64:
			uid = int(v)
		case string:
			if parsed, err := strconv.Atoi(v); err == nil {
				uid = parsed
			}
		}
		if uid > 0 {
			data["uid"] = uid
			hub.SendToUID(uid, msg)
		}
	}
}

func (rt *serverRuntime) sendMessageFn() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return goja.Undefined()
		}
		scope := strings.ToLower(call.Argument(0).String())
		var target interface{}
		var payload interface{}
		if len(call.Arguments) == 2 {
			payload = call.Argument(1).Export()
		} else {
			target = call.Argument(1).Export()
			payload = call.Argument(2).Export()
		}
		rt.sendMessage(scope, target, payload)
		return goja.Undefined()
	}
}

func (rt *serverRuntime) sendToAllFn() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return goja.Undefined()
		}
		rt.sendMessage("global", nil, call.Argument(0).Export())
		return goja.Undefined()
	}
}

func (rt *serverRuntime) sendToRoomFn() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return goja.Undefined()
		}
		rt.sendMessage("room", call.Argument(0).String(), call.Argument(1).Export())
		return goja.Undefined()
	}
}

func (rt *serverRuntime) sendToUserFn() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return goja.Undefined()
		}
		rt.sendMessage("user", call.Argument(0).Export(), call.Argument(1).Export())
		return goja.Undefined()
	}
}

func (rt *serverRuntime) setIntervalFn() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return goja.Undefined()
		}
		fn, ok := goja.AssertFunction(call.Argument(0))
		if !ok {
			return goja.Undefined()
		}
		ms := call.Argument(1).ToInteger()
		if ms < 1000 {
			ms = 1000
		}

		rt.mu.Lock()
		id := rt.nextInterval
		rt.nextInterval++
		stop := make(chan struct{})
		rt.intervalStops[id] = stop
		ticker := time.NewTicker(time.Duration(ms) * time.Millisecond)
		rt.intervals[id] = ticker
		rt.mu.Unlock()

		go func() {
			for {
				select {
				case <-ticker.C:
					rt.call(fn)
				case <-stop:
					return
				}
			}
		}()

		return rt.vm.ToValue(id)
	}
}

func (rt *serverRuntime) clearIntervalFn() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return goja.Undefined()
		}
		id := int(call.Argument(0).ToInteger())
		rt.mu.Lock()
		if stop, ok := rt.intervalStops[id]; ok {
			close(stop)
			delete(rt.intervalStops, id)
		}
		if ticker, ok := rt.intervals[id]; ok {
			ticker.Stop()
			delete(rt.intervals, id)
		}
		rt.mu.Unlock()
		return goja.Undefined()
	}
}
