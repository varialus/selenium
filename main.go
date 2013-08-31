/*
Copyright (c) 2013, Aulus Egnatius Varialus <varialus@gmail.com>

Permission to use, copy, modify, and/or distribute this software for any purpose with or without fee is hereby granted, provided that the above copyright notice and this permission notice appear in all copies.

THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
*/

// http://code.google.com/p/chromedriver/wiki/CapabilitiesAndSwitches
// http://code.google.com/p/chromedriver/wiki/GettingStarted
// http://code.google.com/p/chromedriver/wiki/TroubleshootingAndSupport
// http://code.google.com/p/selenium/source/browse/java/server/src/org/openqa/grid/common/defaults/DefaultNode.json
// http://code.google.com/p/selenium/source/browse/py/
// http://code.google.com/p/selenium/source/browse/py/selenium/webdriver/__init__.py
// http://code.google.com/p/selenium/source/browse/py/selenium/webdriver/chrome/__init__.py
// http://code.google.com/p/selenium/source/browse/py/selenium/webdriver/chrome/options.py
// http://code.google.com/p/selenium/source/browse/py/selenium/webdriver/chrome/service.py
// http://code.google.com/p/selenium/source/browse/py/selenium/webdriver/chrome/webdriver.py
// http://code.google.com/p/selenium/source/browse/py/selenium/webdriver/common/desired_capabilities.py
// http://code.google.com/p/selenium/source/browse/py/selenium/webdriver/common/utils.py
// http://code.google.com/p/selenium/source/browse/py/selenium/webdriver/remote/command.py
// http://code.google.com/p/selenium/source/browse/py/selenium/webdriver/remote/remote_connection.py
// http://code.google.com/p/selenium/source/browse/py/selenium/webdriver/remote/webdriver.py
// http://code.google.com/p/selenium/source/browse/py/README
// http://code.google.com/p/selenium/w/list
// http://code.google.com/p/selenium/wiki/ArchitecturalOverview
// http://code.google.com/p/selenium/wiki/ChromeDriver
// http://code.google.com/p/selenium/wiki/ChromeDriver#Overriding_the_Chrome_binary_location
// http://code.google.com/p/selenium/wiki/DefiningNewWireProtocolCommands
// http://code.google.com/p/selenium/wiki/DesiredCapabilities
// http://code.google.com/p/selenium/wiki/JsonWireProtocol
// http://code.google.com/p/selenium/wiki/RemoteWebDriver
// http://selenium.googlecode.com/git/docs/api/py/webdriver_remote/selenium.webdriver.remote.webdriver.html

// ChromeDriver Default Address 127.0.0.1:9515

// TODO: Recommendation: <mortdeus> varialus, fyi if you call your file main_linux_amd64.go, you can take out the runtime checks.

package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// http://code.google.com/p/selenium/wiki/JsonWireProtocol
//
// /session
//
//	POST /session
//
//		Create a new session.
//		The server should attempt to create a session that most closely matches the desired and required capabilities.
//		Required capabilities have higher priority than desired capabilities and must be set for the session to be created.
//
//		JSON Parameters:
//			desiredCapabilities - {object} An object describing the session's desired capabilities.
//			requiredCapabilities - {object} An object describing the session's required capabilities (Optional).
//
//		Returns:
//			{object} An object describing the session's capabilities.
//
//		Potential Errors:
//			SessionNotCreatedException - If a required capability could not be set.
func Session(url string, capabilities interface{}, v interface{}) error {
	if json_bytes, err := json.Marshal(capabilities); err != nil {
		return fmt.Errorf("Error: problem while calling json.Marshal(%v); err == %s", capabilities, err.Error())
	} else {
		fmt.Println("json_bytes ==", json_bytes)
		buffered_data := bytes.NewBuffer(json_bytes)
		if response, err := http.Post(url + "/session", "application/json", buffered_data); err != nil {
			return fmt.Errorf("Error: problem while calling http.Post(%s, \"application/json\", json.Marshal(%v)); err == %s", url + "/session", capabilities, err.Error())
		} else {
			fmt.Println("response ==", response)
			if bytes, err := ioutil.ReadAll(response.Body); err != nil {
				return fmt.Errorf("Error: problem getting response while starting new session; %s", err.Error())
			} else {
				defer response.Body.Close()
				fmt.Println("string(bytes) ==", string(bytes))
				return json.Unmarshal(bytes, v)
			}
		}
	}
}

