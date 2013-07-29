/*
Copyright (c) 2013, Aulus Egnatius Varialus <varialus@gmail.com>

Permission to use, copy, modify, and/or distribute this software for any purpose with or without fee is hereby granted, provided that the above copyright notice and this permission notice appear in all copies.

THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
*/

// http://code.google.com/p/selenium/source/browse/py/selenium/webdriver/chrome/service.py
// http://code.google.com/p/selenium/wiki/ChromeDriver

package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"os"
	"os/user"
	"runtime"
	"strings"
	"archive/zip"
)

var browser_name string

func main() {
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
	if latest_chrome_driver_url, err := LatestLinuxChrome64DriverURL(); err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println("latest_chrome_driver_url ==", latest_chrome_driver_url)
		driver_version := "v" + latest_chrome_driver_url[strings.LastIndex(latest_chrome_driver_url, "_")+1:strings.LastIndex(latest_chrome_driver_url, ".")]
		fmt.Println("driver_version ==", driver_version)
		home_dir := UserHomeDir()
		driver_dir := path.Join(home_dir, ".selenium", "drivers", "chrome", driver_version)
		fmt.Println("driver_dir ==", driver_dir)
		if _, err := os.Stat(driver_dir); err != nil && os.IsNotExist(err){
			if file_info, err := os.Stat(home_dir); err != nil && os.IsNotExist(err) {
				fmt.Println(fmt.Errorf("Error: %s does not exist", home_dir))
			} else {
				if err := os.MkdirAll(driver_dir, file_info.Mode()); err != nil {
					fmt.Println(fmt.Errorf("Error: unable to create %s", driver_dir))
				}
			}
		}
		driver_path := path.Join(driver_dir, "chromedriver")
		if _, err := os.Stat(driver_path); err != nil && os.IsNotExist(err){
			zip_name := latest_chrome_driver_url[strings.LastIndex(latest_chrome_driver_url, "/")+1:]
			if zip_file, err := ioutil.TempFile("", zip_name); err != nil {
				fmt.Println(fmt.Errorf("Error: unable to create temporary file %s", zip_name))
			} else {
				defer zip_file.Close()
				zip_path := zip_file.Name()
				defer os.Remove(zip_path)
				if resp, err := http.Get(latest_chrome_driver_url); err != nil {
					fmt.Println(fmt.Errorf("Error: unable to get response from %s", latest_chrome_driver_url))
				} else {
					defer resp.Body.Close()
					if _, err := io.Copy(zip_file, resp.Body); err != nil {
						fmt.Println(fmt.Errorf("Error: unable to download %s", latest_chrome_driver_url))
					} else {
						if zip_reader, err := zip.OpenReader(zip_path); err != nil {
							fmt.Println(fmt.Errorf("Error: unable to open file %s", zip_path))
						} else {
							defer zip_reader.Close()
							for _, file := range zip_reader.File {
								if file_contents, err := file.Open(); err != nil {
									fmt.Println(fmt.Errorf("Error: unable to open file %s within %s", file.Name, zip_path))
								} else {
									file_path := path.Join(driver_dir, file.Name)
									if chrome_driver, err := os.Create(file_path); err != nil {
										fmt.Println(fmt.Errorf("Error: unable to create file %s", file_path))
									} else {
										defer chrome_driver.Close()
										if _, err := io.Copy(chrome_driver, file_contents); err != nil {
											fmt.Println(fmt.Errorf("Error: unable to unzip %s into %s", zip_path, file_path))
										} else {
											fmt.Println("Successfully downloaded and unzipped chromedriver")
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

func LatestLinuxChrome64DriverURL() (latest_chrome_driver_url string, err error) {
	latest_chrome_drivers_url := "https://code.google.com/p/chromedriver/downloads/list"
	if resp, err := http.Get(latest_chrome_drivers_url); err != nil {
		err = fmt.Errorf("Error: unable to get latest %s drivers from %s", browser_name, latest_chrome_drivers_url)
		return latest_chrome_driver_url, err
	} else {
		defer resp.Body.Close()
		if bytes, err := ioutil.ReadAll(resp.Body); err != nil {
			err = fmt.Errorf("Error: unable to read bytes from body while getting %s drivers from %s", browser_name, latest_chrome_drivers_url)
			return latest_chrome_driver_url, err
		} else {
			latest_chrome_driver_url = string(bytes)
			latest_chrome_driver_url = latest_chrome_driver_url[strings.Index(latest_chrome_driver_url, "'//chromedriver.googlecode.com/files/chromedriver_linux64_"):strings.LastIndex(latest_chrome_driver_url, "supports Chrome")]
			latest_chrome_driver_url = latest_chrome_driver_url[strings.Index(latest_chrome_driver_url, "//"):strings.Index(latest_chrome_driver_url, "',")]
			latest_chrome_driver_url = "https:" + latest_chrome_driver_url
			return latest_chrome_driver_url, err
		}
	}
}

func UserHomeDir() string {
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
