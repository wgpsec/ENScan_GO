package utils

import (
	"bufio"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/tidwall/gjson"
	"github.com/wgpsec/ENScan/common/gologger"
	"math"
	"math/big"
	mrand "math/rand"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Md5 MD5加密
// src 源字符
func Md5(src string) string {
	m := md5.New()
	m.Write([]byte(src))
	res := hex.EncodeToString(m.Sum(nil))
	return res
}

// SetStr 数据去重
// target 输入数据
func SetStr(target []string) []string {
	setMap := make(map[string]int)
	var result []string
	for _, v := range target {
		if v != "" {
			if _, ok := setMap[v]; !ok {
				setMap[v] = 0
				result = append(result, v)
			}
		}
	}
	return result
}

// CheckList 检查列表发现空返回false
func CheckList(target []string) bool {
	if len(target) == 0 {
		return false
	}
	for _, v := range target {
		if v == "" {
			return false
		}
	}
	return true
}

// RangeRand 生成区间[-m, n]的安全随机数
func RangeRand(min, max int64) int64 {
	if min > max {
		panic("the min is greater than max!")
	}

	if min < 0 {
		f64Min := math.Abs(float64(min))
		i64Min := int64(f64Min)
		result, _ := rand.Int(rand.Reader, big.NewInt(max+1+i64Min))

		return result.Int64() - i64Min
	} else {
		result, _ := rand.Int(rand.Reader, big.NewInt(max-min+1))
		return min + result.Int64()
	}
}

func IsInList(target string, list []string) bool {
	if len(list) == 0 {
		return false
	}
	for _, v := range list {
		if v == target {
			return true
		}
	}
	return false
}

func DelInList(target string, list []string) []string {
	var result []string
	for _, v := range list {
		if v != target {
			result = append(result, v)
		}
	}
	return result
}
func DelListInList(original []string, remove []string) []string {
	removeMap := make(map[string]bool)
	for _, v := range remove {
		removeMap[v] = true
	}

	n := 0
	for _, v := range original {
		if !removeMap[v] {
			original[n] = v
			n++
		}
	}
	return original[:n]
}

func ReadFileOutLine(filename string) []string {
	var result []string
	if FileExists(filename) {
		f, err := os.Open(filename)
		if err != nil {
			gologger.Fatal().Msgf("read fail", err)
		}
		fileScanner := bufio.NewScanner(f)
		// read line by line
		for fileScanner.Scan() {
			result = append(result, fileScanner.Text())
		}
		// handle first encountered error while reading
		if err := fileScanner.Err(); err != nil {
			gologger.Fatal().Msgf("Error while reading file: %s\n", err)
		}
		_ = f.Close()
	}
	result = SetStr(result)
	return result
}

func GetConfigPath() string { // 获得配置文件的绝对路径
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return dir
}
func DName(str string) (srt string) { // 获得文件名
	str = strings.ReplaceAll(str, "(", "（")
	str = strings.ReplaceAll(str, ")", "）")
	str = strings.ReplaceAll(str, "<em>", "")
	str = strings.ReplaceAll(str, "</em>", "")
	return str
}

// CheckPid 检查pid是哪家单位
func CheckPid(pid string) (res string) {
	if len(pid) == 32 {
		res = "qcc"
	} else if len(pid) == 14 {
		res = "aqc"
	} else if len(pid) == 8 || len(pid) == 7 || len(pid) == 6 || len(pid) == 9 || len(pid) == 10 {
		res = "tyc"
	} else if len(pid) == 33 || len(pid) == 34 {
		if pid[0] == 'p' {
			gologger.Error().Msgf("无法查询法人信息\n")
			res = ""
		}
		res = "xlb"
	} else {
		gologger.Error().Msgf("pid长度%d不正确，pid: %s\n", len(pid), pid)
		return ""
	}
	return res
}

func FormatInvest(scale string) float64 {
	if scale == "-" || scale == "" || scale == " " {
		return -1
	} else {
		scale = strings.ReplaceAll(scale, "%", "")
	}

	num, err := strconv.ParseFloat(scale, 64)
	if err != nil {
		gologger.Error().Msgf("转换失败：%s\n", err)
		return -1
	}
	return num
}

func WriteFile(str string, path string) {
	//os.O_WRONLY | os.O_CREATE
	file, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("file open error:", err)
		return
	}
	defer file.Close()

	//使用缓存方式写入
	writer := bufio.NewWriter(file)

	count, w_err := writer.WriteString(str)

	//将缓存中数据刷新到文本中
	writer.Flush()

	if w_err != nil {
		fmt.Println("写入出错")
	} else {
		fmt.Printf("写入成功,共写入字节：%v", count)
	}
}

func VerifyEmailFormat(email string) bool {
	pattern := `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*` //匹配电子邮箱
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}