func main() {
	var browser_name string
	if runtime.GOOS != "linux" {
		fmt.Println(fmt.Errorf("Error: selenium not yet implemented on %s", runtime.GOOS))
		return
	}
	if runtime.GOARCH != "amd64" {
		fmt.Println(fmt.Errorf("Error: selenium not yet implemented on %s", runtime.GOARCH))
		return
	}
	flag.StringVar(&browser_name, "browser", "chromium", "-browser=chromium|chrome|firefox|iceweasel|ie|opera")
	flag.Parse()
	if browser_name != "chromium" {
		fmt.Println(fmt.Errorf("Error: selenium not yet implemented on %s", browser_name))
		return
	}
	fmt.Println("browser ==", browser_name)
	if port, err := free_port(); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("port ==", port)
		go func() {
			if err := startChromeDriver(true, port); err != nil {
				go func() {
					startChromeDriver(false, port)
				}()
			}
		}()
		defer stopChromeDriver(port)
		waitForChromeDriver(port)
		chrome_options := map[string]interface{}{
			// function returns slice of extension *.crx files read as Python base64 data
			// it gets called and formatted before it is sent to chromedriver
			//"extensions":		func()
			"binary": "/usr/bin/chromium",
		}
		desired_chrome_capabilities := map[string]interface{}{
			"browserName":       "chrome",
			"version":           "",
			"platform":          "ANY",
			"javascriptEnabled": true,
			"chromeOptions":     chrome_options,
		}
		fmt.Println("desired_chrome_capabilities ==", desired_chrome_capabilities)
		//command_executor_url := "http://127.0.0.1:4444/wd/hub"
		//command_executor_url := fmt.Sprintf("http://127.0.0.1:%d/wd/hub", port)
		command_executor_url := fmt.Sprintf("http://127.0.0.1:%d", port)
		fmt.Println("command_executor_url", command_executor_url)
		commands := map[string][2]string{
			"newSession": [2]string{"POST", "/session"},
		}
		fmt.Println("commands ==", commands)
		driver_command_to_execute := "newSession"
		fmt.Println("driver_command_to_execute ==", driver_command_to_execute)
		params_to_execute := map[string]map[string]interface{}{
			"desiredCapabilities": desired_chrome_capabilities,
		}
		fmt.Println("params_to_execute ==", params_to_execute)
		var result interface{}
		if err := Session(command_executor_url, params_to_execute, &result); err != nil {
			fmt.Println(fmt.Errorf("Error: problem starting session; error == %s", err.Error()))
		} else {
			fmt.Println("result ==", result)
		}
		time.Sleep(1 * time.Second)
	}
}

