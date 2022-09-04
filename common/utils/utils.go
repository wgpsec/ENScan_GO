package utils

import (
	"bufio"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/wgpsec/ENScan/common/utils/gologger"
	"math"
	"math/big"
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

//CheckList 检查列表发现空返回false
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

func ReadFile(filename string) []string {
	var result []string
	if FileExists(filename) {
		f, err := os.Open(filename)
		if err != nil {
			gologger.Fatalf("read fail", err)
		}
		fileScanner := bufio.NewScanner(f)
		// read line by line
		for fileScanner.Scan() {
			result = append(result, fileScanner.Text())
		}
		// handle first encountered error while reading
		if err := fileScanner.Err(); err != nil {
			gologger.Fatalf("Error while reading file: %s\n", err)
		}
		_ = f.Close()
	}
	result = SetStr(result)
	gologger.Infof("读取到 %d 条信息（已去重）\n", len(result))
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
			gologger.Errorf("无法查询法人信息\n")
			res = ""
		}
		res = "xlb"
	} else {
		gologger.Errorf("pid长度%d不正确，pid: %s", len(pid), pid)
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
		gologger.Errorf("转换失败：%s\n", err)
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