// TBS 展示表格
func TBS(h []string, ep []string, name string, data []gjson.Result) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(h)
	for _, v := range data {
		var tmp []string
		for _, r := range gjson.GetMany(v.String(), ep...) {
			rs := r.String()
			if len([]rune(rs)) > 30 {
				rs = string([]rune(rs)[:30])
			}
			tmp = append(tmp, rs)
		}
		table.Append(tmp)
	}
	gologger.Info().Msgf(name)
	table.Render()
}

// MergeMap  合并map
// s: 源map，list: 目标map
func MergeMap(s map[string][]map[string]string, list map[string][]map[string]string) {
	for k, v := range s {
		if l, ok := list[k]; ok {
			list[k] = append(l, v...)
		} else {
			list[k] = v
		}
	}
}

func fd(arr []string, target string) int {
	// 查找目标字符串在切片中的首次出现位置
	for i, word := range arr {
		if word == target {
			return i
		}
	}
	return -1
}

// RandomElement 从一个字符串切片中随机返回一个元素。
func RandomElement(c string) string {
	slice := strings.Split(c, "|")
	// 确保切片不为空，避免 panic
	if len(slice) == 0 {
		return ""
	}
	if len(slice) == 1 {
		return slice[0]
	}
	// 生成一个在 [0, len(slice)) 范围内的随机索引
	index := mrand.Intn(len(slice))
	// 返回该索引对应的元素
	return slice[index]
}

func ExtractPortString(rawURL string) (string, error) {
	// 解析 URL
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("无效的 URL: %v", err)
	}
	// 提取 host:port 部分
	_, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		return "", fmt.Errorf("未指定端口且协议 %q 无默认端口", u.Scheme)
	}

	return port, nil // port 已经是字符串
}

// DeduplicateMapList 对map列表进行去重
// dataType: 数据类型，用于确定去重的关键字段
// data: 待去重的数据列表
func DeduplicateMapList(dataType string, data []map[string]string) []map[string]string {
	if len(data) == 0 {
		return data
	}

	originalCount := len(data)

	// 定义不同数据类型的唯一标识字段组合
	var keyFields []string
	switch dataType {
	case "enterprise_info", "invest", "branch", "holds", "supplier", "partner":
		// 企业相关数据使用 PID 作为唯一标识
		keyFields = []string{"pid"}
	case "icp":
		// ICP备案使用域名+备案号作为唯一标识
		keyFields = []string{"domain", "icp"}
	case "app":
		// APP使用名称+Bundle ID作为唯一标识
		keyFields = []string{"name", "bundle_id"}
	case "wx_app", "wechat":
		// 微信小程序和公众号使用名称作为唯一标识
		keyFields = []string{"name"}
	case "weibo":
		// 微博使用链接作为唯一标识
		keyFields = []string{"profile_url"}
	case "job":
		// 招聘信息使用职位名称+发布日期+地点作为唯一标识
		keyFields = []string{"name", "publish_time", "location"}
	case "copyright":
		// 软件著作权使用登记号作为唯一标识
		keyFields = []string{"reg_num"}
	default:
		// 默认使用名称作为唯一标识
		keyFields = []string{"name"}
	}

	// 使用map记录已经出现过的记录
	seen := make(map[string]bool)
	result := make([]map[string]string, 0, len(data))

	for _, item := range data {
		// 生成唯一键
		var keyParts []string
		for _, field := range keyFields {
			keyParts = append(keyParts, item[field])
		}
		key := strings.Join(keyParts, "|")

		// 如果key为空（所有字段都为空）或者已经见过，跳过
		if key == strings.Repeat("|", len(keyFields)-1) {
			continue
		}

		if !seen[key] {
			seen[key] = true
			result = append(result, item)
		}
	}

	// 记录去重统计信息
	duplicateCount := originalCount - len(result)
	if duplicateCount > 0 {
		gologger.Debug().Msgf("⌈%s⌋ 去重: 原始 %d 条，去重后 %d 条，移除重复 %d 条", dataType, originalCount, len(result), duplicateCount)
	}

	return result
}

// DeduplicateData 对整个数据集进行去重
func DeduplicateData(data map[string][]map[string]string) map[string][]map[string]string {
	result := make(map[string][]map[string]string)
	totalOriginal := 0
	totalDeduplicated := 0

	for dataType, items := range data {
		originalCount := len(items)
		totalOriginal += originalCount
		result[dataType] = DeduplicateMapList(dataType, items)
		totalDeduplicated += len(result[dataType])
	}

	totalRemoved := totalOriginal - totalDeduplicated
	if totalRemoved > 0 {
		gologger.Info().Msgf("数据去重完成: 原始 %d 条，去重后 %d 条，移除重复 %d 条", totalOriginal, totalDeduplicated, totalRemoved)
	}

	return result
}