func linuxChrome64DriverURL(latest bool) (chrome_driver_url string, driver_version string, err error) {
	chrome_drivers_url := "https://code.google.com/p/chromedriver/downloads/list"
	if resp, err := http.Get(chrome_drivers_url); err != nil {
		err = fmt.Errorf("Error: unable to get latest driver from %s", chrome_drivers_url)
		return chrome_driver_url, driver_version, err
	} else {
		defer resp.Body.Close()
		if bytes, err := ioutil.ReadAll(resp.Body); err != nil {
			err = fmt.Errorf("Error: unable to read bytes from body while getting driver from %s", chrome_drivers_url)
			return chrome_driver_url, driver_version, err
		} else {
			chrome_driver_url = string(bytes)
			if latest {
				chrome_driver_url = chrome_driver_url[strings.Index(chrome_driver_url, "'//chromedriver.googlecode.com/files/chromedriver_linux64_"):strings.LastIndex(chrome_driver_url, "supports Chrome")]
				driver_version = chrome_driver_url[strings.LastIndex(chrome_driver_url, "(")+1:strings.LastIndex(chrome_driver_url, ")")]
				chrome_driver_url = chrome_driver_url[strings.Index(chrome_driver_url, "//"):strings.Index(chrome_driver_url, "',")]
				chrome_driver_url = "https:" + chrome_driver_url
			} else {
				chrome_driver_url = chrome_driver_url[strings.LastIndex(chrome_driver_url, "'//chromedriver.googlecode.com/files/chromedriver_linux64_"):strings.LastIndex(chrome_driver_url, "deprecated")]
				driver_version = "v" + chrome_driver_url[strings.LastIndex(chrome_driver_url, "chromedriver_linux64_")+21:strings.LastIndex(chrome_driver_url, ".zip")]
				chrome_driver_url = chrome_driver_url[strings.Index(chrome_driver_url, "//"):strings.Index(chrome_driver_url, "',")]
				chrome_driver_url = "https:" + chrome_driver_url
			}
			return chrome_driver_url, driver_version, err
		}
	}
}

func userHomeDir() string {
	if usr, err := user.Current(); err != nil {
		if runtime.GOOS == "windows" {
			home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
			if home == "" {
				home = os.Getenv("USERPROFILE")
			}
			return home
		}
		return os.Getenv("HOME")
	} else {
		return usr.HomeDir
	}
}

func free_port() (int, error) {
	tcp_address := net.TCPAddr{net.ParseIP("127.0.0.1"), 0, ""}
	if tcp_listener, err := net.ListenTCP("tcp4", &tcp_address); err != nil {
		return 0, errors.New("Error: unable to listen on ephemeral port")
	} else {
		defer tcp_listener.Close()
		if port, err := strconv.Atoi(tcp_listener.Addr().String()[strings.LastIndex(tcp_listener.Addr().String(), ":")+1:]); err != nil {
			return 0, errors.New("Error: unable to retrieve ephemeral port")
		} else {
			return port, nil
		}
	}
}

