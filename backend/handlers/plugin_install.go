package handlers

import (
	"archive/zip"
	"bytes"
	"chemistryuno/backend/database"
	"chemistryuno/backend/game"
	"chemistryuno/backend/plugins"
	"chemistryuno/backend/repository"
	"chemistryuno/backend/websocket"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ---- .cumod 鏂囦欢鍐呴儴缁撴瀯 ----

// pluginManifest 瀵瑰簲 manifest.json
type pluginManifest struct {
	Name         string          `json:"name"`
	Version      string          `json:"version"`
	Author       string          `json:"author"`
	Description  string          `json:"description"`
	GameVersion  string          `json:"game_version"` // optional game version compatibility
	Script       string          `json:"script"`       // legacy client script entry
	ConfigSchema json.RawMessage `json:"config_schema"`
	Scripts      struct {
		Client string `json:"client"` // client script path
		Server string `json:"server"` // server script path
	} `json:"scripts"`
}

// cumodCardDef 瀵瑰簲 cards.json 涓殑鍗曞紶鍗＄墝瀹氫箟
type cumodCardDef struct {
	Symbol       string          `json:"symbol"`
	DisplayName  string          `json:"display_name"`
	EffectType   string          `json:"effect_type"`
	EffectConfig json.RawMessage `json:"effect_config"`
	DefaultCount int             `json:"default_count"`
	Color        string          `json:"color"`
}

// ---- 鏈嶅姟鍣ㄩ噸鍚姸鎬?----

var (
	restartScheduled bool
	restartMu        sync.Mutex
	restartCancel    chan struct{}
)

const restartChildDelayMs = 1500

func shouldSelfRestart() bool {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("CHEMISTRYUNO_RESTART_MODE")))
	if mode == "exit" {
		return false
	}
	if mode == "self" {
		return true
	}
	// auto: detect supervisor/systemd to avoid double-run
	for _, key := range []string{
		"SUPERVISOR_ENABLED",
		"SUPERVISOR_SERVER_URL",
		"SUPERVISOR_PROCESS_NAME",
		"INVOCATION_ID",
		"JOURNAL_STREAM",
		"SYSTEMD_EXEC_PID",
	} {
		if os.Getenv(key) != "" {
			return false
		}
	}
	return true
}

func spawnSelfRestart() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	args := os.Args[1:]
	cmd := exec.Command(exe, args...)
	if _, ok := os.LookupEnv("CHEMISTRYUNO_START_DELAY_MS"); ok {
		cmd.Env = os.Environ()
	} else {
		cmd.Env = append(os.Environ(), fmt.Sprintf("CHEMISTRYUNO_START_DELAY_MS=%d", restartChildDelayMs))
	}
	if wd, err := os.Getwd(); err == nil {
		cmd.Dir = wd
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Start(); err != nil {
		return err
	}
	_ = cmd.Process.Release()
	return nil
}

// ---- 甯搁噺 ----

const (
	maxCumodFileSize  = 10 << 20 // 10MB: max .cumod size
	maxCumodInnerFile = 2 << 20  // 2MB: max single file size after unzip
)

var errZipFileNotFound = errors.New("zip file not found")

func readZipFile(zipReader *zip.Reader, name string) ([]byte, error) {
	for _, zf := range zipReader.File {
		if zf.Name == name {
			rc, err := zf.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			data, readErr := io.ReadAll(io.LimitReader(rc, maxCumodInnerFile))
			if readErr != nil {
				return nil, readErr
			}
			return data, nil
		}
	}
	return nil, errZipFileNotFound
}

func normalizeScriptPath(p string) (string, error) {
	path := strings.TrimSpace(p)
	if path == "" {
		return "", nil
	}
	path = strings.TrimPrefix(path, "./")
	if strings.Contains(path, "..") || strings.HasPrefix(path, "/") || strings.HasPrefix(path, "\\") || strings.Contains(path, "\\") {
		return "", errors.New("invalid path")
	}
	return path, nil
}

