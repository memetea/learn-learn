package contextkeys

// 定义 contextKey 类型来避免键名冲突
type contextKey string

// 定义不可导出的变量，只在包内可见
var (
	// DBKey 用于存取 *gorm.DB 实例
	DBKey = contextKey("db")

	User = contextKey("user")
)