func getChromeDriver(latest bool) (driver_path string, err error) {
	home_dir := userHomeDir()
	if chrome_driver_url, driver_version, err := linuxChrome64DriverURL(latest); err != nil {
		driver_dir := path.Join(home_dir, ".selenium", "drivers", "chrome", driver_version)
		driver_path := path.Join(driver_dir, "chromedriver")
		fmt.Println("driver_path5 ==", driver_path)
		return driver_path, err
	} else {
		driver_dir := path.Join(home_dir, ".selenium", "drivers", "chrome", driver_version)
		driver_path := path.Join(driver_dir, "chromedriver")
		fmt.Println("driver_path6 ==", driver_path)
		fmt.Println("chrome_driver_url ==", chrome_driver_url)
		fmt.Println("driver_version ==", driver_version)
		fmt.Println("driver_dir ==", driver_dir)
		if _, err := os.Stat(driver_dir); err != nil && os.IsNotExist(err){
			if file_info, err := os.Stat(home_dir); err != nil && os.IsNotExist(err) {
				fmt.Println(fmt.Errorf("Error: %s does not exist", home_dir))
				return driver_path, fmt.Errorf("Error: %s does not exist; err == %s", home_dir, err.Error())
			} else {
				if err := os.MkdirAll(driver_dir, file_info.Mode()); err != nil {
					return driver_path, fmt.Errorf("Error: unable to create %s; err == %s", driver_dir, err.Error())
				}
			}
		}
		if _, err := os.Stat(driver_path); err != nil && os.IsNotExist(err) {
			zip_name := chrome_driver_url[strings.LastIndex(chrome_driver_url, "/")+1:]
			if zip_file, err := ioutil.TempFile("", zip_name); err != nil {
				return driver_path, fmt.Errorf("Error: unable to create temporary file %s; err == %s", zip_name, err.Error())
			} else {
				defer zip_file.Close()
				zip_path := zip_file.Name()
				defer os.Remove(zip_path)
				if resp, err := http.Get(chrome_driver_url); err != nil {
					return driver_path, fmt.Errorf("Error: unable to get response from %s; err == %s", chrome_driver_url, err.Error())
				} else {
					defer resp.Body.Close()
					if _, err := io.Copy(zip_file, resp.Body); err != nil {
						return driver_path, fmt.Errorf("Error: unable to download %s; err == %s", chrome_driver_url, err.Error())
					} else {
						if zip_reader, err := zip.OpenReader(zip_path); err != nil {
							return driver_path, fmt.Errorf("Error: unable to open file %s; err == %s", zip_path, err.Error())
						} else {
							defer zip_reader.Close()
							for _, file := range zip_reader.File {
								if file_contents, err := file.Open(); err != nil {
									return driver_path, fmt.Errorf("Error: unable to open file %s within %s; err == %s", file.Name, zip_path, err.Error())
								} else {
									file_path := path.Join(driver_dir, file.Name)
									if chrome_driver, err := os.Create(file_path); err != nil {
										return driver_path, fmt.Errorf("Error: unable to create file %s; err == %s", file_path, err.Error())
									} else {
										if _, err := io.Copy(chrome_driver, file_contents); err != nil {
											return driver_path, fmt.Errorf("Error: unable to unzip %s into %s; err == %s", zip_path, file_path, err.Error())
										} else {
											chrome_driver.Close()
											if file_info, err := os.Stat(driver_path); err != nil {
												return driver_path, fmt.Errorf("Error: unable to stat %s; err == %s", driver_path, err.Error())
											} else {
												file_mode := file_info.Mode()
												fmt.Println("file_mode ==", file_mode)
												file_mode = file_mode | 0100
												fmt.Println("file_mode ==", file_mode)
												if err := os.Chmod(driver_path, file_mode); err != nil {
													return driver_path, fmt.Errorf("Error: unable to chmod %s; err == %s", driver_path, err.Error())
												} else {
													fmt.Println("Successfully downloaded and unzipped chromedriver")
													return driver_path, err
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
		return driver_path, err
	}
}

func startChromeDriver(latest bool, port int) error {
	// TODO: While making this work with multiple browsers, restructure to use existing driver while downloading new driver in goroutine for next use
	if driver_path, err := getChromeDriver(latest); err != nil {
		fmt.Println("driver_path1 ==", driver_path)
		return err
	} else {
		fmt.Println("driver_path2 ==", driver_path)
		fmt.Println("port ==", port)
		driver_command := exec.Command(driver_path, "--port="+strconv.Itoa(port))
		var stdout_buffer bytes.Buffer
		driver_command.Stdout = &stdout_buffer
		var stderr_buffer bytes.Buffer
		driver_command.Stderr = &stderr_buffer
		if err := driver_command.Run(); err != nil {
			fmt.Printf("latest driver_command.Stdout == %q\n", stdout_buffer.String())
			fmt.Printf("latest driver_command.Stderr == %q\n", stderr_buffer.String())
			return fmt.Errorf("Error: unable to run command, err == %s", err.Error())
		} else {
			fmt.Printf("latest driver_command.Stdout == %q\n", stdout_buffer.String())
			fmt.Printf("latest driver_command.Stderr == %q\n", stderr_buffer.String())
			return nil
		}
	}
}

func stopChromeDriver(port int) {
	shutdown_url := "http://127.0.0.1:" + strconv.Itoa(port) + "/shutdown"
	http.Get(shutdown_url)
}

func waitForChromeDriver(port int) {
	counter := 0
	for _, err := http.Get("http://127.0.0.1:" + strconv.Itoa(port)); err != nil && counter < 5000; _, err = http.Get("http://127.0.0.1:" + strconv.Itoa(port)) {
		counter++
		time.Sleep(time.Millisecond)
	}
}
