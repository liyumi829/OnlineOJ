-- =========================
-- OnlineOj 初始化 SQL
-- 作用：
-- 1. 创建数据库
-- 2. 创建题目表、语言模板表、测试用例表
-- 3. 插入 17 道题目的基础数据
-- 4. 插入测试用例
-- 5. 插入 Go / C++ 代码模板
-- =========================

DROP DATABASE IF EXISTS `OnlineOj`;
CREATE DATABASE `OnlineOj`
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

USE `OnlineOj`;

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- =========================
-- 删除旧表
-- =========================
DROP TABLE IF EXISTS `problem_language_template`;
DROP TABLE IF EXISTS `problem_test_case`;
DROP TABLE IF EXISTS `problem`;

-- =========================
-- 题目主表
-- =========================
CREATE TABLE `problem` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `number` VARCHAR(64) NOT NULL COMMENT '题目编号',
  `title` VARCHAR(255) NOT NULL COMMENT '题目标题',
  `star` VARCHAR(32) NOT NULL COMMENT '题目难度：Easy/Medium/Hard',
  `cpulimit` INT NOT NULL DEFAULT 500000000 COMMENT '时间限制',
  `memlimit` INT NOT NULL DEFAULT 65536 COMMENT '内存限制',
  `description` TEXT NOT NULL COMMENT '题目描述',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_problem_number` (`number`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='题目基础信息表';

-- =========================
-- 题目多语言模板表
-- prepend_code：隐藏头部代码
-- template_code：给用户填写的代码
-- test_code：评测入口代码
-- =========================
CREATE TABLE `problem_language_template` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `problem_id` BIGINT UNSIGNED NOT NULL COMMENT '题目ID',
  `language` VARCHAR(32) NOT NULL COMMENT '编程语言',
  `prepend_code` MEDIUMTEXT NOT NULL COMMENT '隐藏的头部代码',
  `template_code` MEDIUMTEXT NOT NULL COMMENT '模板代码',
  `test_code` MEDIUMTEXT NOT NULL COMMENT '测试代码',
  `enabled` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否启用',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_problem_language` (`problem_id`, `language`),
  KEY `idx_language` (`language`),
  CONSTRAINT `fk_problem_language_template_problem`
    FOREIGN KEY (`problem_id`) REFERENCES `problem` (`id`)
    ON DELETE CASCADE
    ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='题目多语言模板表';

