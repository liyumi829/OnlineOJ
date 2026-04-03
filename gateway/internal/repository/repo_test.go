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

// 初始化 DB 连接（你的数据库信息）
func initTestDB() *gorm.DB {
	dsn := "root:967096489@tcp(127.0.0.1:3306)/OnlineOj?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // 打印SQL
	})
	if err != nil {
		log.Fatalf("数据库连接失败：%v", err)
	}
	return db
}

// 测试所有Repository方法
func TestAllRepositoryMethods(t *testing.T) {
	// 1. 初始化DB & Repository
	db := initTestDB()
	repo := NewRepository(db)

	// 测试用题目ID（你之前插入的 回文数 LC-9）
	// testProblemID := uint64(1)
	testProblemID := uint64(2) // 测试题目2

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

	// ===================== 测试2：获取题目简略列表 =====================
	start2 := time.Now()
	simpleList, err := repo.GetProblemSimpleList()
	cost2 := time.Since(start2)

	fmt.Println("=====================================")
	fmt.Println("测试2：GetProblemSimpleList")
	fmt.Println("耗时：", cost2)
	if err != nil {
		t.Fatalf("失败：%v", err)
	}
	for i, v := range simpleList {
		fmt.Printf("题目%d：%+v\n", i+1, v)
	}

	// ===================== 测试3：获取样例测试用例 =====================
	start3 := time.Now()
	samples, err := repo.GetSampleCases(testProblemID)
	cost3 := time.Since(start3)

	fmt.Println("=====================================")
	fmt.Println("测试3：GetSampleCases")
	fmt.Println("耗时：", cost3)
	if err != nil {
		t.Fatalf("失败：%v", err)
	}
	for i, v := range samples {
		fmt.Printf("样例%d：input=%s\toutput=%s\n", i+1, v.Input, v.Output)
	}

	// ===================== 测试4：获取所有测试用例 =====================
	start4 := time.Now()
	allCases, err := repo.GetAllCases(testProblemID)
	cost4 := time.Since(start4)

	fmt.Println("=====================================")
	fmt.Println("测试4：GetAllCases")
	fmt.Println("耗时：", cost4)
	if err != nil {
		t.Fatalf("失败：%v", err)
	}
	for i, v := range allCases {
		fmt.Printf("用例%d：input=%s\toutput=%s\n", i+1, v.Input, v.Output)
	}

	// ===================== 测试5：获取Go语言模板代码 =====================
	start5 := time.Now()
	goTpl, err := repo.GetTemplateCode(testProblemID, "go")
	cost5 := time.Since(start5)

	fmt.Println("=====================================")
	fmt.Println("测试5：GetTemplateCode (go)")
	fmt.Println("耗时：", cost5)
	if err != nil {
		t.Fatalf("失败：%v", err)
	}
	fmt.Printf("是否启用：%v\n代码为空:%v\n代码:\n%s\n", goTpl.Enabled, goTpl.TemplateCode == "", goTpl.TemplateCode)

	// ===================== 测试6：获取Go测试代码 =====================
	start6 := time.Now()
	goTestCode, err := repo.GetTestCode(testProblemID, "go")
	cost6 := time.Since(start6)

	fmt.Println("=====================================")
	fmt.Println("测试6：GetTestCode (go)")
	fmt.Println("耗时：", cost6)
	if err != nil {
		t.Fatalf("失败：%v", err)
	}
	fmt.Printf("测试代码：\n%s\n", goTestCode)

	// ===================== 测试7：获取C++模板代码 =====================
	start7 := time.Now()
	cppTpl, err := repo.GetTemplateCode(testProblemID, "cpp")
	cost7 := time.Since(start7)

	fmt.Println("=====================================")
	fmt.Println("测试7：GetTemplateCode (cpp)")
	fmt.Println("耗时：", cost7)
	if err != nil {
		t.Fatalf("失败：%v", err)
	}
	fmt.Printf("是否启用：%v\n代码为空:%v\n代码:\n%s\n", cppTpl.Enabled, cppTpl.TemplateCode == "", cppTpl.TemplateCode)

	// ===================== 测试8：获取C++测试代码 =====================
	start8 := time.Now()
	cppTestCode, err := repo.GetTestCode(testProblemID, "cpp")
	cost8 := time.Since(start8)

	fmt.Println("=====================================")
	fmt.Println("测试8：GetTestCode (cpp)")
	fmt.Println("耗时：", cost8)
	if err != nil {
		t.Fatalf("失败：%v", err)
	}
	fmt.Printf("测试代码：\n%s\n", cppTestCode)

	fmt.Println("=====================================")
	fmt.Println("✅ 所有测试用例执行完成！")
}