// ---- 澶勭悊鍣?----

// InstallPlugin 浠?.cumod 鏂囦欢瀹夎鎻掍欢
// POST /api/admin/plugins/install  (multipart/form-data, field: "file")
func InstallPlugin(c *gin.Context) {
	uid, _ := c.Get("uid")
	authorUID := 0
	switch v := uid.(type) {
	case int:
		authorUID = v
	case uint:
		authorUID = int(v)
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "璇蜂笂浼?.cumod 鏂囦欢"})
		return
	}
	if !strings.HasSuffix(strings.ToLower(fileHeader.Filename), ".cumod") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "鏂囦欢蹇呴』涓?.cumod 鏍煎紡"})
		return
	}
	// 鏂囦欢澶у皬鍓嶇疆妫€鏌ワ紙闃?OOM DoS锛?
	if fileHeader.Size > maxCumodFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("鏂囦欢涓嶈兘瓒呰繃 %dMB", maxCumodFileSize>>20)})
		return
	}

	f, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "璇诲彇鏂囦欢澶辫触"})
		return
	}
	defer f.Close()

	// 浣跨敤 LimitReader 鍙岄噸淇濋櫓锛岄槻姝?Content-Length 琚吉閫?
	fileBytes, err := io.ReadAll(io.LimitReader(f, maxCumodFileSize+1))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "璇诲彇鏂囦欢鍐呭澶辫触"})
		return
	}
	if int64(len(fileBytes)) > maxCumodFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("鏂囦欢涓嶈兘瓒呰繃 %dMB", maxCumodFileSize>>20)})
		return
	}

	// SHA256 闃查噸澶嶅畨瑁?
	hashBytes := sha256.Sum256(fileBytes)
	cumodHash := hex.EncodeToString(hashBytes[:])
	if existing, _ := repository.PluginRepo.GetPluginByHash(cumodHash); existing != nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": fmt.Sprintf("插件已安装 (ID: %d, 名称: %s)", existing.ID, existing.Name),
		})
		return
	}

	// 瑙ｆ瀽 ZIP 褰掓。
	zipReader, err := zip.NewReader(bytes.NewReader(fileBytes), int64(len(fileBytes)))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无法解析 .cumod 文件（损坏的 ZIP 归档）"})
		return
	}

	var manifest *pluginManifest
	var cardDefs []cumodCardDef
	var clientScriptBytes []byte
	var serverScriptBytes []byte

	for _, zf := range zipReader.File {
		switch zf.Name {
		case "manifest.json":
			rc, err := zf.Open()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "璇诲彇 manifest.json 澶辫触"})
				return
			}
			data, readErr := io.ReadAll(io.LimitReader(rc, maxCumodInnerFile))
			rc.Close()
			if readErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "璇诲彇 manifest.json 鍐呭澶辫触"})
				return
			}
			manifest = &pluginManifest{}
			data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
			if err := json.Unmarshal(data, manifest); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "manifest.json 鏍煎紡閿欒: " + err.Error()})
				return
			}

		case "cards.json":
			rc, err := zf.Open()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "璇诲彇 cards.json 澶辫触"})
				return
			}
			data, readErr := io.ReadAll(io.LimitReader(rc, maxCumodInnerFile))
			rc.Close()
			if readErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "璇诲彇 cards.json 鍐呭澶辫触"})
				return
			}
			data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
			if err := json.Unmarshal(data, &cardDefs); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "cards.json 鏍煎紡閿欒: " + err.Error()})
				return
			}
		}
	}

	// 鏍￠獙 manifest.json锛堝繀椤诲瓨鍦級
	if manifest == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "闈炴硶鎻掍欢锛氱己灏?manifest.json"})
		return
	}
	if strings.TrimSpace(manifest.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "manifest.json 涓?name 瀛楁涓嶈兘涓虹┖"})
		return
	}
	if strings.TrimSpace(manifest.Version) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "manifest.json 涓?version 瀛楁涓嶈兘涓虹┖"})
		return
	}

	normalizedSchema, parsedSchema, schemaErr := normalizePluginConfigSchema(manifest.ConfigSchema)
	if schemaErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "manifest.json 中 config_schema 非法: " + schemaErr.Error()})
		return
	}

	// 瑙ｆ瀽鑴氭湰鍏ュ彛锛堝彲閫夛級
	clientPath, err := normalizeScriptPath(manifest.Scripts.Client)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "manifest.json 涓?scripts.client 璺緞闈炴硶"})
		return
	}
	if clientPath == "" {
		clientPath, err = normalizeScriptPath(manifest.Script)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "manifest.json 涓?script 璺緞闈炴硶"})
			return
		}
	}
	serverPath, err := normalizeScriptPath(manifest.Scripts.Server)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "manifest.json 涓?scripts.server 璺緞闈炴硶"})
		return
	}

	if clientPath != "" {
		data, readErr := readZipFile(zipReader, clientPath)
		if readErr == errZipFileNotFound {
			c.JSON(http.StatusBadRequest, gin.H{"error": "manifest.json 中 client 脚本文件不存在"})
			return
		}
		if readErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "璇诲彇 client 鑴氭湰鏂囦欢澶辫触"})
			return
		}
		clientScriptBytes = data
	} else {
		if data, readErr := readZipFile(zipReader, "script.js"); readErr == nil {
			clientScriptBytes = data
		} else if readErr != errZipFileNotFound {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "璇诲彇 script.js 澶辫触"})
			return
		} else if data, readErr := readZipFile(zipReader, "client.js"); readErr == nil {
			clientScriptBytes = data
		} else if readErr != errZipFileNotFound {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "璇诲彇 client.js 澶辫触"})
			return
		}
	}

	if serverPath != "" {
		data, readErr := readZipFile(zipReader, serverPath)
		if readErr == errZipFileNotFound {
			c.JSON(http.StatusBadRequest, gin.H{"error": "manifest.json 中 server 脚本文件不存在"})
			return
		}
		if readErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "璇诲彇 server 鑴氭湰鏂囦欢澶辫触"})
			return
		}
		serverScriptBytes = data
	} else {
		if data, readErr := readZipFile(zipReader, "server.js"); readErr == nil {
			serverScriptBytes = data
		} else if readErr != errZipFileNotFound {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "璇诲彇 server.js 澶辫触"})
			return
		}
	}

	// 鏍￠獙鎵€鏈夊崱鐗屽畾涔?
	validTypes := map[string]bool{"swap": true, "force_play": true, "convert": true}
	for i, cd := range cardDefs {
		symbol := strings.ToUpper(strings.TrimSpace(cd.Symbol))
		if symbol == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("cards.json 绗?%d 涓崱鐗岀己灏?symbol 瀛楁", i+1)})
			return
		}
		if isBuiltinDeckCardSymbol(symbol) {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("cards.json symbol %s conflicts with builtin deck cards", symbol)})
			return
		}
		if !validTypes[cd.EffectType] {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("鍗＄墝 %s 鐨?effect_type 鏃犳晥锛屽繀椤讳负 swap / force_play / convert", cd.Symbol)})
			return
		}
		if err := validateEffectConfig(cd.EffectType, cd.EffectConfig); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("鍗＄墝 %s 鐨?effect_config 閿欒: %v", cd.Symbol, err)})
			return
		}
	}

	// 鍐欏叆 Plugin 璁板綍
	plugin := &database.Plugin{
		Name:         strings.TrimSpace(manifest.Name),
		Description:  manifest.Description,
		Author:       manifest.Author,
		Version:      strings.TrimSpace(manifest.Version),
		CumodHash:    cumodHash,
		AuthorUID:    authorUID,
		Script:       string(clientScriptBytes),
		ServerScript: string(serverScriptBytes),
		ConfigSchema: normalizedSchema,
		IsActive:     true,
		CreatedAt:    time.Now(),
	}
	if err := repository.PluginRepo.CreatePlugin(plugin); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "鍒涘缓鎻掍欢璁板綍澶辫触: " + err.Error()})
		return
	}

	// 鍐欏叆 PluginCard 璁板綍
	var createdCards []*database.PluginCard
	for _, cd := range cardDefs {
		defaultCount := cd.DefaultCount
		if defaultCount <= 0 {
			defaultCount = 2
		}
		if defaultCount > 20 {
			defaultCount = 20
		}
		card := &database.PluginCard{
			PluginID:     plugin.ID,
			Symbol:       strings.ToUpper(strings.TrimSpace(cd.Symbol)),
			DisplayName:  cd.DisplayName,
			EffectType:   cd.EffectType,
			EffectConfig: string(cd.EffectConfig),
			DefaultCount: defaultCount,
			Color:        cd.Color,
			CreatedAt:    time.Now(),
		}
		if err := repository.PluginRepo.CreateCard(card); err != nil {
			log.Printf("[Plugin] failed to create card %s: %v (skipped)", cd.Symbol, err)
			continue
		}
		createdCards = append(createdCards, card)
	}

	for _, field := range parsedSchema {
		defaultValue := convertDefaultValue(field)
		if defaultValue == "" {
			continue
		}
		if err := upsertPluginSetting(plugin.ID, field.Key, defaultValue); err != nil {
			log.Printf("[Plugin] failed to seed setting %s for plugin %d: %v", field.Key, plugin.ID, err)
		}
	}

	game.LoadPluginCards()
	plugins.LoadServerScripts()
	log.Printf("[Plugin] installed %s v%s, cards=%d", manifest.Name, manifest.Version, len(createdCards))
	c.JSON(http.StatusCreated, gin.H{
		"plugin": plugin,
		"cards":  createdCards,
		"count":  len(createdCards),
	})
}

