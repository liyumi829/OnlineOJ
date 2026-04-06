package repository

import (
	"fmt"
	"log"
	"testing"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 初始化 DB 连接（带连接池）
func initTestDB() *gorm.DB {
	dsn := "root:967096489@tcp(127.0.0.1:3306)/OnlineOj?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // 打印SQL
	})
	if err != nil {
		log.Fatalf("数据库连接失败：%v", err)
	}

	// ========== 连接池配置 ==========
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("获取底层DB失败: %v", err)
	}
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Second)
	sqlDB.SetConnMaxIdleTime(10 * time.Second)

	return db
}

// 测试所有 Repository 方法（适配新表：prepend_code + 2公开+2隐藏用例）
func TestAllRepositoryMethods(t *testing.T) {
	// 1. 初始化DB & Repository
	db := initTestDB()
	repo := NewRepository(db)

	// 你现在题目从 3 开始！！！
	testProblemID := uint64(3)

	// ===================== 测试1：获取题目完整信息 =====================
	start1 := time.Now()
	problem, err := repo.GetProblemByID(testProblemID)
	cost1 := time.Since(start1)

	fmt.Println("=====================================")
	fmt.Println("测试1：GetProblemByID")
	fmt.Println("耗时：", cost1)
	if err != nil {
		t.Fatalf("失败：%v", err)
	}
	fmt.Printf("题目信息：%+v\n", problem)

	// ===================== 测试2：分页获取题目简略列表 =====================
	start2 := time.Now()
	simpleList, err := repo.GetProblemSimpleList(0, 100)
	cost2 := time.Since(start2)

	fmt.Println("=====================================")
	fmt.Println("测试2：GetProblemSimpleList (分页 0-100)")
	fmt.Println("耗时：", cost2)
	if err != nil {
		t.Fatalf("失败：%v", err)
	}
	for i, v := range simpleList {
		fmt.Printf("题目%d：ID=%d 编号=%s 标题=%s 难度=%s\n",
			i+1, v.ID, v.Number, v.Title, v.Star)
	}

	// ===================== 测试3：获取【公开样例】2个 =====================
	start3 := time.Now()
	samples, err := repo.GetSampleCases(testProblemID)
	cost3 := time.Since(start3)

	fmt.Println("=====================================")
	fmt.Println("测试3：GetSampleCases (公开样例 2个)")
	fmt.Println("耗时：", cost3)
	if err != nil {
		t.Fatalf("失败：%v", err)
	}
	for i, v := range samples {
		fmt.Printf("样例%d：input=%s | output=%s | is_sample=%v\n",
			i+1, v.Input, v.Output, v.IsSample)
	}

	// ===================== 测试4：获取【所有用例】4个 =====================
	start4 := time.Now()
	allCases, err := repo.GetAllTestCases(testProblemID)
	cost4 := time.Since(start4)

	fmt.Println("=====================================")
	fmt.Println("测试4：GetAllTestCases (全部4个)")
	fmt.Println("耗时：", cost4)
	if err != nil {
		t.Fatalf("失败：%v", err)
	}
	for i, v := range allCases {
		fmt.Printf("用例%d：input=%s | output=%s | is_sample=%v\n",
			i+1, v.Input, v.Output, v.IsSample)
	}

	// ===================== 测试5：获取所有语言模板（含prepend_code） =====================
	start5 := time.Now()
	templateList, err := repo.GetTemplateCodesByProblemID(testProblemID)
	cost5 := time.Since(start5)

	fmt.Println("=====================================")
	fmt.Println("测试5：GetTemplateCodesByProblemID")
	fmt.Println("耗时：", cost5)
	if err != nil {
		t.Fatalf("失败：%v", err)
	}

	for _, tpl := range templateList {
		fmt.Printf("===== 语言：%s =====\n", tpl.Language)
		fmt.Printf("prepend_code (头部):\n%s\n", tpl.PrependCode)
		fmt.Printf("template_code (用户代码):\n%s\n", tpl.TemplateCode)
		fmt.Printf("test_code (评测):\n%s\n", tpl.TestCode)
		fmt.Printf("Enabled: %v\n", tpl.Enabled)
		fmt.Println()
	}

	// ===================== 测试6：获取 GO 完整模板 =====================
	start6 := time.Now()
	goTpl, err := repo.GetTestCodeByLang(testProblemID, "go")
	cost6 := time.Since(start6)

	fmt.Println("=====================================")
	fmt.Println("测试6：GetTestCodeByLang (go)")
	fmt.Println("耗时：", cost6)
	if err != nil {
		t.Fatalf("GO失败：%v", err)
	}
	fmt.Printf("GO prepend:\n%s\n", goTpl.PrependCode)
	fmt.Printf("GO test:\n%s\n", goTpl.TestCode)

	// ===================== 测试7：获取 C++ 完整模板 =====================
	start7 := time.Now()
	cppTpl, err := repo.GetTestCodeByLang(testProblemID, "cpp")
	cost7 := time.Since(start7)

	fmt.Println("=====================================")
	fmt.Println("测试7：GetTestCodeByLang (cpp)")
	fmt.Println("耗时：", cost7)
	if err != nil {
		t.Fatalf("CPP失败：%v", err)
	}
	fmt.Printf("CPP prepend:\n%s\n", cppTpl.PrependCode)
	fmt.Printf("CPP test:\n%s\n", cppTpl.TestCode)

	fmt.Println("=====================================")
	fmt.Println("✅ 所有测试用例执行完成！")
}
