package repository

// Repository包提供数据访问层抽象
//
// 本包将数据库操作从业务逻辑中分离出来，提供以下优势：
//
// 1. 数据库无关性 - 使用GORM支持多种数据库
// 2. 代码复用 - 常见查询模式封装为方法
// 3. 类型安全 - 使用Go结构体替代SQL字符串
// 4. 易于测试 - 可以轻松mock Repository接口
// 5. 关注点分离 - Handler专注业务逻辑，Repository处理数据访问
//
// 使用示例：
//
//	userRepo := repository.NewUserRepository()
//	user, err := userRepo.FindByUsername("alice")
//	if err != nil {
//	    return err
//	}
//	err = userRepo.UpdatePassword(user.UID, newHash)