// GetPluginsWithCards 鑾峰彇鎵€鏈夋縺娲绘彃浠跺強鍏跺崱鐗岋紙鐢ㄦ埛鍙瑙嗗浘锛?
// GET /api/plugins
func GetPluginsWithCards(c *gin.Context) {
	plugins, err := repository.PluginRepo.GetAllPlugins()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "鑾峰彇鎻掍欢鍒楄〃澶辫触"})
		return
	}

	type pluginWithCards struct {
		database.Plugin
		Cards           []database.PluginCard `json:"cards"`
		HasScript       bool                  `json:"has_script"`
		HasClientScript bool                  `json:"has_client_script"`
		HasServerScript bool                  `json:"has_server_script"`
	}

	result := make([]pluginWithCards, 0)
	for _, p := range plugins {
		if !p.IsActive {
			continue
		}
		cards, _ := repository.PluginRepo.GetCardsByPlugin(p.ID)
		if cards == nil {
			cards = []database.PluginCard{}
		}
		hasClient := strings.TrimSpace(p.Script) != ""
		hasServer := strings.TrimSpace(p.ServerScript) != ""
		result = append(result, pluginWithCards{
			Plugin:          p,
			Cards:           cards,
			HasScript:       hasClient,
			HasClientScript: hasClient,
			HasServerScript: hasServer,
		})
	}
	c.JSON(http.StatusOK, result)
}