-- =========================
-- 题目测试用例表
-- =========================
CREATE TABLE `problem_test_case` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '测试用例主键',
  `problem_id` BIGINT UNSIGNED NOT NULL COMMENT '题目ID',
  `input` MEDIUMTEXT NOT NULL COMMENT '输入',
  `output` MEDIUMTEXT NOT NULL COMMENT '输出',
  `is_sample` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否样例',
  `sort_order` INT NOT NULL DEFAULT 1 COMMENT '排序',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_problem_test_case_problem_id` (`problem_id`),
  KEY `idx_problem_test_case_problem_sort` (`problem_id`, `sort_order`),
  CONSTRAINT `fk_problem_test_case_problem`
    FOREIGN KEY (`problem_id`) REFERENCES `problem` (`id`)
    ON DELETE CASCADE
    ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='题目测试用例表';

START TRANSACTION;

-- =========================
-- 插入题目基础数据
-- 难度统一使用英文
-- =========================
INSERT INTO `problem`
(`id`, `number`, `title`, `star`, `cpulimit`, `memlimit`, `description`)
VALUES
(1, 'OJ-0001', '回文数', 'Easy', 500000000, 65536, '给定一个整数 x，如果 x 是一个回文整数，返回 true，否则返回 false。'),
(2, 'OJ-0002', '两数之和', 'Easy', 500000000, 65536, '给定一个整数数组 nums 和一个整数目标值 target，请你在该数组中找出和为目标值 target 的那两个整数，并返回它们的数组下标。'),
(3, 'OJ-0003', '删除有序数组中的重复项', 'Easy', 500000000, 65536, '给定一个排序数组，原地删除重复元素，返回新数组长度。'),
(4, 'OJ-0004', '移除元素', 'Easy', 500000000, 65536, '给定一个数组和一个值，原地移除所有等于该值的元素，返回新长度。'),
(5, 'OJ-0005', '搜索插入位置', 'Easy', 500000000, 65536, '给定排序数组和目标值，找到目标位置，不存在则返回应插入位置。'),
(6, 'OJ-0006', '最大子数组和', 'Medium', 500000000, 65536, '找出数组中具有最大和的连续子数组，返回其最大和。'),
(7, 'OJ-0007', '合并两个有序数组', 'Easy', 500000000, 65536, '合并两个有序数组，使结果也为有序数组。'),
(8, 'OJ-0008', '杨辉三角', 'Easy', 500000000, 65536, '给定非负整数 numRows，生成杨辉三角的前 numRows 行。'),
(9, 'OJ-0009', '买卖股票的最佳时机', 'Easy', 500000000, 65536, '选择一天买入、另一天卖出，计算能获取的最大利润。'),
(10, 'OJ-0010', '只出现一次的数字', 'Easy', 500000000, 65536, '数组中除一个元素外均出现两次，找出只出现一次的数字。'),
(11, 'OJ-0011', '多数元素', 'Easy', 500000000, 65536, '找出数组中出现次数大于 n/2 的元素。'),
(12, 'OJ-0012', '存在重复元素', 'Easy', 500000000, 65536, '判断数组中是否存在至少出现两次的元素。'),
(13, 'OJ-0013', '丢失的数字', 'Easy', 500000000, 65536, '给定 [0,n] 的数组，找出其中缺失的数字。'),
(14, 'OJ-0014', '移动零', 'Easy', 500000000, 65536, '将数组中所有 0 移动到末尾，保持非零元素相对顺序。'),
(15, 'OJ-0015', '找到所有数组中消失的数字', 'Easy', 500000000, 65536, '找出 1~n 中没有出现在数组中的所有数字。'),
(16, 'OJ-0016', '最大连续1的个数', 'Easy', 500000000, 65536, '给定二进制数组，返回最大连续 1 的个数。'),
(17, 'OJ-0017', '数组拆分 I', 'Easy', 500000000, 65536, '将长度为 2n 的数组分为 n 对，使每对最小值之和最大。');

-- =========================
-- 插入测试用例
-- =========================
INSERT INTO `problem_test_case`
(`problem_id`, `input`, `output`, `is_sample`, `sort_order`)
VALUES
(1,'121','true',1,1),
(1,'-121','false',1,2),
(1,'10','false',0,3),
(1,'0','true',0,4),

(2,'[2,7,11,15]\n9','[0,1]',1,1),
(2,'[3,2,4]\n6','[1,2]',1,2),
(2,'[3,3]\n6','[0,1]',0,3),
(2,'[1,5,3,7]\n8','[0,3]',0,4),

(3,'[1,1,2]','2',1,1),
(3,'[0,0,1,1,1,2,2,3,3,4]','5',1,2),
(3,'[1,2,2,3,3,3]','3',0,3),
(3,'[0,1,2,3,4]','5',0,4),

(4,'[3,2,2,3]\n3','2',1,1),
(4,'[0,1,2,2,3,0,4,2]\n2','5',1,2),
(4,'[1,2,1,1]\n1','1',0,3),
(4,'[0,0,0,1]\n0','1',0,4),

(5,'[1,3,5,6]\n5','2',1,1),
(5,'[1,3,5,6]\n2','1',1,2),
(5,'[1,3,5,6]\n7','4',0,3),
(5,'[2,3,5,7]\n1','0',0,4),

(6,'[-2,1,-3,4,-1,2,1,-5,4]','6',1,1),
(6,'[5,4,-1,7,8]','23',1,2),
(6,'[1]','1',0,3),
(6,'[-1,-2,-3]','-1',0,4),

(7,'[1,2,3,0,0,0]\n3\n[2,5,6]\n3','[1,2,2,3,5,6]',1,1),
(7,'[0]\n0\n[1]\n1','[1]',1,2),
(7,'[2,0]\n1\n[1]\n1','[1,2]',0,3),
(7,'[4,5,6,0,0,0]\n3\n[1,2,3]\n3','[1,2,3,4,5,6]',0,4),

(8,'3','[[1],[1,1],[1,2,1]]',1,1),
(8,'5','[[1],[1,1],[1,2,1],[1,3,3,1],[1,4,6,4,1]]',1,2),
(8,'1','[[1]]',0,3),
(8,'2','[[1],[1,1]]',0,4),

(9,'[7,1,5,3,6,4]','5',1,1),
(9,'[7,6,4,3,1]','0',1,2),
(9,'[1,2,3,4]','3',0,3),
(9,'[5,3,4,1]','1',0,4),

(10,'[2,2,1]','1',1,1),
(10,'[4,1,2,1,2]','4',1,2),
(10,'[1]','1',0,3),
(10,'[0,0,1]','1',0,4),

(11,'[3,2,3]','3',1,1),
(11,'[2,2,1,1,1,2,2]','2',1,2),
(11,'[1]','1',0,3),
(11,'[2,2,2]','2',0,4),

(12,'[1,2,3,1]','true',1,1),
(12,'[1,2,3,4]','false',1,2),
(12,'[2,2]','true',0,3),
(12,'[1,2,3,4,5]','false',0,4),

(13,'[3,0,1]','2',1,1),
(13,'[0,1]','1',1,2),
(13,'[1,2]','0',0,3),
(13,'[0]','1',0,4),

(14,'[0,1,0,3,12]','[1,3,12,0,0]',1,1),
(14,'[1,0,3,0,12,0]','[1,3,12,0,0,0]',1,2),
(14,'[0,0,1]','[1,0,0]',0,3),
(14,'[1,0,2,0]','[1,2,0,0]',0,4),

(15,'[4,3,2,7,8,2,3,1]','[5,6]',1,1),
(15,'[1,1]','[2]',1,2),
(15,'[2,2,3]','[1]',0,3),
(15,'[1,2,3,4]','[]',0,4),

(16,'[1,1,0,1,1,1]','3',1,1),
(16,'[0,0,1,0,0,1,1]','2',1,2),
(16,'[1,0,1]','1',0,3),
(16,'[0,0,0]','0',0,4),

(17,'[1,4,3,2]','4',1,1),
(17,'[6,2,6,5,1,2]','9',1,2),
(17,'[1,2]','1',0,3),
(17,'[3,1,2,4]','4',0,4);

-- =========================
-- 插入语言模板
-- 最终拼接顺序：
-- prepend_code + '\n\n' + template_code + '\n\n' + test_code
-- =========================
INSERT INTO `problem_language_template`
(`problem_id`, `language`, `prepend_code`, `template_code`, `test_code`, `enabled`)
VALUES

-- =========================
-- 1. 回文数
-- =========================
(1, 'go',
'package main

import (
    "fmt"
)',
'func isPalindrome(x int) bool {
    // 在这里编写代码
    return false
}',
'func main() {
    var x int
    fmt.Scan(&x)
    fmt.Println(isPalindrome(x))
}', 1),

(1, 'cpp',
'#include <iostream>
using namespace std;',
'class Solution {
public:
    bool isPalindrome(int x) {
        // 在这里编写代码
        return false;
    }
};',
'int main() {
    int x;
    cin >> x;
    Solution sol;
    cout << boolalpha << sol.isPalindrome(x) << endl;
    return 0;
}', 1),

-- =========================
-- 2. 两数之和
-- =========================
(2, 'go',
'package main

import (
    "bufio"
    "fmt"
    "os"
    "strconv"
    "strings"
)',
'func twoSum(nums []int, target int) []int {
    // 在这里编写代码
    return nil
}',
'func parseIntArray(line string) []int {
    line = strings.TrimSpace(line)
    line = strings.Trim(line, "[]")
    if line == "" {
        return []int{}
    }

    parts := strings.Split(line, ",")
    nums := make([]int, 0, len(parts))
    for _, p := range parts {
        n, _ := strconv.Atoi(strings.TrimSpace(p))
        nums = append(nums, n)
    }
    return nums
}

func main() {
    sc := bufio.NewScanner(os.Stdin)

    sc.Scan()
    nums := parseIntArray(sc.Text())

    sc.Scan()
    target, _ := strconv.Atoi(strings.TrimSpace(sc.Text()))

    res := twoSum(nums, target)
    if len(res) >= 2 {
        fmt.Printf("[%d,%d]\n", res[0], res[1])
    } else {
        fmt.Println("[]")
    }
}', 1),

(2, 'cpp',
'#include <iostream>
#include <vector>
#include <sstream>
#include <string>
using namespace std;',
'class Solution {
public:
    vector<int> twoSum(vector<int>& nums, int target) {
        // 在这里编写代码
        return {};
    }
};',
'int main() {
    string line;
    getline(cin, line);

    if (line.size() >= 2) {
        line = line.substr(1, line.size() - 2);
    } else {
        line = "";
    }

    vector<int> nums;
    if (!line.empty()) {
        stringstream ss(line);
        string s;
        while (getline(ss, s, '','')) {
            nums.push_back(stoi(s));
        }
    }

    int target;
    cin >> target;

    Solution sol;
    vector<int> res = sol.twoSum(nums, target);

    if (res.size() >= 2) {
        cout << "[" << res[0] << "," << res[1] << "]" << endl;
    } else {
        cout << "[]" << endl;
    }

    return 0;
}', 1),

-- =========================
-- 3. 删除有序数组中的重复项
-- =========================
(3, 'go',
'package main

import (
    "bufio"
    "fmt"
    "os"
    "strconv"
    "strings"
)',
'func removeDuplicates(nums []int) int {
    return 0
}',
'func parseIntArray(line string) []int {
    line = strings.TrimSpace(line)
    line = strings.Trim(line, "[]")
    if line == "" {
        return []int{}
    }
    parts := strings.Split(line, ",")
    nums := make([]int, 0, len(parts))
    for _, p := range parts {
        n, _ := strconv.Atoi(strings.TrimSpace(p))
        nums = append(nums, n)
    }
    return nums
}

func main() {
    sc := bufio.NewScanner(os.Stdin)
    sc.Scan()
    nums := parseIntArray(sc.Text())
    fmt.Println(removeDuplicates(nums))
}', 1),

(3, 'cpp',
'#include <iostream>
#include <vector>
#include <sstream>
using namespace std;',
'class Solution {
public:
    int removeDuplicates(vector<int>& nums) {
        return 0;
    }
};',
'int main() {
    string line;
    getline(cin, line);

    if (line.size() >= 2) line = line.substr(1, line.size() - 2);
    else line = "";

    vector<int> nums;
    if (!line.empty()) {
        stringstream ss(line);
        string s;
        while (getline(ss, s, '','')) nums.push_back(stoi(s));
    }

    Solution sol;
    cout << sol.removeDuplicates(nums) << endl;
    return 0;
}', 1),

-- =========================
-- 4. 移除元素
-- =========================
(4, 'go',
'package main

import (
    "bufio"
    "fmt"
    "os"
    "strconv"
    "strings"
)',
'func removeElement(nums []int, val int) int {
    return 0
}',
'func parseIntArray(line string) []int {
    line = strings.TrimSpace(line)
    line = strings.Trim(line, "[]")
    if line == "" {
        return []int{}
    }
    parts := strings.Split(line, ",")
    nums := make([]int, 0, len(parts))
    for _, p := range parts {
        n, _ := strconv.Atoi(strings.TrimSpace(p))
        nums = append(nums, n)
    }
    return nums
}

func main() {
    sc := bufio.NewScanner(os.Stdin)
    sc.Scan()
    nums := parseIntArray(sc.Text())
    sc.Scan()
    val, _ := strconv.Atoi(strings.TrimSpace(sc.Text()))
    fmt.Println(removeElement(nums, val))
}', 1),

(4, 'cpp',
'#include <iostream>
#include <vector>
#include <sstream>
using namespace std;',
'class Solution {
public:
    int removeElement(vector<int>& nums, int val) {
        return 0;
    }
};',
'int main() {
    string line;
    getline(cin, line);

    if (line.size() >= 2) line = line.substr(1, line.size() - 2);
    else line = "";

    vector<int> nums;
    if (!line.empty()) {
        stringstream ss(line);
        string s;
        while (getline(ss, s, '','')) nums.push_back(stoi(s));
    }

    int val;
    cin >> val;

    Solution sol;
    cout << sol.removeElement(nums, val) << endl;
    return 0;
}', 1),

-- =========================
-- 5. 搜索插入位置
-- =========================
(5, 'go',
'package main

import (
    "bufio"
    "fmt"
    "os"
    "strconv"
    "strings"
)',
'func searchInsert(nums []int, target int) int {
    return 0
}',
'func parseIntArray(line string) []int {
    line = strings.TrimSpace(line)
    line = strings.Trim(line, "[]")
    if line == "" {
        return []int{}
    }
    parts := strings.Split(line, ",")
    nums := make([]int, 0, len(parts))
    for _, p := range parts {
        n, _ := strconv.Atoi(strings.TrimSpace(p))
        nums = append(nums, n)
    }
    return nums
}

func main() {
    sc := bufio.NewScanner(os.Stdin)
    sc.Scan()
    nums := parseIntArray(sc.Text())
    sc.Scan()
    target, _ := strconv.Atoi(strings.TrimSpace(sc.Text()))
    fmt.Println(searchInsert(nums, target))
}', 1),

(5, 'cpp',
'#include <iostream>
#include <vector>
#include <sstream>
using namespace std;',
'class Solution {
public:
    int searchInsert(vector<int>& nums, int target) {
        return 0;
    }
};',
'int main() {
    string line;
    getline(cin, line);

    if (line.size() >= 2) line = line.substr(1, line.size() - 2);
    else line = "";

    vector<int> nums;
    if (!line.empty()) {
        stringstream ss(line);
        string s;
        while (getline(ss, s, '','')) nums.push_back(stoi(s));
    }

    int target;
    cin >> target;

    Solution sol;
    cout << sol.searchInsert(nums, target) << endl;
    return 0;
}', 1),

-- =========================
-- 6. 最大子数组和
-- =========================
(6, 'go',
'package main

import (
    "bufio"
    "fmt"
    "os"
    "strconv"
    "strings"
)',
'func maxSubArray(nums []int) int {
    return 0
}',
'func parseIntArray(line string) []int {
    line = strings.TrimSpace(line)
    line = strings.Trim(line, "[]")
    if line == "" {
        return []int{}
    }
    parts := strings.Split(line, ",")
    nums := make([]int, 0, len(parts))
    for _, p := range parts {
        n, _ := strconv.Atoi(strings.TrimSpace(p))
        nums = append(nums, n)
    }
    return nums
}

func main() {
    sc := bufio.NewScanner(os.Stdin)
    sc.Scan()
    nums := parseIntArray(sc.Text())
    fmt.Println(maxSubArray(nums))
}', 1),

(6, 'cpp',
'#include <iostream>
#include <vector>
#include <sstream>
using namespace std;',
'class Solution {
public:
    int maxSubArray(vector<int>& nums) {
        return 0;
    }
};',
'int main() {
    string line;
    getline(cin, line);

    if (line.size() >= 2) line = line.substr(1, line.size() - 2);
    else line = "";

    vector<int> nums;
    if (!line.empty()) {
        stringstream ss(line);
        string s;
        while (getline(ss, s, '','')) nums.push_back(stoi(s));
    }

    Solution sol;
    cout << sol.maxSubArray(nums) << endl;
    return 0;
}', 1),

-- =========================
-- 7. 合并两个有序数组
-- =========================
(7, 'go',
'package main

import (
    "bufio"
    "fmt"
    "os"
    "strconv"
    "strings"
)',
'func merge(nums1 []int, m int, nums2 []int, n int) {
}',
'func parseIntArray(line string) []int {
    line = strings.TrimSpace(line)
    line = strings.Trim(line, "[]")
    if line == "" {
        return []int{}
    }
    parts := strings.Split(line, ",")
    nums := make([]int, 0, len(parts))
    for _, p := range parts {
        n, _ := strconv.Atoi(strings.TrimSpace(p))
        nums = append(nums, n)
    }
    return nums
}

func formatIntArray(nums []int) string {
    if len(nums) == 0 {
        return "[]"
    }
    var b strings.Builder
    b.WriteByte(''['')
    for i, v := range nums {
        if i > 0 {
            b.WriteByte('','')
        }
        b.WriteString(strconv.Itoa(v))
    }
    b.WriteByte('']'')
    return b.String()
}

func main() {
    sc := bufio.NewScanner(os.Stdin)

    sc.Scan()
    nums1 := parseIntArray(sc.Text())

    sc.Scan()
    m, _ := strconv.Atoi(strings.TrimSpace(sc.Text()))

    sc.Scan()
    nums2 := parseIntArray(sc.Text())

    sc.Scan()
    n, _ := strconv.Atoi(strings.TrimSpace(sc.Text()))

    merge(nums1, m, nums2, n)
    fmt.Println(formatIntArray(nums1))
}', 1),

(7, 'cpp',
'#include <iostream>
#include <vector>
#include <sstream>
#include <string>
using namespace std;',
'class Solution {
public:
    void merge(vector<int>& nums1, int m, vector<int>& nums2, int n) {
    }
};',
'static string formatVector(const vector<int>& nums) {
    string out = "[";
    for (size_t i = 0; i < nums.size(); ++i) {
        if (i) out += ",";
        out += to_string(nums[i]);
    }
    out += "]";
    return out;
}

int main() {
    string line1, line2;
    getline(cin, line1);
    int m;
    cin >> m;
    cin.ignore();
    getline(cin, line2);
    int n;
    cin >> n;

    auto parse = [](string line) {
        vector<int> nums;
        if (line.size() >= 2) line = line.substr(1, line.size() - 2);
        else line = "";
        if (!line.empty()) {
            stringstream ss(line);
            string s;
            while (getline(ss, s, '','')) nums.push_back(stoi(s));
        }
        return nums;
    };

    vector<int> nums1 = parse(line1);
    vector<int> nums2 = parse(line2);

    Solution sol;
    sol.merge(nums1, m, nums2, n);
    cout << formatVector(nums1) << endl;
    return 0;
}', 1),

-- =========================
-- 8. 杨辉三角
-- =========================
(8, 'go',
'package main

import (
    "bufio"
    "fmt"
    "os"
    "strconv"
    "strings"
)',
'func generate(numRows int) [][]int {
    return nil
}',
'func format2D(nums [][]int) string {
    if len(nums) == 0 {
        return "[]"
    }
    var b strings.Builder
    b.WriteByte(''['')
    for i, row := range nums {
        if i > 0 {
            b.WriteByte('','')
        }
        b.WriteByte(''['')
        for j, v := range row {
            if j > 0 {
                b.WriteByte('','')
            }
            b.WriteString(strconv.Itoa(v))
        }
        b.WriteByte('']'')
    }
    b.WriteByte('']'')
    return b.String()
}

func main() {
    sc := bufio.NewScanner(os.Stdin)
    sc.Scan()
    n, _ := strconv.Atoi(strings.TrimSpace(sc.Text()))
    fmt.Println(format2D(generate(n)))
}', 1),

(8, 'cpp',
'#include <iostream>
#include <vector>
#include <string>
using namespace std;',
'class Solution {
public:
    vector<vector<int>> generate(int n) {
        return {};
    }
};',
'static string format2D(const vector<vector<int>>& nums) {
    string out = "[";
    for (size_t i = 0; i < nums.size(); ++i) {
        if (i) out += ",";
        out += "[";
        for (size_t j = 0; j < nums[i].size(); ++j) {
            if (j) out += ",";
            out += to_string(nums[i][j]);
        }
        out += "]";
    }
    out += "]";
    return out;
}

int main() {
    int n;
    cin >> n;
    Solution sol;
    cout << format2D(sol.generate(n)) << endl;
    return 0;
}', 1),

-- =========================
-- 9. 买卖股票的最佳时机
-- =========================
(9, 'go',
'package main

import (
    "bufio"
    "fmt"
    "os"
    "strconv"
    "strings"
)',
'func maxProfit(prices []int) int {
    return 0
}',
'func parseIntArray(line string) []int {
    line = strings.TrimSpace(line)
    line = strings.Trim(line, "[]")
    if line == "" {
        return []int{}
    }
    parts := strings.Split(line, ",")
    nums := make([]int, 0, len(parts))
    for _, p := range parts {
        n, _ := strconv.Atoi(strings.TrimSpace(p))
        nums = append(nums, n)
    }
    return nums
}

func main() {
    sc := bufio.NewScanner(os.Stdin)
    sc.Scan()
    prices := parseIntArray(sc.Text())
    fmt.Println(maxProfit(prices))
}', 1),

(9, 'cpp',
'#include <iostream>
#include <vector>
#include <sstream>
using namespace std;',
'class Solution {
public:
    int maxProfit(vector<int>& prices) {
        return 0;
    }
};',
'int main() {
    string line;
    getline(cin, line);

    if (line.size() >= 2) line = line.substr(1, line.size() - 2);
    else line = "";

    vector<int> prices;
    if (!line.empty()) {
        stringstream ss(line);
        string s;
        while (getline(ss, s, '','')) prices.push_back(stoi(s));
    }

    Solution sol;
    cout << sol.maxProfit(prices) << endl;
    return 0;
}', 1),

-- =========================
-- 10. 只出现一次的数字
-- =========================
(10, 'go',
'package main

import (
    "bufio"
    "fmt"
    "os"
    "strconv"
    "strings"
)',
'func singleNumber(nums []int) int {
    return 0
}',
'func parseIntArray(line string) []int {
    line = strings.TrimSpace(line)
    line = strings.Trim(line, "[]")
    if line == "" {
        return []int{}
    }
    parts := strings.Split(line, ",")
    nums := make([]int, 0, len(parts))
    for _, p := range parts {
        n, _ := strconv.Atoi(strings.TrimSpace(p))
        nums = append(nums, n)
    }
    return nums
}

func main() {
    sc := bufio.NewScanner(os.Stdin)
    sc.Scan()
    nums := parseIntArray(sc.Text())
    fmt.Println(singleNumber(nums))
}', 1),

(10, 'cpp',
'#include <iostream>
#include <vector>
#include <sstream>
using namespace std;',
'class Solution {
public:
    int singleNumber(vector<int>& nums) {
        return 0;
    }
};',
'int main() {
    string line;
    getline(cin, line);

    if (line.size() >= 2) line = line.substr(1, line.size() - 2);
    else line = "";

    vector<int> nums;
    if (!line.empty()) {
        stringstream ss(line);
        string s;
        while (getline(ss, s, '','')) nums.push_back(stoi(s));
    }

    Solution sol;
    cout << sol.singleNumber(nums) << endl;
    return 0;
}', 1),

-- =========================
-- 11. 多数元素
-- =========================
(11, 'go',
'package main

import (
    "bufio"
    "fmt"
    "os"
    "strconv"
    "strings"
)',
'func majorityElement(nums []int) int {
    return 0
}',
'func parseIntArray(line string) []int {
    line = strings.TrimSpace(line)
    line = strings.Trim(line, "[]")
    if line == "" {
        return []int{}
    }
    parts := strings.Split(line, ",")
    nums := make([]int, 0, len(parts))
    for _, p := range parts {
        n, _ := strconv.Atoi(strings.TrimSpace(p))
        nums = append(nums, n)
    }
    return nums
}

func main() {
    sc := bufio.NewScanner(os.Stdin)
    sc.Scan()
    nums := parseIntArray(sc.Text())
    fmt.Println(majorityElement(nums))
}', 1),

(11, 'cpp',
'#include <iostream>
#include <vector>
#include <sstream>
using namespace std;',
'class Solution {
public:
    int majorityElement(vector<int>& nums) {
        return 0;
    }
};',
'int main() {
    string line;
    getline(cin, line);

    if (line.size() >= 2) line = line.substr(1, line.size() - 2);
    else line = "";

    vector<int> nums;
    if (!line.empty()) {
        stringstream ss(line);
        string s;
        while (getline(ss, s, '','')) nums.push_back(stoi(s));
    }

    Solution sol;
    cout << sol.majorityElement(nums) << endl;
    return 0;
}', 1),

-- =========================
-- 12. 存在重复元素
-- =========================
(12, 'go',
'package main

import (
    "bufio"
    "fmt"
    "os"
    "strconv"
    "strings"
)',
'func containsDuplicate(nums []int) bool {
    return false
}',
'func parseIntArray(line string) []int {
    line = strings.TrimSpace(line)
    line = strings.Trim(line, "[]")
    if line == "" {
        return []int{}
    }
    parts := strings.Split(line, ",")
    nums := make([]int, 0, len(parts))
    for _, p := range parts {
        n, _ := strconv.Atoi(strings.TrimSpace(p))
        nums = append(nums, n)
    }
    return nums
}

func main() {
    sc := bufio.NewScanner(os.Stdin)
    sc.Scan()
    nums := parseIntArray(sc.Text())
    if containsDuplicate(nums) {
        fmt.Println("true")
    } else {
        fmt.Println("false")
    }
}', 1),

(12, 'cpp',
'#include <iostream>
#include <vector>
#include <sstream>
using namespace std;',
'class Solution {
public:
    bool containsDuplicate(vector<int>& nums) {
        return false;
    }
};',
'int main() {
    string line;
    getline(cin, line);

    if (line.size() >= 2) line = line.substr(1, line.size() - 2);
    else line = "";

    vector<int> nums;
    if (!line.empty()) {
        stringstream ss(line);
        string s;
        while (getline(ss, s, '','')) nums.push_back(stoi(s));
    }

    Solution sol;
    cout << (sol.containsDuplicate(nums) ? "true" : "false") << endl;
    return 0;
}', 1),

-- =========================
-- 13. 丢失的数字
-- =========================
(13, 'go',
'package main

import (
    "bufio"
    "fmt"
    "os"
    "strconv"
    "strings"
)',
'func missingNumber(nums []int) int {
    return 0
}',
'func parseIntArray(line string) []int {
    line = strings.TrimSpace(line)
    line = strings.Trim(line, "[]")
    if line == "" {
        return []int{}
    }
    parts := strings.Split(line, ",")
    nums := make([]int, 0, len(parts))
    for _, p := range parts {
        n, _ := strconv.Atoi(strings.TrimSpace(p))
        nums = append(nums, n)
    }
    return nums
}

func main() {
    sc := bufio.NewScanner(os.Stdin)
    sc.Scan()
    nums := parseIntArray(sc.Text())
    fmt.Println(missingNumber(nums))
}', 1),

(13, 'cpp',
'#include <iostream>
#include <vector>
#include <sstream>
using namespace std;',
'class Solution {
public:
    int missingNumber(vector<int>& nums) {
        return 0;
    }
};',
'int main() {
    string line;
    getline(cin, line);

    if (line.size() >= 2) line = line.substr(1, line.size() - 2);
    else line = "";

    vector<int> nums;
    if (!line.empty()) {
        stringstream ss(line);
        string s;
        while (getline(ss, s, '','')) nums.push_back(stoi(s));
    }

    Solution sol;
    cout << sol.missingNumber(nums) << endl;
    return 0;
}', 1),

-- =========================
-- 14. 移动零
-- =========================
(14, 'go',
'package main

import (
    "bufio"
    "fmt"
    "os"
    "strconv"
    "strings"
)',
'func moveZeroes(nums []int) {
}',
'func parseIntArray(line string) []int {
    line = strings.TrimSpace(line)
    line = strings.Trim(line, "[]")
    if line == "" {
        return []int{}
    }
    parts := strings.Split(line, ",")
    nums := make([]int, 0, len(parts))
    for _, p := range parts {
        n, _ := strconv.Atoi(strings.TrimSpace(p))
        nums = append(nums, n)
    }
    return nums
}

func formatIntArray(nums []int) string {
    if len(nums) == 0 {
        return "[]"
    }
    var b strings.Builder
    b.WriteByte(''['')
    for i, v := range nums {
        if i > 0 {
            b.WriteByte('','')
        }
        b.WriteString(strconv.Itoa(v))
    }
    b.WriteByte('']'')
    return b.String()
}

func main() {
    sc := bufio.NewScanner(os.Stdin)
    sc.Scan()
    nums := parseIntArray(sc.Text())
    moveZeroes(nums)
    fmt.Println(formatIntArray(nums))
}', 1),

(14, 'cpp',
'#include <iostream>
#include <vector>
#include <sstream>
#include <string>
using namespace std;',
'class Solution {
public:
    void moveZeroes(vector<int>& nums) {
    }
};',
'static string formatVector(const vector<int>& nums) {
    string out = "[";
    for (size_t i = 0; i < nums.size(); ++i) {
        if (i) out += ",";
        out += to_string(nums[i]);
    }
    out += "]";
    return out;
}

int main() {
    string line;
    getline(cin, line);

    if (line.size() >= 2) line = line.substr(1, line.size() - 2);
    else line = "";

    vector<int> nums;
    if (!line.empty()) {
        stringstream ss(line);
        string s;
        while (getline(ss, s, '','')) nums.push_back(stoi(s));
    }

    Solution sol;
    sol.moveZeroes(nums);
    cout << formatVector(nums) << endl;
    return 0;
}', 1),

-- =========================
-- 15. 找到所有数组中消失的数字
-- =========================
(15, 'go',
'package main

import (
    "bufio"
    "fmt"
    "os"
    "strconv"
    "strings"
)',
'func findDisappearedNumbers(nums []int) []int {
    return nil
}',
'func parseIntArray(line string) []int {
    line = strings.TrimSpace(line)
    line = strings.Trim(line, "[]")
    if line == "" {
        return []int{}
    }
    parts := strings.Split(line, ",")
    nums := make([]int, 0, len(parts))
    for _, p := range parts {
        n, _ := strconv.Atoi(strings.TrimSpace(p))
        nums = append(nums, n)
    }
    return nums
}

func formatIntArray(nums []int) string {
    if len(nums) == 0 {
        return "[]"
    }
    var b strings.Builder
    b.WriteByte(''['')
    for i, v := range nums {
        if i > 0 {
            b.WriteByte('','')
        }
        b.WriteString(strconv.Itoa(v))
    }
    b.WriteByte('']'')
    return b.String()
}

func main() {
    sc := bufio.NewScanner(os.Stdin)
    sc.Scan()
    nums := parseIntArray(sc.Text())
    fmt.Println(formatIntArray(findDisappearedNumbers(nums)))
}', 1),

(15, 'cpp',
'#include <iostream>
#include <vector>
#include <sstream>
#include <string>
using namespace std;',
'class Solution {
public:
    vector<int> findDisappearedNumbers(vector<int>& nums) {
        return {};
    }
};',
'static string formatVector(const vector<int>& nums) {
    string out = "[";
    for (size_t i = 0; i < nums.size(); ++i) {
        if (i) out += ",";
        out += to_string(nums[i]);
    }
    out += "]";
    return out;
}

int main() {
    string line;
    getline(cin, line);

    if (line.size() >= 2) line = line.substr(1, line.size() - 2);
    else line = "";

    vector<int> nums;
    if (!line.empty()) {
        stringstream ss(line);
        string s;
        while (getline(ss, s, '','')) nums.push_back(stoi(s));
    }

    Solution sol;
    cout << formatVector(sol.findDisappearedNumbers(nums)) << endl;
    return 0;
}', 1),

-- =========================
-- 16. 最大连续1的个数
-- =========================
(16, 'go',
'package main

import (
    "bufio"
    "fmt"
    "os"
    "strconv"
    "strings"
)',
'func findMaxConsecutiveOnes(nums []int) int {
    return 0
}',
'func parseIntArray(line string) []int {
    line = strings.TrimSpace(line)
    line = strings.Trim(line, "[]")
    if line == "" {
        return []int{}
    }
    parts := strings.Split(line, ",")
    nums := make([]int, 0, len(parts))
    for _, p := range parts {
        n, _ := strconv.Atoi(strings.TrimSpace(p))
        nums = append(nums, n)
    }
    return nums
}

func main() {
    sc := bufio.NewScanner(os.Stdin)
    sc.Scan()
    nums := parseIntArray(sc.Text())
    fmt.Println(findMaxConsecutiveOnes(nums))
}', 1),

(16, 'cpp',
'#include <iostream>
#include <vector>
#include <sstream>
using namespace std;',
'class Solution {
public:
    int findMaxConsecutiveOnes(vector<int>& nums) {
        return 0;
    }
};',
'int main() {
    string line;
    getline(cin, line);

    if (line.size() >= 2) line = line.substr(1, line.size() - 2);
    else line = "";

    vector<int> nums;
    if (!line.empty()) {
        stringstream ss(line);
        string s;
        while (getline(ss, s, '','')) nums.push_back(stoi(s));
    }

    Solution sol;
    cout << sol.findMaxConsecutiveOnes(nums) << endl;
    return 0;
}', 1),

-- =========================
-- 17. 数组拆分 I
-- =========================
(17, 'go',
'package main

import (
    "bufio"
    "fmt"
    "os"
    "strconv"
    "strings"
)',
'func arrayPairSum(nums []int) int {
    return 0
}',
'func parseIntArray(line string) []int {
    line = strings.TrimSpace(line)
    line = strings.Trim(line, "[]")
    if line == "" {
        return []int{}
    }
    parts := strings.Split(line, ",")
    nums := make([]int, 0, len(parts))
    for _, p := range parts {
        n, _ := strconv.Atoi(strings.TrimSpace(p))
        nums = append(nums, n)
    }
    return nums
}

func main() {
    sc := bufio.NewScanner(os.Stdin)
    sc.Scan()
    nums := parseIntArray(sc.Text())
    fmt.Println(arrayPairSum(nums))
}', 1),

(17, 'cpp',
'#include <iostream>
#include <vector>
#include <sstream>
#include <algorithm>
using namespace std;',
'class Solution {
public:
    int arrayPairSum(vector<int>& nums) {
        return 0;
    }
};',
'int main() {
    string line;
    getline(cin, line);

    if (line.size() >= 2) line = line.substr(1, line.size() - 2);
    else line = "";

    vector<int> nums;
    if (!line.empty()) {
        stringstream ss(line);
        string s;
        while (getline(ss, s, '','')) nums.push_back(stoi(s));
    }

    Solution sol;
    cout << sol.arrayPairSum(nums) << endl;
    return 0;
}', 1);

COMMIT;

SET FOREIGN_KEY_CHECKS = 1;