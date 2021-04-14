package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/alecthomas/kong"
	"github.com/sirupsen/logrus"
	asc2art "github.com/yinghau76/go-ascii-art"
	"image"
	"io/ioutil"
	"os"
	"strings"
)

var bot *client.QQClient

var config struct {
	QQ         int64  `required:"" help:"qq number"`
	Password   string `required:"" help:"qq password"`
	DevicePath string `type:"path" help:"device.json path, default './device.json', create new one if not found"`
}

func Login() error {
	if bot != nil {
		return errors.New("bot is already login")
	}

	kong.Parse(&config)

	var devicePath = config.DevicePath

	if devicePath == "" {
		devicePath = "device.json"
	}

	dj, err := ioutil.ReadFile(devicePath)
	if err != nil {
		if os.IsNotExist(err) {
			if config.DevicePath != "" {
				logrus.WithField("device-path", config.DevicePath).Errorf("device.json path not exist")
				return errors.New("device path not exist")
			}
			client.GenRandomDevice()
			ioutil.WriteFile("device.json", client.SystemDeviceInfo.ToJson(), os.FileMode(0755))
		} else {
			return err
		}
	} else {
		if err := client.SystemDeviceInfo.ReadJson(dj); err != nil {
			logrus.WithField("device-path", config.DevicePath).Errorf("device.json error")
			return errors.New("device.json error")
		}
	}

	client.SystemDeviceInfo.Protocol = client.AndroidPhone

	bot = client.NewClient(config.QQ, config.Password)

	// see github.com/Logiase/MiraiGo-Template
	resp, err := bot.Login()
	console := bufio.NewReader(os.Stdin)
	for {
		if err != nil {
			logrus.WithError(err).Fatal("unable to login")
		}

		var text string
		if !resp.Success {
			switch resp.Error {
			case client.NeedCaptcha:
				img, _, _ := image.Decode(bytes.NewReader(resp.CaptchaImage))
				fmt.Println(asc2art.New("image", img).Art)
				fmt.Print("please input captcha: ")
				text, _ := console.ReadString('\n')
				resp, err = bot.SubmitCaptcha(strings.ReplaceAll(text, "\n", ""), resp.CaptchaSign)
				continue

			case client.UnsafeDeviceError:
				fmt.Printf("device lock -> %v\n", resp.VerifyUrl)
				os.Exit(4)

			case client.SMSNeededError:
				fmt.Println("device lock enabled, Need SMS Code")
				fmt.Printf("Send SMS to %s ? (yes)", resp.SMSPhone)
				t, _ := console.ReadString('\n')
				t = strings.TrimSpace(t)
				if t != "yes" {
					os.Exit(2)
				}
				if !bot.RequestSMS() {
					logrus.Warnf("unable to request SMS Code")
					os.Exit(2)
				}
				logrus.Warn("please input SMS Code: ")
				text, _ = console.ReadString('\n')
				resp, err = bot.SubmitSMS(strings.ReplaceAll(strings.ReplaceAll(text, "\n", ""), "\r", ""))
				continue

			case client.TooManySMSRequestError:
				fmt.Printf("too many SMS request, please try later.\n")
				os.Exit(6)

			case client.SMSOrVerifyNeededError:
				fmt.Println("device lock enabled, choose way to verify:")
				fmt.Println("1. Send SMS Code to ", resp.SMSPhone)
				fmt.Println("2. Scan QR Code")
				fmt.Print("input (1,2):")
				text, _ = console.ReadString('\n')
				text = strings.TrimSpace(text)
				switch text {
				case "1":
					if !bot.RequestSMS() {
						logrus.Warnf("unable to request SMS Code")
						os.Exit(2)
					}
					logrus.Warn("please input SMS Code: ")
					text, _ = console.ReadString('\n')
					resp, err = bot.SubmitSMS(strings.ReplaceAll(strings.ReplaceAll(text, "\n", ""), "\r", ""))
					continue
				case "2":
					fmt.Printf("device lock -> %v\n", resp.VerifyUrl)
					os.Exit(2)
				default:
					fmt.Println("invalid input")
					os.Exit(2)
				}

			case client.SliderNeededError:
				if client.SystemDeviceInfo.Protocol == client.AndroidPhone {
					logrus.Warn("Android Phone Protocol DO NOT SUPPORT Slide verify")
					logrus.Warn("please use other protocol")
					os.Exit(2)
				}
				bot.AllowSlider = false
				bot.Disconnect()
				resp, err = bot.Login()
				continue

			case client.OtherLoginError, client.UnknownLoginError:
				logrus.Fatalf("login failed: %v", resp.ErrorMessage)
				os.Exit(3)
			}
		}
		break
	}
	if err := bot.ReloadFriendList(); err != nil {
		logrus.WithError(err).Error("ReloadFriendList failed")
		return err
	}
	if err := bot.ReloadGroupList(); err != nil {
		logrus.WithError(err).Error("ReloadGroupList failed")
		return err
	}
	return nil
}

func Shutdown() {
	if bot != nil {
		bot.Disconnect()
	}
}