// GetPluginScript 鑾峰彇鎻掍欢鑴氭湰锛堜粎闄愬凡婵€娲绘彃浠讹級
// GET /api/plugins/:id/script
func GetPluginScript(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "鏃犳晥鐨勬彃浠?ID"})
		return
	}
	plugin, err := repository.PluginRepo.GetPlugin(uint(id))
	if err != nil || plugin == nil || !plugin.IsActive {
		c.JSON(http.StatusNotFound, gin.H{"error": "插件不存在或未启用"})
		return
	}
	if strings.TrimSpace(plugin.Script) == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "插件未提供客户端脚本"})
		return
	}
	c.Header("Content-Type", "application/javascript; charset=utf-8")
	c.String(http.StatusOK, plugin.Script)
}

// ScheduleServerRestart 鍊掕鏃堕噸鍚湇鍔″櫒
// POST /api/admin/server/restart
// Body: { "delay_seconds": 30, "reason": "..." }
func ScheduleServerRestart(c *gin.Context) {
	var req struct {
		DelaySeconds int    `json:"delay_seconds"`
		Reason       string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "鍙傛暟閿欒: " + err.Error()})
		return
	}
	if req.DelaySeconds < 10 {
		req.DelaySeconds = 10
	}
	if req.DelaySeconds > 300 {
		req.DelaySeconds = 300
	}
	if req.Reason == "" {
		req.Reason = "管理员操作，服务器即将重启"
	}

	restartMu.Lock()
	if restartScheduled {
		restartMu.Unlock()
		c.JSON(http.StatusConflict, gin.H{"error": "宸叉湁閲嶅惎璁″垝杩涜涓紝濡傞渶鍙栨秷璇疯皟鐢?/cancel"})
		return
	}
	restartScheduled = true
	restartCancel = make(chan struct{})
	restartMu.Unlock()

	hub := websocket.GlobalHub
	restartAt := time.Now().Add(time.Duration(req.DelaySeconds) * time.Second)

	if hub != nil {
		hub.BroadcastToAll(websocket.Message{
			Type: "server_restart",
			Data: map[string]interface{}{
				"seconds":    req.DelaySeconds,
				"reason":     req.Reason,
				"restart_at": restartAt.UnixMilli(),
			},
		})
	}

	log.Printf("[Server] 鈿狅笍  鏈嶅姟鍣ㄥ皢鍦?%d 绉掑悗閲嶅惎锛屽師鍥? %s", req.DelaySeconds, req.Reason)

	cancel := restartCancel
	go func() {
		select {
		case <-time.After(time.Duration(req.DelaySeconds) * time.Second):
			if hub != nil {
				hub.BroadcastToAll(websocket.Message{
					Type: "server_restart_now",
					Data: "鏈嶅姟鍣ㄦ鍦ㄩ噸鍚紝璇风◢鍊?..",
				})
			}
			time.Sleep(2 * time.Second)
			log.Println("[Server] 馃攧 鎵ц閲嶅惎...")
			if shouldSelfRestart() {
				if err := spawnSelfRestart(); err != nil {
					log.Printf("[Server] 鉂?鑷惎鍔ㄩ噸鍚け璐? %v (灏嗙洿鎺ラ€€鍑?", err)
				} else {
					log.Println("[Server] new process started, exiting current process")
				}
			}
			os.Exit(0)
		case <-cancel:
			log.Println("[Server] restart cancelled")
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"message":    fmt.Sprintf("鏈嶅姟鍣ㄥ皢鍦?%d 绉掑悗閲嶅惎", req.DelaySeconds),
		"restart_at": restartAt.UnixMilli(),
	})
}

// CancelServerRestart 鍙栨秷宸叉帓绋嬬殑閲嶅惎
// POST /api/admin/server/restart/cancel
func CancelServerRestart(c *gin.Context) {
	restartMu.Lock()
	defer restartMu.Unlock()

	if !restartScheduled {
		c.JSON(http.StatusNotFound, gin.H{"error": "褰撳墠娌℃湁杩涜涓殑閲嶅惎璁″垝"})
		return
	}
	close(restartCancel)
	restartScheduled = false

	hub := websocket.GlobalHub
	if hub != nil {
		hub.BroadcastToAll(websocket.Message{
			Type: "server_restart_cancelled",
			Data: "管理员已取消服务器重启",
		})
	}
	c.JSON(http.StatusOK, gin.H{"message": "重启已取消"})
}
