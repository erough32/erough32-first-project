package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	cryptoRand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

var pushCommJson map[string]string

func init() {
	var pushCommInfo []ConfigDb
	err := Db.Select(&pushCommInfo, "SELECT * FROM `config` WHERE `key` = 'tgbot'")
	if err != nil {
		fmt.Println("SqlError", err)
	}

	json.Unmarshal(pushCommInfo[0].Config, &pushCommJson)
}

func SQLInject(info string) bool {
	ok, _ := regexp.MatchString("\\`"+`|\"|\'|\#|select|SELECT|delete|DELETE|insert|INSERT|union|UNION`, info)
	return ok
}

func GetSHA256HashCode(message []byte) string {
	hash := sha256.New()
	hash.Write(message)
	bytes := hash.Sum(nil)
	hashCode := hex.EncodeToString(bytes)
	return hashCode
}

func GetMD5HashCode(message []byte) string {
	hash := md5.New()
	hash.Write(message)
	bytes := hash.Sum(nil)
	hashCode := hex.EncodeToString(bytes)
	return hashCode
}

func ReadAll(filePth string) []byte {
	f, err := os.Open(filePth)
	if err != nil {
		return nil
	}

	file, _ := ioutil.ReadAll(f)
	return file
}

func generateRandomNumber(start int, end int, count int) []int {
	if end < start || (end-start) < count {
		return nil
	}

	nums := make([]int, 0)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for len(nums) < count {
		num := r.Intn((end - start)) + start

		exist := false
		for _, v := range nums {
			if v == num {
				exist = true
				break
			}
		}

		if !exist {
			nums = append(nums, num)
		}
	}

	return nums
}
func SliceRemoveDuplicates(slice []int) []int {
	i := 0
	var j int
	for {
		if i >= len(slice)-1 {
			break
		}
		for j = i + 1; j < len(slice) && slice[i] == slice[j]; j++ {
		}
		slice = append(slice[:i+1], slice[j:]...)
		i++
	}
	return slice
}

func SliceRemoveDuplicatesString(slice []string) []string {
	i := 0
	var j int
	for {
		if i >= len(slice)-1 {
			break
		}
		for j = i + 1; j < len(slice) && slice[i] == slice[j]; j++ {
		}
		slice = append(slice[:i+1], slice[j:]...)
		i++
	}
	return slice
}

func intInSlice(a int, list []int) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func delIntSlice(a []int, d int) []int {
	for i := 0; i < len(a); i++ {
		if a[i] == d {
			a = append(a[:i], a[i+1:]...)
			return a
		}
	}
	return a
}

func VerifyEmailFormat(email string) bool {
	pattern := `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}

// func delRepeatElem(arr []string) (newArr []string) {
// 	newArr = make([]string, 0)
// 	for i := 0; i < len(arr); i++ {
// 		repeat := false
// 		for j := i + 1; j < len(arr); j++ {
// 			if arr[i] == arr[j] {
// 				repeat = true
// 				break
// 			}
// 		}
// 		if !repeat {
// 			newArr = append(newArr, arr[i])
// 		}
// 	}
// 	return
// }

func md5V(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func getIpArea(ip string) string {
	resp, err := http.Get("https://global-api.noyteam.online/api/geoip?ip=" + ip)
	if err != nil {
		fmt.Println("GetError", err)
		return "ERROR"
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode == 200 {
		var jsonInfo map[string]string
		json.Unmarshal(body, &jsonInfo)
		if jsonInfo["country_name"] != "" {
			return jsonInfo["country_name"]
		} else {
			return "ERROR"
		}
	} else {
		return "ERROR"
	}
}

func isSensitive(text string) bool {
	for i := range commConfig {
		if strings.Contains(text, commConfig[i]) {
			return true
		}
	}
	return false
}

func pushCommToGroup(msg, chat string) {
	var chatId string
	if chat != "" {
		chatId = chat
	} else {
		chatId = pushCommJson["group"]
	}
	resp, err := http.Get("https://api.telegram.org/bot" + pushCommJson["token"] + "/sendMessage?chat_id=" + chatId + "&text=" + url.QueryEscape(msg))
	if err != nil {
		fmt.Println("GetError", err)
		return
	}
	defer resp.Body.Close()
}

func operator3(c bool, v1, v2 int) int {
	if c {
		return v1
	} else {
		return v2
	}
}

func httpPost(url, postinfo string) string {
	resp, err := http.Post(url,
		"application/x-www-form-urlencoded", strings.NewReader(postinfo))
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	return string(body)
}

func aesEncrypt(key []byte, text string) string {
	plaintext := []byte(text)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(cryptoRand.Reader, iv); err != nil {
		panic(err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return base64.URLEncoding.EncodeToString(ciphertext)
}
