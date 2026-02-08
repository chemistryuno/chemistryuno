package main

import (
	"fmt"
	"log"
	"os"

	"chemistryuno/database"
)

func main() {
	log.Println("=== 数据库初始化脚本 ===")

	// 切换到 backend 目录（如果需要）
	if _, err := os.Stat("./backend"); err == nil {
		if err := os.Chdir("./backend"); err != nil {
			log.Fatalf("切换目录失败: %v", err)
		}
		log.Println("📂 切换到 backend 目录")
	}

	// 检查数据库文件是否存在
	dbPath := "./chemistryuno.db"
	if _, err := os.Stat(dbPath); err == nil {
		log.Printf("⚠️  警告: 数据库文件 %s 已存在", dbPath)
		log.Println("正在删除旧数据库...")
		if err := os.Remove(dbPath); err != nil {
			log.Printf("\n❌ 错误: 删除旧数据库失败: %v\n", err)
			log.Println("\n可能的原因：")
			log.Println("  • 后端服务正在运行，请先停止服务")
			log.Println("  • 数据库文件被其他程序占用")
			log.Println("  • 文件权限不足")
			log.Println("\n解决方案：")
			log.Println("  1. 如果后端正在运行，请先停止（Ctrl+C）")
			log.Println("  2. 确保没有其他程序正在访问数据库文件")
			log.Println("  3. 关闭所有可能占用数据库的进程，然后重试")
			os.Exit(1)
		}
		log.Println("✅ 旧数据库已删除")
	}

	// 同时删除 WAL 和 SHM 文件（如果存在）
	for _, suffix := range []string{"-wal", "-shm"} {
		walPath := dbPath + suffix
		if _, err := os.Stat(walPath); err == nil {
			if err := os.Remove(walPath); err != nil {
				log.Printf("⚠️  警告: 无法删除 %s: %v", walPath, err)
			} else {
				log.Printf("✅ 已删除 %s", walPath)
			}
		}
	}

	log.Println("\n开始初始化数据库...")

	// 初始化数据库（会自动运行迁移和默认数据）
	if err := database.InitDB(dbPath); err != nil {
		log.Fatalf("❌ 数据库初始化失败: %v", err)
	}

	log.Println("\n=== 数据库初始化完成 ===")
	fmt.Println("✅ 数据库已成功初始化")
	fmt.Println("📊 默认管理员账户:")
	fmt.Println("   用户名: admin@chemistryuno.com")
	fmt.Println("   密码: admin123")
	fmt.Println("\n⚠️  请在生产环境中立即修改默认密码！")
}
