package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"easy-echo/cache"
	"easy-echo/config"
	"easy-echo/constants"
	"easy-echo/db"
	"easy-echo/logger"
	"easy-echo/response"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/labstack/echo"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func main() {
	validateResult:=validate("license.dat")
	if !validateResult {
		fmt.Println("license valid fail~~~~~~~~~~~")
		return
	}

	err := config.InitConfig(&config.Config{})
	if err != nil {
		log.Println("InitConfig failure, err:", err.Error())
		logger.WriteLog("initialize", time.Now().Format(constants.LogTimeFormat),
			fmt.Sprintf("InitConfig failure, err: %s", err.Error()))
		return
	}

	err = cache.InitRedis()
	if err != nil {
		log.Println("InitRedis failure, err:", err.Error())
		logger.WriteLog("initialize", time.Now().Format(constants.LogTimeFormat),
			fmt.Sprintf("InitRedis failure, err: %s", err.Error()))
		return
	}

	err = db.InitMongoDB()
	if err != nil {
		log.Println("InitMongoDB failure, err:", err.Error())
		logger.WriteLog("initialize", time.Now().Format(constants.LogTimeFormat),
			fmt.Sprintf("InitMongoDB failure, err: %s", err.Error()))
		return
	}

	e := echo.New()
	e.GET("/hello", func(c echo.Context) error {
		request := c.Request()

		fmt.Println("x-tif-uid: ", request.Header.Get("x-tif-uid"))
		fmt.Println("x-tif-uinfo: ", request.Header.Get("x-tif-uinfo"))
		fmt.Println("x-tif-ext: ", request.Header.Get("x-tif-ext"))



		fmt.Println("id 2", request.FormValue("id"))

		logger.Info("abcd")
		return response.ShowSuccess(c, "success", "Hello, World!")
	})

	e.Logger.Fatal(e.Start(config.Cfg.Host))
}
func runInLinuxWithErr(cmd string) (string, error) {
	fmt.Println("Running Linux cmd:"+cmd)
	result, err := exec.Command("/bin/sh", "-c", cmd).Output()
	if err != nil {
		fmt.Println(err.Error())
	}
	return strings.TrimSpace(string(result)), err
}

func runInWindowsWithErr(cmd string) (string, error){
	fmt.Println("Running Windows cmd:"+cmd)
	result, err := exec.Command("cmd", "/c", cmd).Output()
	if err != nil {
		fmt.Println(err.Error())
	}
	return strings.TrimSpace(string(result)), err
}

func RunCommandWithErr(cmd string) (string, error){
	if runtime.GOOS == "windows" {
		return runInWindowsWithErr(cmd)
	} else {
		return runInLinuxWithErr(cmd)
	}
}

func createLicenseFile(sn string,expire string,fileName string) {
	message :=sn+";"+expire

	publicKey, err := ioutil.ReadFile("public.pem")
	if err != nil {
		fmt.Println(err.Error())
	}
	cipher, err := RsaEncrypt([]byte(message), publicKey)
	b64:=base64.StdEncoding.EncodeToString(cipher)
	if err != nil {
		log.Fatalf("Cannot encrypt message\n")
	}
	f, err := os.Create(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = f.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	l, err := f.WriteString(b64)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(l, " bytes written successfully!")

}
func validate(licenseFile string) bool {
	keybuffer, err := ioutil.ReadFile(licenseFile)
	if err != nil {
		fmt.Println(err.Error())
	}
	cipher,err := base64.StdEncoding.DecodeString(string(keybuffer))


	privateKey, err := ioutil.ReadFile("private.pem")
	if err != nil {
		fmt.Println(err.Error())
	}

	plain, err := RsaDecrypt(cipher,privateKey)

	if err != nil {
		log.Fatalf("Cannot decrypt message\n")
	}
	plainText:=string(plain)
	arr:=strings.Split(plainText,";")

	if len(arr)!=2 {
		log.Fatal("licenseFile bad")
		return false
	} else {
		sn :=arr[0]
		expire, err :=strconv.ParseInt(arr[1], 10, 64)
		var cmd string
		if runtime.GOOS == "windows" {
			cmd="wmic cpu get processorid"
		} else {
			cmd="dmidecode | grep 'Serial Number' | awk -F ':' '{print $2}' | head -n 1"
		}
		serverSn, err := RunCommandWithErr(cmd)
		if err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println("execute ok :"+ serverSn)
		}
		if serverSn==sn && time.Now().Unix()<int64(expire) {
			return true
		}
	}
	return false
}



// 加密
func RsaEncrypt(origData []byte,publicKey []byte) ([]byte, error) {
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return nil, errors.New("public key error")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pub := pubInterface.(*rsa.PublicKey)
	return rsa.EncryptPKCS1v15(rand.Reader, pub, origData)
}

// 解密
func RsaDecrypt(ciphertext []byte,privateKey []byte) ([]byte, error) {
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil, errors.New("private key error!")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptPKCS1v15(rand.Reader, priv, ciphertext)
}



func DumpPrivateKeyFile(privatekey *rsa.PrivateKey, filename string) error {
	var keybytes []byte = x509.MarshalPKCS1PrivateKey(privatekey)
	block := &pem.Block{
		Type  : "RSA PRIVATE KEY",
		Bytes :  keybytes,
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	err = pem.Encode(file, block)
	if err != nil {
		return err
	}
	return nil
}

func DumpPublicKeyFile(publickey *rsa.PublicKey, filename string) error {
	keybytes, err := x509.MarshalPKIXPublicKey(publickey)
	if err != nil {
		return err
	}
	block := &pem.Block{
		Type  : "PUBLIC KEY",
		Bytes :  keybytes,
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	err = pem.Encode(file, block)
	if err != nil {
		return err
	}
	return nil
}

func GenerateKey() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privatekey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	publickey := &privatekey.PublicKey
	return privatekey, publickey, nil
}
func genPairFile() {
	privatekey, publickey, err := GenerateKey()
	if err != nil {
		log.Fatalf("Cannot generate RSA key\n")
	}

	// dump private key to file
	err = DumpPrivateKeyFile(privatekey, "private.pem")
	if err != nil {
		log.Fatalf("Cannot dump private key file\n")
	}
	// dump public key to file
	err = DumpPublicKeyFile(publickey, "public.pem")
	if err != nil {
		log.Fatalf("Cannot dump public key file\n")
	}
}